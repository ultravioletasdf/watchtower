package main

import (
	"context"
	"database/sql"
	"errors"
	"time"
	"videoapp/proto"
	sqlc "videoapp/sql"
	"videoapp/utils"

	"github.com/minio/minio-go/v7"
	"github.com/rabbitmq/amqp091-go"
	protobuf "google.golang.org/protobuf/proto"
)

const (
	StageNotUploaded int64 = iota
	StageAwaitingProcessing
	StageProcessing
	StageProcessed
)
const (
	VisibilityPublic int64 = iota
	VisibilityUnlisted
	VisibilityPrivate
)

type videoService struct {
	proto.UnimplementedVideosServer
}

func (s *videoService) CreateUpload(ctx context.Context, req *proto.CreateUploadRequest) (*proto.CreateUploadResponse, error) {
	if len(req.Session) != SESSION_TOKEN_LENGTH {
		return nil, ErrSessionWrongSize
	}
	user, err := executor.GetUserFromSession(ctx, req.Session)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSessionNotFound
	}

	id := snowflakeNode.Generate()
	policy := minio.NewPostPolicy()
	policy.SetBucket("staging")
	policy.SetKey(id.String())
	policy.SetContentTypeStartsWith("video")
	policy.SetExpires(time.Now().Add(15 * time.Minute))

	policy.SetContentLengthRange(1024*1024, 1024*1024*1024*10)

	// url, err := minioClient.PresignedPostPolicy(ctx, "staging", id.String(), 15*time.Minute)
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

func (s *videoService) Create(ctx context.Context, req *proto.CreateVideoRequest) (*proto.CreateVideoResponse, error) {
	if len(req.Session) != SESSION_TOKEN_LENGTH {
		return nil, ErrSessionWrongSize
	}
	user, err := executor.GetUserFromSession(ctx, req.Session)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSessionNotFound
	}

	if !utils.Between(len(req.Title), 5, 100) {
		return nil, ErrTitleWrongSize
	}
	if len(req.Description) > 1000 {
		return nil, ErrDescriptionWrongSize
	}
	if !utils.Between(int(req.Visibility), int(VisibilityPublic), int(VisibilityPrivate)) {
		return nil, ErrInvalidVisibility
	}
	id := snowflakeNode.Generate()
	err = executor.CreateVideo(ctx, sqlc.CreateVideoParams{ID: id.Int64(), UploadID: req.UploadId, UserID: user.ID, Title: req.Title, Description: req.Description, Visibility: req.Visibility})
	if err != nil {
		return nil, ErrInternal(err)
	}
	err = executor.UpdateUploadStage(ctx, sqlc.UpdateUploadStageParams{Stage: StageAwaitingProcessing, ID: req.UploadId})
	if err != nil {
		return nil, ErrInternal(err)
	}
	message := proto.AnalyseVideoMessage{UploadId: req.UploadId, VideoId: id.Int64()}
	body, err := protobuf.Marshal(&message)
	if err != nil {
		return nil, ErrInternal(err)
	}
	err = rabbit.channel.PublishWithContext(ctx, "", rabbit.queues.AnalyseVideos.Name, true, false, amqp091.Publishing{ContentType: "application", DeliveryMode: amqp091.Persistent, Body: body})
	if err != nil {
		return nil, ErrInternal(err)
	}
	return &proto.CreateVideoResponse{Id: id.Int64()}, nil
}
