package main

import "math"
import "strings"
import "os/exec"
import "bufio"
import "strconv"

const SECTOR_SIZE_BYTES int64 = 512
const BLOCKSIZE_BYTES int64 = 4096
const FS_ESTIMATE_BUFFER float64 = 1.2

func estimateTarSizeBytes(tarpath string) (int64, error) {
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
