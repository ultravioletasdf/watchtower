package main

import (
	"context"
	common "videoapp/internal/errors"
	"videoapp/internal/generated/proto"
	"videoapp/internal/generated/sqlc"
)

type reactionsServer struct {
	proto.UnimplementedReactionsServer
}

func (s *reactionsServer) React(ctx context.Context, req *proto.ReactRequest) (*proto.Empty, error) {
	return nil, common.ErrInternal(executor.React(ctx, sqlc.ReactParams{TargetID: req.VideoId, Type: req.Type, Token: req.Session}))
}
func (s *reactionsServer) Remove(ctx context.Context, req *proto.RemoveRequest) (*proto.Empty, error) {
	return nil, common.ErrInternal(executor.RemoveReaction(ctx, sqlc.RemoveReactionParams{TargetID: req.VideoId, Token: req.Session}))
}
