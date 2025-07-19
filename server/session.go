package main

import (
	"context"
	"database/sql"
	"errors"
	"videoapp/proto"
	sqlc "videoapp/sql"
	"videoapp/utils"

	"golang.org/x/crypto/bcrypt"
)

const SESSION_TOKEN_LENGTH = 32

type sessionServer struct {
	proto.UnimplementedSessionsServer
}

func (s *sessionServer) Create(ctx context.Context, creds *proto.Crededentials) (*proto.Session, error) {
	if !utils.IsEmailValid(creds.Email) {
		return nil, ErrInvalidEmail
	}
	if !utils.Between(len(creds.Password), 8, 72) {
		return nil, ErrPasswordWrongSize
	}

	user, err := executor.GetPasswordFromEmail(ctx, creds.Email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserWithEmailNotFound
	} else if err != nil {
		return nil, ErrInternal(err)
	}
	err = bcrypt.CompareHashAndPassword(user.Password, []byte(creds.Password))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return nil, ErrIncorrectPassword
	} else if err != nil {
		return nil, ErrInternal(err)
	}

	token, err := utils.GenerateString(SESSION_TOKEN_LENGTH)
	if err != nil {
		return nil, ErrInternal(err)
	}
	err = executor.CreateSession(ctx, sqlc.CreateSessionParams{Token: token, UserID: user.ID})
	if err != nil {
		return nil, ErrInternal(err)
	}
	return &proto.Session{Token: token}, nil
}
func (s *sessionServer) GetUser(ctx context.Context, req *proto.Session) (*proto.User, error) {
	if len(req.Token) != SESSION_TOKEN_LENGTH {
		return nil, ErrSessionWrongSize
	}
	user, err := executor.GetUserFromSession(ctx, req.Token)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSessionNotFound
	}
	return &proto.User{Id: user.ID, Email: user.Email, Username: user.Username, CreatedAt: user.CreatedAt, Flags: uint64(user.Flags)}, nil
}
func (s *sessionServer) Delete(ctx context.Context, req *proto.Session) (*proto.Empty, error) {
	if len(req.Token) != SESSION_TOKEN_LENGTH {
		return nil, ErrSessionWrongSize
	}
	err := executor.DeleteSession(ctx, req.Token)
	if err != nil {
		return nil, ErrInternal(err)
	}
	return nil, nil
}
