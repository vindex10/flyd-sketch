package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	log "github.com/sirupsen/logrus"
)

const CACHE_DIR = "imagecache"

func listObjectsForImage(imageId string) ([]string, error) {
	if imageId == "" {
		return []string{}, errors.New("imageId can't be empty")
	}
	// Get the first page of results for ListObjectsV2 for a bucket
	output, err := s3Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(CFG.imageBucket),
		Prefix: aws.String(CFG.imagePrefix + "/" + imageId),
	})
	if err != nil {
		return []string{}, err
	}

	// if there are too many objects - process pages.
	res := []string{}
	for _, object := range output.Contents {
		log.
			WithField("key", aws.ToString(object.Key)).
			WithField("size", *object.Size).
			Info("object")
		res = append(res, aws.ToString(object.Key))
	}
	return res, nil
}

func imageLocalPath(imageId string) string {
	return filepath.Join(CFG.stateDir, CACHE_DIR, imageId)
}

func imageRemotePath(imageId string) string {
	remotePath := "s3://" + CFG.imageBucket + "/" + CFG.imagePrefix + "/" + imageId
	return remotePath
}

func imageIsCached(imageId string) (string, bool) {
	cachedPath := imageLocalPath(imageId)
	if _, err := os.Stat(cachedPath); os.IsNotExist(err) {
		return cachedPath, false
	}
	return cachedPath, true
}

func ensureLocalImage(imageId string) error {
	if _, isCached := imageIsCached(imageId); isCached {
		log.WithField("image", imageId).Info("image exists locally. using cached")
		return nil
	}
	return downloadImage(imageId)
}

func downloadImage(imageId string) error {
	remotePath := imageRemotePath(imageId)
	cachedPath := imageLocalPath(imageId)
	objs, err := listObjectsForImage(imageId)
	if err != nil {
		return err
	} else if len(objs) == 0 {
		return errors.New("image not found in remote")
	}
	cmd := exec.Command("aws", "s3", "cp", "--no-sign-request", "--recursive", remotePath, cachedPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		return err
	}
	return nil
}

func main() {
	initConfig()
	initS3Client()
	_, err := listObjectsForImage("golang")
	if err != nil {
		log.Info(err)
	}
	err = ensureLocalImage("python")
	log.Info(err)
}
