package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
	common "videoapp/internal/errors"
	"videoapp/internal/generated/proto"
	"videoapp/internal/generated/sqlc"

	client "github.com/gorse-io/gorse-go"
)

type reactionsServer struct {
	proto.UnimplementedReactionsServer
}

func (s *reactionsServer) React(ctx context.Context, req *proto.ReactRequest) (*proto.Empty, error) {
	user, err := executor.GetUserFromSession(ctx, req.Session)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, common.ErrUnauthorized
	} else if err != nil {
		return nil, common.ErrInternal(err)
	}

	if err := executor.React(ctx, sqlc.ReactParams{TargetID: req.VideoId, Type: req.Type, UserID: user.ID}); err != nil {
		return nil, common.ErrInternal(err)
	}
	switch req.Type {
	case 1:
		_, err := gorse.InsertFeedback(ctx, []client.Feedback{{UserId: fmt.Sprint(user.ID), ItemId: fmt.Sprint(req.VideoId), FeedbackType: "like", Timestamp: time.Now()}})
		if err != nil {
			fmt.Println("Couldn't insert feedback, ", err.Error())
		}
	case 2:
		_, err := gorse.DeleteFeedback(ctx, "like", fmt.Sprint(user.ID), fmt.Sprint(req.VideoId))
		if err != nil {
			fmt.Println("Couldn't delete feedback, ", err.Error())
		}
	}
	return nil, nil
}
func (s *reactionsServer) Remove(ctx context.Context, req *proto.RemoveRequest) (*proto.Empty, error) {
	user, err := executor.GetUserFromSession(ctx, req.Session)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, common.ErrUnauthorized
	} else if err != nil {
		return nil, common.ErrInternal(err)
	}

	if err := executor.RemoveReaction(ctx, sqlc.RemoveReactionParams{TargetID: req.VideoId, UserID: user.ID}); err != nil {
		return nil, common.ErrInternal(err)
	}

	if _, err := gorse.DeleteFeedback(ctx, "like", fmt.Sprint(user.ID), fmt.Sprint(req.VideoId)); err != nil {
		fmt.Println("Couldn't delete feedback, ", err.Error())
	}

	return nil, nil
}
