package utils

import (
	"context"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func ConnectS3(cfg Config) *minio.Client {
	s3, err := minio.New(cfg.Minio.Endpoint, &minio.Options{Creds: credentials.NewStaticV4(cfg.Minio.AccessKey, cfg.Minio.SecretKey, ""), Secure: false})
	if err != nil {
		log.Fatalf("failed to connect to minio: %v", err)
	}

	buckets := []string{"staging", "videos", "thumbnails", "staging-thumbnails", "avatars"}

	for _, bucket := range buckets {
		exists, err := s3.BucketExists(context.Background(), bucket)
		if err != nil {
			log.Fatalf("failed to verify bucket exists: %v", err)
		}
		if !exists {
			if err := s3.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{}); err != nil {
				log.Fatalf("failed to create bucket: %v", err)
			}
			log.Printf("Created bucket %s\n", bucket)
		}
	}
	return s3
}
