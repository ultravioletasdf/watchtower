package main

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"
	"videoapp/proto"
	sqlc "videoapp/sql"
	"videoapp/vips"

	"github.com/minio/minio-go/v7"
)

type thumbnailsServer struct {
	proto.UnimplementedThumbnailsServer
}

func (s *thumbnailsServer) CreateUpload(ctx context.Context, req *proto.CreateUploadRequest) (*proto.CreateUploadResponse, error) {
	if len(req.Session) != SESSION_TOKEN_LENGTH {
		return nil, ErrSessionWrongSize
	}
	user, err := executor.GetUserFromSession(ctx, req.Session)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSessionNotFound
	}

	id := snowflakeNode.Generate()
	policy := minio.NewPostPolicy()
	policy.SetBucket("staging-thumbnails")
	policy.SetKey(id.String())
	policy.SetContentTypeStartsWith("image")
	policy.SetExpires(time.Now().Add(15 * time.Minute))

	policy.SetContentLengthRange(1024*1024, 1024*1024*2)

	url, fd, err := minioClient.PresignedPostPolicy(ctx, policy)
	if err != nil {
		return nil, ErrInternal(err)
	}

	err = executor.CreateUpload(ctx, sqlc.CreateUploadParams{ID: id.Int64(), UserID: user.ID, Stage: StageNotUploaded, CreatedAt: time.Now().Unix()})
	if err != nil {
		return nil, ErrInternal(err)
	}

	return &proto.CreateUploadResponse{Url: url.String(), Id: id.Int64(), FormData: fd}, nil
}
func (s *thumbnailsServer) Process(ctx context.Context, req *proto.ThumbnailsProcessRequest) (*proto.ThumbnailsProcessResponse, error) {
	if len(req.Session) != SESSION_TOKEN_LENGTH {
		return nil, ErrSessionWrongSize
	}
	user, err := executor.GetUserFromSession(ctx, req.Session)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSessionNotFound
	}
	thumb, err := executor.GetUpload(ctx, req.Id)
	if err == sql.ErrNoRows {
		return nil, ErrNoUploadFound
	}
	if err != nil {
		return nil, ErrInternal(err)
	}
	if thumb.UserID != user.ID {
		return nil, ErrUnauthorized
	}

	idString := strconv.FormatInt(req.Id, 10)
	obj, err := minioClient.GetObject(ctx, "staging-thumbnails", idString, minio.GetObjectOptions{})
	if err != nil {
		return nil, ErrInternal(err)
	}
	source := vips.NewSource(obj)
	defer source.Close()

	image, err := vips.NewThumbnailSource(source, 1280, &vips.ThumbnailSourceOptions{Height: 720, Size: vips.SizeBoth, FailOn: vips.FailOnError})
	if err != nil {
		return nil, ErrInternal(err)
	}
	defer image.Close()
	err = image.Webpsave(idString+".webp", &vips.WebpsaveOptions{Q: 85, Effort: 6, SmartSubsample: true, Lossless: true})
	if err != nil {
		return nil, ErrInternal(err)
	}
	return &proto.ThumbnailsProcessResponse{}, nil
}
