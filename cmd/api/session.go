package main

import (
	"context"
	"database/sql"
	"errors"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/types/known/timestamppb"

	common "videoapp/internal/errors"
	"videoapp/internal/generated/proto"
	sqlc "videoapp/internal/generated/sqlc"
	"videoapp/internal/utils"
)

const SESSION_TOKEN_LENGTH = 32

type sessionServer struct {
	proto.UnimplementedSessionsServer
}

func (s *sessionServer) Create(ctx context.Context, creds *proto.Crededentials) (*proto.Session, error) {
	if !utils.IsEmailValid(creds.Email) {
		return nil, common.ErrInvalidEmail
	}
	if !utils.Between(len(creds.Password), 8, 72) {
		return nil, common.ErrPasswordWrongSize
	}

	user, err := executor.GetPasswordFromEmail(ctx, creds.Email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, common.ErrUserWithEmailNotFound
	} else if err != nil {
		return nil, common.ErrInternal(err)
	}
	err = bcrypt.CompareHashAndPassword(user.Password, []byte(creds.Password))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return nil, common.ErrIncorrectPassword
	} else if err != nil {
		return nil, common.ErrInternal(err)
	}

	token, err := utils.GenerateString(SESSION_TOKEN_LENGTH)
	if err != nil {
		return nil, common.ErrInternal(err)
	}
	err = executor.CreateSession(ctx, sqlc.CreateSessionParams{Token: token, UserID: user.ID})
	if err != nil {
		return nil, common.ErrInternal(err)
	}
	return &proto.Session{Token: token}, nil
}
func (s *sessionServer) GetUser(ctx context.Context, req *proto.Session) (*proto.User, error) {
	if len(req.Token) != SESSION_TOKEN_LENGTH {
		return nil, common.ErrSessionWrongSize
	}
	user, err := executor.GetUserFromSession(ctx, req.Token)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, common.ErrSessionNotFound
	}
	return &proto.User{Id: user.ID, Email: user.Email, Username: user.Username, CreatedAt: timestamppb.New(user.CreatedAt.Time), Flags: uint64(user.Flags), FollowerCount: user.FollowerCount, FollowingCount: user.FollowingCount, DisplayName: user.DisplayName.String, Description: user.Description.String}, nil
}
func (s *sessionServer) Delete(ctx context.Context, req *proto.Session) (*proto.Empty, error) {
	if len(req.Token) != SESSION_TOKEN_LENGTH {
		return nil, common.ErrSessionWrongSize
	}
	err := executor.DeleteSession(ctx, req.Token)
	if err != nil {
		return nil, common.ErrInternal(err)
	}
	return nil, nil
}
