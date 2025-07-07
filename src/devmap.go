package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

const DEV_PREFIX = "dev"
const MNT_DIR = "mnt"

func registerOrdinal(id int) error {
	err := runCmdVerbose(
		"dmsetup",
		"message",
		CFG.ThinPoolDevice,
		"0",
		fmt.Sprintf("create_thin %d", id),
	)
	if err != nil {
		return err
	}
	return nil
}

func extractImageToDevice(imageId string, volumeId int) error {
	devName := getDeviceName(volumeId)
	localPath := imageLocalPath(imageId)
	totSizeBytes, err := imageEstimateUnpackedSizeBytes(imageId)
	if err != nil {
		return err
	}
	totSectors := estimateNewVolumeSizeSectors(totSizeBytes)
	err = runCmdVerbose(
		"dmsetup",
		"create",
		devName,
		"--table",
		fmt.Sprintf("0 %d thin %s %d", totSectors, CFG.ThinPoolDevice, volumeId),
	)
	if err != nil {
		return err
	}
	devPath := "/dev/mapper/" + devName
	err = runCmdVerbose("mkfs.ext4", devPath)
	if err != nil {
		return err
	}
	tmpMountDir := filepath.Join(CFG.StateDir, MNT_DIR, devName)
	os.MkdirAll(tmpMountDir, 0700)
	defer os.Remove(tmpMountDir)
	err = runCmdVerbose("mount", devPath, tmpMountDir)
	if err != nil {
		return err
	}
	defer func() {
		runCmdVerbose("umount", tmpMountDir)
	}()
	entries, err := os.ReadDir(localPath)
	if err != nil {
		return err
	}
	for _, e := range entries {
		err := runCmdVerbose("tar", "xf", filepath.Join(localPath, e.Name()), "-C", tmpMountDir)
		if err != nil {
			return err
		}
	}
	return nil
}

func suspendVolume(volumeId int) error {
	devPath := "/dev/mapper/" + getDeviceName(volumeId)
	err := runCmdVerbose("dmsetup", "suspend", devPath)
	if err != nil {
		return err
	}
	return nil
}

func createSnapshot(volumeId int, snapshotId int) error {
	err := runCmdVerbose(
		"dmsetup",
		"message",
		CFG.ThinPoolDevice,
		"0",
		fmt.Sprintf("create_snap %d %d", snapshotId, volumeId),
	)
	if err != nil {
		return err
	}
	return nil
}

func getDeviceName(volumeId int) string {
	return DEV_PREFIX + strconv.Itoa(volumeId)
}

func estimateNewVolumeSizeSectors(totBytes int64) int64 {
	// 5% of blocks are reserved by fs. add some buffer
	return int64(float64(totBytes/SECTOR_SIZE_BYTES) * FS_ESTIMATE_BUFFER)
}
