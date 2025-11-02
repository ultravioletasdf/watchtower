package main

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	client "github.com/gorse-io/gorse-go"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mattn/go-sqlite3"
	"github.com/minio/minio-go/v7"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/types/known/timestamppb"

	common "videoapp/internal/errors"
	"videoapp/internal/generated/proto"
	sqlc "videoapp/internal/generated/sqlc"
	"videoapp/internal/generated/vips"
	"videoapp/internal/utils"
)

type userServer struct {
	proto.UnimplementedUsersServer
}

func (s *userServer) Create(ctx context.Context, req *proto.CreateRequest) (*proto.CreateResponse, error) {
	if !utils.IsEmailValid(req.Email) {
		return nil, common.ErrInvalidEmail
	}
	if !utils.Between(len(req.Username), 3, 32) {
		return nil, common.ErrUsernameWrongSize
	}
	if !utils.IsUsernameValid(req.Username) {
		return nil, common.ErrInvalidUsername
	}
	if !utils.Between(len(req.Password), 8, 72) {
		return nil, common.ErrPasswordWrongSize
	}

	id := snowflakeNode.Generate()
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, common.ErrInternal(err)
	}
	verifyCode, err := utils.GenerateVerifyCode()
	if err != nil {
		return nil, common.ErrInternal(err)
	}
	expireAt := time.Now().Add(15 * time.Minute)

	err = executor.CreateUser(ctx, sqlc.CreateUserParams{ID: id.Int64(), Email: req.Email, Username: req.Username, Password: hash, VerifyCode: verifyCode, VerifyExpireAt: pgtype.Timestamptz{Time: expireAt, Valid: true}})
	if err, ok := err.(sqlite3.Error); ok && err.ExtendedCode == sqlite3.ErrConstraintUnique {
		switch utils.ErrUniqueConstraintFieldName(err) {
		case "users.username":
			return nil, common.ErrUsernameTaken
		case "users.email":
			return nil, common.ErrEmailTaken
		}
	}
	if err != nil {
		return nil, common.ErrInternal(err)
	}
	if _, err := gorse.InsertUser(ctx, client.User{UserId: id.String(), Comment: req.Username}); err != nil {
		fmt.Println(err)
	}
	return &proto.CreateResponse{Id: uint64(id)}, nil
}
func (s *userServer) Verify(ctx context.Context, req *proto.VerifyRequest) (*proto.Empty, error) {
	if len(req.Session) != SESSION_TOKEN_LENGTH {
		return nil, common.ErrSessionWrongSize
	}
	user, err := executor.GetUserFromSession(ctx, req.Session)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, common.ErrSessionNotFound
	}
	if user.VerifyCode != req.Code {
		return nil, common.ErrIncorrectVerifyCode
	}
	err = executor.SetUserFlag(ctx, sqlc.SetUserFlagParams{ID: user.ID, Flags: user.Flags | int32(FlagVerified)})
	return &proto.Empty{}, common.ErrInternal(err)
}
func (s *userServer) Follow(ctx context.Context, req *proto.FollowRequest) (*proto.Empty, error) {
	return &proto.Empty{}, common.ErrInternal(executor.FollowUser(ctx, sqlc.FollowUserParams{Token: req.Session, UserID: req.IdToFollow}))
}
func (s *userServer) Unfollow(ctx context.Context, req *proto.FollowRequest) (*proto.Empty, error) {
	return &proto.Empty{}, common.ErrInternal(executor.UnfollowUser(ctx, sqlc.UnfollowUserParams{Token: req.Session, UserID: req.IdToFollow}))
}
func (s *userServer) GetFollowers(ctx context.Context, req *proto.GetFollowsRequest) (*proto.FollowUsers, error) {
	follows, err := executor.GetUserFollowers(ctx, sqlc.GetUserFollowersParams{UserID: req.UserId, Limit: 10, Offset: req.Page * 10})
	if err != nil {
		return nil, common.ErrInternal(err)
	}
	followsProto := make([]*proto.FollowUser, len(follows))
	for i := range follows {
		followsProto[i] = &proto.FollowUser{UserId: follows[i].FollowerID, CreatedAt: timestamppb.New(follows[i].CreatedAt.Time), Username: follows[i].Username.String}
	}
	return &proto.FollowUsers{Users: followsProto}, nil
}
func (s *userServer) GetFollowing(ctx context.Context, req *proto.GetFollowsRequest) (*proto.FollowUsers, error) {
	follows, err := executor.GetUserFollows(ctx, sqlc.GetUserFollowsParams{FollowerID: req.UserId, Limit: 10, Offset: req.Page * 10})
	if err != nil {
		return nil, common.ErrInternal(err)
	}
	followsProto := make([]*proto.FollowUser, len(follows))
	for i := range follows {
		followsProto[i] = &proto.FollowUser{UserId: follows[i].UserID, CreatedAt: timestamppb.New(follows[i].CreatedAt.Time), Username: follows[i].Username.String}
	}
	return &proto.FollowUsers{Users: followsProto}, nil
}
func (s *userServer) Get(ctx context.Context, req *proto.UsersGetRequest) (*proto.UsersGetResponse, error) {
	user, err := executor.GetUser(ctx, req.Username)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, common.ErrUserNotFound
	} else if err != nil {
		return nil, common.ErrInternal(err)
	}

	var isFollowing bool
	if req.Session != "" {
		f, _ := executor.IsFollowing(ctx, sqlc.IsFollowingParams{UserID: user.ID, Token: req.Session})
		isFollowing = f == 1
	}
	return &proto.UsersGetResponse{
		User: &proto.User{
			Id:             user.ID,
			Email:          user.Email,
			Username:       user.Username,
			CreatedAt:      timestamppb.New(user.CreatedAt.Time),
			Flags:          uint64(user.Flags),
			FollowerCount:  user.FollowerCount,
			FollowingCount: user.FollowingCount,
			DisplayName:    user.DisplayName.String,
			Description:    user.Description.String},
		IsFollowing: isFollowing,
	}, nil
}
func (s *userServer) GetById(ctx context.Context, req *proto.UsersGetByIdRequest) (*proto.UsersGetResponse, error) {
	user, err := executor.GetUserById(ctx, req.Id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, common.ErrUserNotFound
	} else if err != nil {
		return nil, common.ErrInternal(err)
	}

	var isFollowing bool
	if req.Session != "" {
		f, _ := executor.IsFollowing(ctx, sqlc.IsFollowingParams{UserID: user.ID, Token: req.Session})
		isFollowing = f == 1
	}
	return &proto.UsersGetResponse{
		User: &proto.User{
			Id:             user.ID,
			Email:          user.Email,
			Username:       user.Username,
			CreatedAt:      timestamppb.New(user.CreatedAt.Time),
			Flags:          uint64(user.Flags),
			FollowerCount:  user.FollowerCount,
			FollowingCount: user.FollowingCount,
			DisplayName:    user.DisplayName.String,
			Description:    user.Description.String},
		IsFollowing: isFollowing,
	}, nil
}
func (s *userServer) GetFollowingVideos(ctx context.Context, req *proto.GetFollowingVideosRequest) (*proto.GetFollowingVideosResponse, error) {
	videos, err := executor.GetUsersFollowingVideos(ctx, sqlc.GetUsersFollowingVideosParams{FollowerID: req.UserId, Offset: req.Page * 10})
	if err != nil {
		return nil, common.ErrInternal(err)
	}

	result := make([]*proto.Video, len(videos))
	for i, v := range videos {
		result[i] = &proto.Video{Id: v.ID, Title: v.Title, Visibility: proto.Visibility(v.Visibility), CreatedAt: timestamppb.New(v.CreatedAt.Time), ThumbnailId: v.ThumbnailID, Stage: proto.Stage(v.Stage), UserId: v.UserID, Username: v.Username.String}
	}
	return &proto.GetFollowingVideosResponse{Videos: result}, nil
}

func (s *userServer) UploadAvatar(ctx context.Context, req *proto.UploadAvatarRequest) (*proto.Empty, error) {
	user, err := executor.GetUserFromSession(ctx, req.Session)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, common.ErrSessionNotFound
	} else if err != nil {
		return nil, common.ErrInternal(err)
	}

	img, err := vips.NewImageFromBuffer(req.Data, vips.DefaultLoadOptions())
	if err != nil {
		return nil, common.ErrInvalidImage
	}
	defer img.Close()

	err = img.ThumbnailImage(512, &vips.ThumbnailImageOptions{Height: 512, Size: vips.SizeBoth, Crop: vips.InterestingAttention})
	if err != nil {
		return nil, common.ErrInvalidImage
	}

	buf := &writeCloser{bytes.NewBuffer(nil)}
	target := vips.NewTarget(buf)
	err = img.WebpsaveTarget(target, vips.DefaultWebpsaveTargetOptions())
	if err != nil {
		return nil, common.ErrInternal(err)
	}
	_, err = s3.PutObject(ctx, "avatars", fmt.Sprintf("%d.webp", user.ID), buf, int64(buf.Len()), minio.PutObjectOptions{})
	if err != nil {
		return nil, common.ErrInternal(err)
	}
	return &proto.Empty{}, nil
}

func (s *userServer) RemoveAvatar(ctx context.Context, req *proto.Session) (*proto.Empty, error) {
	user, err := executor.GetUserFromSession(ctx, req.Token)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, common.ErrSessionNotFound
	} else if err != nil {
		return nil, common.ErrInternal(err)
	}
	err = s3.RemoveObject(ctx, "avatars", fmt.Sprintf("%d.webp", user.ID), minio.RemoveObjectOptions{})
	return nil, common.ErrInternal(err)
}

func (s *userServer) UpdateProfile(ctx context.Context, req *proto.UpdateProfileRequest) (*proto.Empty, error) {
	if len(req.Session) != SESSION_TOKEN_LENGTH {
		return nil, common.ErrSessionWrongSize
	}

	id, err := executor.UpdateProfile(ctx,
		sqlc.UpdateProfileParams{
			DisplayName: utils.PgTextFromPointer(req.DisplayName),
			Description: utils.PgTextFromPointer(req.Description),
			Token:       req.Session,
		})
	if id == 0 {
		return nil, common.ErrSessionNotFound
	}
	return nil, common.ErrInternal(err)
}

func (s *userServer) ListRecommendations(ctx context.Context, req *proto.ListRecommendationsRequest) (*proto.ListRecommendationsResponse, error) {
	var uid int64
	if req.Session != "" {
		user, err := executor.GetUserFromSession(ctx, req.Session)
		if err == nil {
			uid = user.ID
		} else {
			fmt.Println("Failed to get user", err)
		}
	}
	videosIdStrings, err := gorse.GetRecommendOffSet(ctx, fmt.Sprint(uid), "", 10, int(req.Page)*10)
	if err != nil {
		return nil, common.ErrInternal(err)
	}
	videoIds := make([]int64, len(videosIdStrings))
	for i, s := range videosIdStrings {
		videoIds[i], _ = strconv.ParseInt(s, 10, 64)
	}
	videos, err := executor.GetVideoBulk(ctx, videoIds)
	if err != nil {
		return nil, common.ErrInternal(err)
	}
	response := make([]*proto.Video, len(videos))
	for i, v := range videos {
		response[i] = &proto.Video{UserId: v.UserID, Id: v.ID, Title: v.Title, Visibility: proto.Visibility(v.Visibility), ThumbnailId: v.ThumbnailID, Stage: proto.Stage(v.Stage), CreatedAt: timestamppb.New(v.CreatedAt.Time), Username: v.Username.String}
	}

	return &proto.ListRecommendationsResponse{Videos: response}, nil
}
