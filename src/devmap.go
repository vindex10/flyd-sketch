package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

const DEV_PREFIX = "dev"
const MNT_DIR = "mnt"

func registerOrdinal(id int) error {
	cmd := exec.Command("dmsetup", "message", CFG.ThinPoolDevice, "0", "create_thin "+strconv.Itoa(id))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

//func makeSnapshot() (int, error) {
//return 0, nil
//}

func extractImageToDevice(localPath string, volumeId int) error {
	entries, err := os.ReadDir(localPath)
	if err != nil {
		return err
	}
	var totSizeBytes int64
	for _, e := range entries {
		fname := filepath.Join(localPath, e.Name())
		size, err := estimateTarSizeBytes(fname)
		if err != nil {
			return err
		}
		totSizeBytes += size
	}
	// 5% of blocks are reserved. add some buffer
	totSectors := int64(float64(totSizeBytes/SECTOR_SIZE_BYTES) * FS_ESTIMATE_BUFFER)
	devName := DEV_PREFIX + strconv.Itoa(volumeId)
	cmd := exec.Command("dmsetup", "create", devName, "--table", "0 "+strconv.FormatInt(totSectors, 10)+" thin "+CFG.ThinPoolDevice+" "+strconv.Itoa(volumeId))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}
	cmd = exec.Command("mkfs.ext4", "/dev/mapper/"+devName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}
	tmpMountDir := filepath.Join(CFG.StateDir, MNT_DIR, devName)
	os.MkdirAll(tmpMountDir, 0700)
	defer os.Remove(tmpMountDir)
	cmd = exec.Command("mount", "/dev/mapper/"+devName, tmpMountDir)
	defer func() {
		cmd = exec.Command("umount", tmpMountDir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	}()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}
	for _, e := range entries {
		cmd := exec.Command("tar", "xf", filepath.Join(localPath, e.Name()), "-C", tmpMountDir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return err
		}
	}
	return nil
}

func createSnapshot(volumeId int, snapshotId int) error {
	devName := DEV_PREFIX + strconv.Itoa(volumeId)
	cmd := exec.Command("dmsetup", "suspend", "/dev/mapper/"+devName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	cmd = exec.Command("dmsetup", "message", CFG.ThinPoolDevice, "0", "create_snap "+strconv.Itoa(snapshotId)+" "+strconv.Itoa(volumeId))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}
	cmd = exec.Command("dmsetup", "resume", "/dev/mapper/"+devName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
