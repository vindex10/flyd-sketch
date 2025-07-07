package main

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/superfly/fsm"
)

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
	failedResp := fsm.NewResponse(newMsg("failed"))

	imageId := req.Msg.imageId
	_, exists, err := getVolumeId(imageId)
	if err != nil {
		return failedResp, fsm.Abort(err)
	}
	if exists {
		log.WithField("image_id", imageId).Info("Snapshot for this image already exists. Skip extraction.")
		return resp, nil
	}

	volumeId, err := generateVolumeId(imageId)
	if err != nil {
		deleteVolumeRecord(req.Run().ID)
		return failedResp, fsm.Abort(err)
	}
	err = registerOrdinal(volumeId)
	if err != nil {
		return failedResp, fsm.Abort(err)
	}
	err = extractImageToDevice(imageId, volumeId)
	if err != nil {
		deleteVolumeRecord(req.Run().ID)
		return failedResp, fsm.Abort(err)
	}
	err = suspendVolume(volumeId)
	if err != nil {
		deleteVolumeRecord(req.Run().ID)
		return failedResp, fsm.Abort(err)
	}
	return resp, nil
}

func ActivateTransition(cnt context.Context, req *MsgReq) (*MsgResp, error) {
	resp := fsm.NewResponse(newMsg("activated"))
	failedResp := fsm.NewResponse(newMsg("failed"))

	imageId := req.Msg.imageId
	_, exists, err := getSnapshotId(imageId)
	if err != nil {
		return failedResp, fsm.Abort(err)
	}
	if exists {
		log.WithField("image_id", imageId).Info("Snapshot already exists. Assuming the image is activated.")
		return resp, nil
	}
	volumeId, exists, err := getVolumeId(imageId)
	if err != nil || !exists {
		return failedResp, fsm.Abort(err)
	}
	snapshotId, err := generateSnapshotId(req.Run().ID, imageId)
	if err != nil {
		return failedResp, fsm.Abort(err)
	}
	err = createSnapshot(volumeId, snapshotId)
	if err != nil {
		return failedResp, fsm.Abort(err)
	}
	return resp, nil
}
