package handlers

import (
	"database/sql"
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"videoapp/cmd/web/frontend"
	"videoapp/internal/generated/proto"
)

func viewVideo(c *fiber.Ctx) error {
	idInt, err := strconv.ParseInt(c.Params("id"), 10, 0)
	if err != nil {
		return c.SendString("Invalid video id")
	}
	v, err := deps.Clients.Videos.Get(ctx, &proto.GetVideoRequest{Session: c.Cookies("session"), Id: idInt})
	if err != nil {
		status := status.Convert(err)
		if status.Code() == codes.PermissionDenied || status.Code() == codes.Unauthenticated {
			return Render(c, frontend.VideoError(getUser(c), "This video is private"))
		}
		return Render(c, frontend.VideoError(getUser(c), "Unknown Error: "+status.Message()))
	}
	return Render(c, frontend.ViewVideo(getUser(c), v))
}
func videoStatus(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 0)
	if err != nil {
		return c.SendString("Invalid id")
	}
	stage, err := deps.Clients.Videos.GetStage(ctx, &proto.VideosGetStageRequest{Session: c.Params("session"), Id: id})
	if shouldReturn := unwrapGrpcError(c, err, 200); shouldReturn {
		return nil
	}
	return Render(c, frontend.VideoStatusPoller(&proto.GetVideoResponse{Id: id, UploadId: stage.UploadId, Stage: stage.Stage, AuthorizationPayload: stage.AuthorizationPayload, AuthorizationSignature: stage.AuthorizationSignature}))
}
func extraUserInfo(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 0)
	if err != nil {
		return c.SendStatus(400)
	}
	user, err := deps.Clients.Users.GetById(c.Context(), &proto.UsersGetByIdRequest{Session: c.Cookies("session"), Id: id})
	if errors.Is(err, sql.ErrNoRows) {
		return c.SendString("User not found")
	} else if err != nil {
		return c.Status(400).SendString("Unknown error: " + err.Error())
	}
	return Render(c, frontend.ExtraUserInfo(user, getUser(c)))
}

func react(c *fiber.Ctx) error {
	id := c.Params("id")
	idInt, err := strconv.ParseInt(id, 10, 0)
	if err != nil {
		return c.SendStatus(400)
	}

	_type := c.Params("type")
	typeInt, err := strconv.ParseInt(_type, 10, 0)
	if err != nil {
		return c.SendStatus(400)
	}

	if _, err := deps.Clients.Videos.React(c.Context(), &proto.ReactRequest{Session: c.Cookies("session"), VideoId: idInt, Type: int32(typeInt)}); err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return c.SendStatus(200)
}
func deleteReaction(c *fiber.Ctx) error {
	id := c.Params("id")
	idInt, err := strconv.ParseInt(id, 10, 0)
	if err != nil {
		return c.SendStatus(400)
	}

	if _, err := deps.Clients.Videos.RemoveReaction(c.Context(), &proto.RemoveReactionRequest{Session: c.Cookies("session"), VideoId: idInt}); err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return c.SendStatus(200)
}
