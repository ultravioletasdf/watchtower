package main

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/minio/minio-go/v7"

	"videoapp/proto"
	"videoapp/server/common"
	sqlc "videoapp/sql"
	"videoapp/vips"
)

type thumbnailsServer struct {
	proto.UnimplementedThumbnailsServer
}

func (s *thumbnailsServer) CreateUpload(ctx context.Context, req *proto.CreateUploadRequest) (*proto.CreateUploadResponse, error) {
	if len(req.Session) != SESSION_TOKEN_LENGTH {
		return nil, common.ErrSessionWrongSize
	}
	user, err := executor.GetUserFromSession(ctx, req.Session)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, common.ErrSessionNotFound
	}

	id := snowflakeNode.Generate()

	policy := minio.NewPostPolicy()
	policy.SetBucket("staging-thumbnails")
	policy.SetKey(id.String())
	policy.SetContentTypeStartsWith("image")
	policy.SetExpires(time.Now().Add(15 * time.Minute))

	policy.SetContentLengthRange(1024, 1024*1024*2)

	url, fd, err := s3.PresignedPostPolicy(ctx, policy)
	if err != nil {
		return nil, common.ErrInternal(err)
	}
	err = executor.CreateUpload(ctx, sqlc.CreateUploadParams{ID: id.Int64(), UserID: user.ID})
	if err != nil {
		return nil, common.ErrInternal(err)
	}

	return &proto.CreateUploadResponse{Url: url.String(), Id: id.Int64(), FormData: fd}, nil
}
func (s *thumbnailsServer) Process(ctx context.Context, req *proto.ThumbnailsProcessRequest) (*proto.ThumbnailsProcessResponse, error) {
	if len(req.Session) != SESSION_TOKEN_LENGTH {
		return nil, common.ErrSessionWrongSize
	}
	user, err := executor.GetUserFromSession(ctx, req.Session)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, common.ErrSessionNotFound
	}
	thumb, err := executor.GetUpload(ctx, req.Id)
	if err == sql.ErrNoRows {
		return nil, common.ErrNoUploadFound
	}
	if err != nil {
		return nil, common.ErrInternal(err)
	}
	if thumb.UserID != user.ID {
		return nil, common.ErrUnauthorized
	}

	idString := strconv.FormatInt(req.Id, 10)
	obj, err := s3.GetObject(ctx, "staging-thumbnails", idString, minio.GetObjectOptions{})
	if err != nil {
		return nil, common.ErrInternal(err)
	}
	source := vips.NewSource(obj)
	defer source.Close()

	image, err := vips.NewThumbnailSource(source, 1280, &vips.ThumbnailSourceOptions{Height: 720, Size: vips.SizeBoth, FailOn: vips.FailOnError, Crop: vips.InterestingAttention})
	if err != nil {
		return nil, common.ErrInternal(err)
	}
	defer image.Close()

	buf := &writeCloser{bytes.NewBuffer(nil)}
	target := vips.NewTarget(buf)
	err = image.WebpsaveTarget(target, vips.DefaultWebpsaveTargetOptions())
	if err != nil {
		return nil, common.ErrInternal(err)
	}
	_, err = s3.PutObject(ctx, "thumbnails", idString+".webp", buf, int64(buf.Len()), minio.PutObjectOptions{})
	if err != nil {
		return nil, common.ErrInternal(err)
	}

	return &proto.ThumbnailsProcessResponse{}, nil
}

type writeCloser struct {
	*bytes.Buffer
}

func (wc *writeCloser) Close() error {
	return nil
}
