package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/oklog/ulid/v2"
	"github.com/superfly/fsm"

	log "github.com/sirupsen/logrus"
)

func main() {
	db := initStateDb()
	defer db.Close()

	cfg := fsm.Config{
		Logger: log.WithFields(log.Fields{}),
		DBPath: CFG.stateDir,
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
	socket := filepath.Join(CFG.stateDir, "flyd-sketch.sock")
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
	var fsmReq *MsgReq = fsm.NewRequest(
		&ReqMsg{imageId: "parsedBody"},
		newMsg("respmsg"))
	runId := ulid.Make().String()
	s.starter(ctx, runId, fsmReq)
}

func PullTransition(cnt context.Context, req *MsgReq) (*MsgResp, error) {
	resp := fsm.NewResponse(newMsg("pulled"))
	imageId := req.Msg.imageId
	_, exists := imageSnapshotId(imageId)
	if exists {
		log.WithField("image_id", imageId).Info("We pulled this image before. Skipping this step")
		return resp, nil
	}
	log.WithField("image_id", imageId).Info("Pulling image")
	// pull
	return resp, nil
}

func ExtractTransition(cnt context.Context, req *MsgReq) (*MsgResp, error) {
	resp := fsm.NewResponse(newMsg("extracted"))
	return resp, nil
}

func ActivateTransition(cnt context.Context, req *MsgReq) (*MsgResp, error) {
	resp := fsm.NewResponse(newMsg("activated"))
	return resp, nil
}
