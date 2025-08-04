package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func getBitrate(file string) (int, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=bit_rate", "-of", "default=noprint_wrappers=1:nokey=1", file)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(out.String()))
}
func GetBitrate(basePath string, res Resolution) (max int, avg int, err error) {
	files, err := filepath.Glob(fmt.Sprintf("%s/%dx%d_*.ts", basePath, res.Width, res.Height))
	if err != nil {
		return 0, 0, err
	}

	var total int
	for _, file := range files {
		bitrate, err := getBitrate(file)
		if err != nil {
			return 0, 0, err
		}
		if bitrate > max {
			max = bitrate
		}
		total += bitrate
	}
	return max, total / len(files), nil
}
