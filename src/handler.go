package main

import (
	"context"
	"io"
	"net/http"

	"github.com/oklog/ulid/v2"
	log "github.com/sirupsen/logrus"
	"github.com/superfly/fsm"
)

func NewStartHandler(manager *fsm.Manager) startHandler {
	fsmStarter, _, _ := fsm.Register[ReqMsg, Msg](manager, "myFsm").
		Start("PULLING", PullTransition).
		To("EXTRACTING", ExtractTransition).
		To("ACTIVATING", ActivateTransition).
		End("ACTIVATED").
		Build(context.TODO())
	return startHandler{starter: fsmStarter}
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

type startHandler struct {
	starter fsm.Start[ReqMsg, Msg]
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
