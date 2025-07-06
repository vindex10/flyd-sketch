package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/oklog/ulid/v2"
	"github.com/superfly/fsm"

	log "github.com/sirupsen/logrus"
)

func main() {
	initConfig()
	initStateDb()
	initS3Client()

	db := initStateDb()
	defer db.Close()

	cfg := fsm.Config{
		Logger: log.WithFields(log.Fields{}),
		DBPath: CFG.StateDir,
		Queues: map[string]int{"main": 5},
	}
	manager, err := fsm.New(cfg)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fsmStarter, _, _ := fsm.Register[ReqMsg, Msg](manager, "myFsm").
		Start("PULLING", PullTransition).
		To("EXTRACTING", ExtractTransition).
		To("ACTIVATING", ActivateTransition).
		End("ACTIVATED").
		Build(context.TODO())

	mux := http.NewServeMux()
	mux.Handle("start/", startHandler{starter: fsmStarter})
	socket := filepath.Join(CFG.StateDir, "flyd-sketch.sock")
	os.Remove(socket)
	unixListener, _ := net.Listen("unix", socket)
	defer os.Remove(socket)
	http.Serve(unixListener, mux)
	return
}

type ReqMsg struct {
	imageId string
}

type Msg struct {
	msg string
}

type MsgReq = fsm.Request[ReqMsg, Msg]
type MsgResp = fsm.Response[Msg]

func newMsg(s string) *Msg {
	return &Msg{msg: s}
}

type startHandler struct {
	starter fsm.Start[ReqMsg, Msg]
}

func (s startHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		log.Error("Couldn't parse query")
		return
	}
	imageId := string(bodyBytes)

	var fsmReq *MsgReq = fsm.NewRequest(
		&ReqMsg{imageId: imageId},
		newMsg("respmsg"))
	runId := ulid.Make().String()
	s.starter(ctx, runId, fsmReq)
}

func PullTransition(cnt context.Context, req *MsgReq) (*MsgResp, error) {
	resp := fsm.NewResponse(newMsg("pulled"))
	imageId := req.Msg.imageId
	_, exists, err := getVolumeId(req.Run().ID)
	if err != nil {
		return fsm.NewResponse(newMsg("failed")), fsm.Abort(err)
	}
	if exists {
		log.WithField("image_id", imageId).Info("We pulled this image before. Skipping this step")
		return resp, nil
	}
	log.WithField("image_id", imageId).Info("Pulling image")
	err = ensureLocalImage(imageId)
	if err != nil {
		return fsm.NewResponse(newMsg("failed")), fsm.Abort(err)
	}
	return resp, nil
}

func ExtractTransition(cnt context.Context, req *MsgReq) (*MsgResp, error) {
	resp := fsm.NewResponse(newMsg("extracted"))
	imageId := req.Msg.imageId
	_, exists, err := getVolumeId(imageId)
	if err != nil {
		return fsm.NewResponse(newMsg("failed")), fsm.Abort(err)
	}
	if exists {
		log.WithField("image_id", imageId).Info("Snapshot for this image already exists. Skip extraction.")
		return resp, nil
	}
	errc := make(chan error)
	defer func() {
		defer close(errc)
		select {
		case <-errc:
			deleteVolumeRecord(req.Run().ID)
		default:
		}
	}()
	failedResp := fsm.NewResponse(newMsg("failed"))

	volumeId, err := generateVolumeId(imageId)
	if err != nil {
		return failedResp, fsm.Abort(err)
	}
	err = registerOrdinal(volumeId)
	if err != nil {
		return failedResp, fsm.Abort(err)
	}
	localImagePath := imageLocalPath(imageId)
	err = extractImageToDevice(localImagePath, volumeId)
	if err != nil {
		return failedResp, fsm.Abort(err)
	}
	return resp, nil
}

func ActivateTransition(cnt context.Context, req *MsgReq) (*MsgResp, error) {
	resp := fsm.NewResponse(newMsg("activated"))
	imageId := req.Msg.imageId
	_, exists, err := getSnapshotId(imageId)
	if err != nil {
		return fsm.NewResponse(newMsg("failed")), fsm.Abort(err)
	}
	if exists {
		log.WithField("image_id", imageId).Info("Snapshot already exists. Assuming the image is activated.")
		return resp, nil
	}
	volumeId, exists, err := getVolumeId(imageId)
	if err != nil || !exists {
		return fsm.NewResponse(newMsg("failed")), fsm.Abort(err)
	}
	snapshotId, err := generateSnapshotId(req.Run().ID, imageId)
	if err != nil {
		return fsm.NewResponse(newMsg("failed")), fsm.Abort(err)
	}
	err = createSnapshot(volumeId, snapshotId)
	if err != nil {
		return fsm.NewResponse(newMsg("failed")), fsm.Abort(err)
	}
	return resp, nil
}
