package common

import (
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrInvalidEmail          = status.Error(codes.InvalidArgument, "Email is formatted incorrectly")
	ErrEmailTaken            = status.Error(codes.AlreadyExists, "Email is taken")
	ErrUserWithEmailNotFound = status.Error(codes.NotFound, "User with that email was not found")

	ErrInvalidUsername   = status.Error(codes.InvalidArgument, "Username is formatted incorrectly")
	ErrUsernameTaken     = status.Error(codes.AlreadyExists, "Username is taken")
	ErrUsernameWrongSize = status.Error(codes.InvalidArgument, "Username must be between 3 and 32 characters")

	ErrPasswordWrongSize = status.Error(codes.InvalidArgument, "Password must be between 8 and 72 characters")
	ErrIncorrectPassword = status.Error(codes.InvalidArgument, "Incorrect password")

	ErrSessionWrongSize = status.Error(codes.InvalidArgument, "Session token is the wrong length")
	ErrSessionNotFound  = status.Error(codes.NotFound, "No user has that session")

	ErrIncorrectVerifyCode = status.Error(codes.InvalidArgument, "Verify code is incorrect")

	ErrTitleWrongSize       = status.Error(codes.InvalidArgument, "Title must be between 5 and 100 characters")
	ErrDescriptionWrongSize = status.Error(codes.InvalidArgument, "Description must be less than 1000 characters")
	ErrInvalidVisibility    = status.Error(codes.InvalidArgument, "Invalid visibility")

	ErrNoVideosFound = status.Error(codes.NotFound, "No videos were found")
	ErrNoUploadFound = status.Error(codes.NotFound, "That upload was not found")
	ErrVideoNotFound = status.Error(codes.NotFound, "The specified video was not found")

	ErrUnauthorized = status.Error(codes.PermissionDenied, "You do not have authorization to do that")
)

func ErrInternal(err error) error {
	if err == nil {
		return nil
	}
	log.Println(err)
	return status.Error(codes.Internal, err.Error())
}
