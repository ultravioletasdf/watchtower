package main

import (
	"context"
	"io/fs"
	"log"
	"path/filepath"

	"github.com/minio/minio-go/v7"
)

var ctx = context.Background()

func UploadDir(dir string) error {
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		rel, err := filepath.Rel("results", path)
		if err != nil {
			return err
		}
		key := filepath.ToSlash(rel)

		var contentType string
		switch filepath.Ext(path) {
		case "m3u8":
			contentType = "application/vnd.apple.mpegurl"
		case "ts":
			contentType = "video/MP2T"
		}
		_, err = s3.FPutObject(ctx, "videos", key, path, minio.PutObjectOptions{ContentType: contentType})
		if err != nil {
			log.Printf("failed to upload %s: %v\n", path, err)
		}
		return nil
	})
}
