package main

import "math"
import "os"
import "strings"
import "os/exec"
import "bufio"
import "strconv"

const SECTOR_SIZE_BYTES int64 = 512
const BLOCKSIZE_BYTES int64 = 4096
const FS_ESTIMATE_BUFFER float64 = 1.2

func estimateTarSizeBytes(tarpath string) (int64, error) {
	// ext4 counts folders as 4k, tar does not
	// use tar -tvf and round up object sizes to minimal block size
	blocksizeF64 := float64(BLOCKSIZE_BYTES)
	output, err := exec.Command("tar", "tvf", tarpath).Output()
	if err != nil {
		return -1, err
	}
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	var tot int64 = 0
	for scanner.Scan() {
		line := scanner.Text()
		cols := strings.Fields(line)
		size, err := strconv.ParseInt(cols[2], 10, 64)
		if err != nil {
			return -1, nil
		}
		estimate := int64(math.Ceil(float64(size)/blocksizeF64)) * BLOCKSIZE_BYTES
		estimate = max(estimate, BLOCKSIZE_BYTES)
		tot += estimate
	}
	return tot, nil
}

func runCmdVerbose(args ...string) error {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}
