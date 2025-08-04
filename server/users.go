package main

import (
	"context"
	"database/sql"
	"errors"
	"time"
	"videoapp/proto"
	"videoapp/server/common"
	sqlc "videoapp/sql"
	"videoapp/utils"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
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

	id := snowflakeNode.Generate().Int64()
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, common.ErrInternal(err)
	}
	verifyCode, err := utils.GenerateVerifyCode()
	if err != nil {
		return nil, common.ErrInternal(err)
	}
	expireAt := time.Now().Add(15 * time.Minute)

	err = executor.CreateUser(ctx, sqlc.CreateUserParams{ID: id, Email: req.Email, Username: req.Username, Password: hash, VerifyCode: verifyCode, VerifyExpireAt: pgtype.Timestamptz{Time: expireAt, Valid: true}})
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
