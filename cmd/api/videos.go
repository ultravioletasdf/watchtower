package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/minio/minio-go/v7"
	"github.com/rabbitmq/amqp091-go"
	protobuf "google.golang.org/protobuf/proto"

	common "videoapp/internal/errors"
	"videoapp/internal/generated/proto"
	sqlc "videoapp/internal/generated/sqlc"
	"videoapp/internal/utils"
)

type videoServer struct {
	proto.UnimplementedVideosServer
}

func (s *videoServer) CreateUpload(ctx context.Context, req *proto.CreateUploadRequest) (*proto.CreateUploadResponse, error) {
	if len(req.Session) != SESSION_TOKEN_LENGTH {
		return nil, common.ErrSessionWrongSize
	}
	user, err := executor.GetUserFromSession(ctx, req.Session)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, common.ErrSessionNotFound
	}

	id := snowflakeNode.Generate()
	policy := minio.NewPostPolicy()
	policy.SetBucket("staging")
	policy.SetKey(id.String())
	policy.SetContentTypeStartsWith("video")
	policy.SetExpires(time.Now().Add(15 * time.Minute))

	policy.SetContentLengthRange(1024*1024, 1024*1024*1024*10)

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

func (s *videoServer) Create(ctx context.Context, req *proto.VideosCreateRequest) (*proto.VideosCreateResponse, error) {
	if len(req.Session) != SESSION_TOKEN_LENGTH {
		return nil, common.ErrSessionWrongSize
	}
	user, err := executor.GetUserFromSession(ctx, req.Session)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, common.ErrSessionNotFound
	}

	if !utils.Between(len(req.Title), 5, 100) {
		return nil, common.ErrTitleWrongSize
	}
	if len(req.Description) > 1000 {
		return nil, common.ErrDescriptionWrongSize
	}
	if !utils.Between(int(req.Visibility), int(proto.Visibility_Public), int(proto.Visibility_Private)) {
		return nil, common.ErrInvalidVisibility
	}
	_, err = s3.StatObject(ctx, "staging", strconv.FormatInt(req.UploadId, 10), minio.GetObjectOptions{})
	if err != nil {
		return nil, common.ErrNoUploadFound
	}
	id := snowflakeNode.Generate()
	err = executor.CreateVideo(ctx, sqlc.CreateVideoParams{ID: id.Int64(), UploadID: req.UploadId, UserID: user.ID, Title: req.Title, Description: req.Description, Visibility: int32(req.Visibility), ThumbnailID: req.ThumbnailId, Stage: int32(proto.Stage_AwaitingProcessing)})
	if err != nil {
		return nil, common.ErrInternal(err)
	}

	message := proto.AnalyseVideoMessage{UploadId: req.UploadId, VideoId: id.Int64()}
	body, err := protobuf.Marshal(&message)
	if err != nil {
		return nil, common.ErrInternal(err)
	}
	err = rabbit.channel.PublishWithContext(ctx, "", rabbit.queues.AnalyseVideos.Name, true, false, amqp091.Publishing{ContentType: "application", DeliveryMode: amqp091.Persistent, Body: body})
	if err != nil {
		return nil, common.ErrInternal(err)
	}
	err = rabbit.channel.PublishWithContext(ctx, "", rabbit.queues.TranscodeVideos.Name, true, false, amqp091.Publishing{ContentType: "application", DeliveryMode: amqp091.Persistent, Body: body})
	if err != nil {
		return nil, common.ErrInternal(err)
	}
	err = rabbit.channel.PublishWithContext(ctx, "", rabbit.queues.GenerateThumbnails.Name, true, false, amqp091.Publishing{ContentType: "application", DeliveryMode: amqp091.Persistent, Body: body})
	if err != nil {
		return nil, common.ErrInternal(err)
	}
	return &proto.VideosCreateResponse{Id: id.Int64()}, nil
}
func (s *videoServer) GetUserVideos(ctx context.Context, req *proto.GetUserVideosRequest) (*proto.GetUserVideosResponse, error) {
	var showPrivate bool
	if req.Session != "" {
		_, err := executor.GetUserFromSession(ctx, req.Session)
		if err == nil {
			showPrivate = true
		} else if !errors.Is(err, sql.ErrNoRows) {
			fmt.Println(err)
		}
	}

	res, err := executor.GetUserVideos(ctx, sqlc.GetUserVideosParams{UserID: req.Id, ShowPrivate: showPrivate})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, common.ErrNoVideosFound
	}
	if err != nil {
		return nil, common.ErrInternal(err)
	}
	videos := make([]*proto.GetUserVideosResponseVideo, len(res))
	for i := range res {
		videos[i] = &proto.GetUserVideosResponseVideo{Id: res[i].ID, Title: res[i].Title, Visibility: proto.Visibility(res[i].Visibility), CreatedAt: res[i].CreatedAt.Time.Unix(), ThumbnailId: res[i].ThumbnailID, Stage: proto.Stage(res[i].Stage)}
	}
	return &proto.GetUserVideosResponse{Videos: videos}, nil
}

func (s *videoServer) Get(ctx context.Context, req *proto.GetVideoRequest) (*proto.GetVideoResponse, error) {
	v, err := executor.GetVideo(ctx, req.Id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, common.ErrVideoNotFound
	}
	if err != nil {
		return nil, common.ErrInternal(err)
	}

	if v.Visibility == int32(proto.Visibility_Private) {
		if len(req.Session) != SESSION_TOKEN_LENGTH {
			return nil, common.ErrSessionWrongSize
		}
		user, err := executor.GetUserFromSession(ctx, req.Session)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.ErrSessionNotFound
		}
		if v.UserID != user.ID {
			return nil, common.ErrUnauthorized
		}
	}

	var userReaction int32
	if req.Session != "" {
		reaction, err := executor.GetReaction(ctx, sqlc.GetReactionParams{TargetID: req.Id, Token: req.Session})
		if !errors.Is(err, sql.ErrNoRows) && err != nil {
			return nil, common.ErrInternal(err)
		}
		userReaction = reaction
	}

	payload, sig, err := utils.GeneratePresignedUrl(privateKey, v.UploadID, req.Session)
	if err != nil {
		return nil, err
	}

	return &proto.GetVideoResponse{Id: v.ID, Title: v.Title, Visibility: proto.Visibility(v.Visibility), CreatedAt: v.CreatedAt.Time.Unix(), ThumbnailId: v.ThumbnailID, UploadId: v.UploadID, UserId: v.UserID, Stage: proto.Stage(v.Stage), AuthorizationPayload: payload, AuthorizationSignature: sig, Likes: v.Likes, Dislikes: v.Dislikes, UserReaction: userReaction, Comments: v.Comments}, nil
}
func (s *videoServer) Delete(ctx context.Context, req *proto.DeleteVideoRequest) (*proto.DeleteVideoResponse, error) {
	if len(req.Session) != SESSION_TOKEN_LENGTH {

		return nil, common.ErrSessionWrongSize
	}
	user, err := executor.GetUserFromSession(ctx, req.Session)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, common.ErrSessionNotFound
	}
	v, err := executor.GetVideo(ctx, req.VideoId)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, common.ErrVideoNotFound
	}
	if err != nil {
		return nil, common.ErrInternal(err)
	}
	if v.UserID != user.ID {
		return nil, common.ErrUnauthorized
	}

	objectCh := s3.ListObjects(ctx, "videos", minio.ListObjectsOptions{Prefix: strconv.FormatInt(v.UploadID, 10), Recursive: false})
	for err := range s3.RemoveObjects(ctx, "videos", objectCh, minio.RemoveObjectsOptions{}) {
		log.Printf("failed to delete %s: %v\n", err.ObjectName, err.Err)
	}
	if err = executor.DeleteVideo(ctx, v.ID); err != nil {
		return nil, common.ErrInternal(err)
	}
	return &proto.DeleteVideoResponse{}, nil
}
func (s *videoServer) GetStage(ctx context.Context, req *proto.VideosGetStageRequest) (*proto.VideosGetStageResponse, error) {
	stage, err := executor.GetStage(ctx, req.Id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, common.ErrVideoNotFound
	}
	if err != nil {
		return nil, common.ErrInternal(err)
	}
	var pl, sg string
	if stage.Stage == int32(proto.Stage_Processed) {
		payload, sig, err := utils.GeneratePresignedUrl(privateKey, stage.UploadID, req.Session)
		if err != nil {
			return nil, err
		}
		pl, sg = payload, sig
	}
	return &proto.VideosGetStageResponse{Stage: proto.Stage(stage.Stage), UploadId: stage.UploadID, AuthorizationPayload: pl, AuthorizationSignature: sg}, nil
}
func (s *videoServer) GetStages(ctx context.Context, req *proto.VideosGetStagesRequest) (*proto.VideosGetStagesResponse, error) {
	json, err := executor.GetStages(ctx, req.Ids)
	if err != nil {
		return nil, common.ErrInternal(err)
	}
	return &proto.VideosGetStagesResponse{Result: string(json)}, nil
}
func (s *videoServer) CreateComment(ctx context.Context, req *proto.CreateCommentRequest) (*proto.Comment, error) {
	if !utils.Between(len(req.Content), 3, 1000) {
		return nil, common.ErrInvalidComment
	}

	user, err := executor.GetUserFromSession(ctx, req.Session)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, common.ErrUnauthorized
	} else if err != nil {
		return nil, common.ErrInternal(err)
	}

	id := snowflakeNode.Generate().Int64()

	err = executor.CreateComment(ctx, sqlc.CreateCommentParams{ID: id, VideoID: req.VideoId, ReferenceID: pgtype.Int8{Int64: req.ReferenceId, Valid: req.ReferenceId != 0}, Content: req.Content, UserID: user.ID})
	return &proto.Comment{Id: id, UserId: user.ID, VideoId: req.VideoId, ReferenceId: req.ReferenceId, Content: req.Content, Username: user.Username}, common.ErrInternal(err)
}
func (s *videoServer) ListComments(ctx context.Context, req *proto.ListCommentsRequest) (*proto.ListCommentsResponse, error) {
	var userId int64
	// Get the users id if a session is provided, so we can get the users reaction
	if req.Session != "" {
		u, err := executor.GetUserFromSession(ctx, req.Session)
		if err == nil {
			userId = u.ID
		} else if !errors.Is(err, sql.ErrNoRows) { // Continue even if we couldn't get the users ID
			fmt.Printf("Couldn't get users session: %v\n", err)
		}
	}

	var ref pgtype.Int8
	if req.ReferenceId != 0 {
		ref.Int64 = req.ReferenceId
		ref.Valid = true
	}
	fmt.Println(req.SortOrder)
	comments, err := executor.ListComments(ctx, sqlc.ListCommentsParams{VideoID: req.VideoId, Offset: req.Page * 10, UserID: userId, ReferenceID: ref, SortOrder: int32(req.SortOrder)})
	if err != nil {
		return nil, common.ErrInternal(err)
	}
	cs := make([]*proto.Comment, len(comments))
	for i, c := range comments {
		cs[i] = &proto.Comment{Id: c.ID, UserId: c.UserID, Content: c.Content, Username: c.Username.String, Likes: c.Likes, Dislikes: c.Dislikes, Reaction: c.Type.Int32, VideoId: c.VideoID, ReferenceId: c.ReferenceID.Int64, Replies: c.Replies}
	}
	return &proto.ListCommentsResponse{Comments: cs}, nil
}
