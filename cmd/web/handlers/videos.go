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

func createComment(c *fiber.Ctx) error {
	id := c.Params("id")
	idInt, err := strconv.ParseInt(id, 10, 0)
	if err != nil {
		return c.SendStatus(400)
	}

	comment := c.FormValue("comment")

	referenceId := c.Query("reference_id")
	var referenceIdInt int64
	if referenceId != "" {
		referenceIdInt, err = strconv.ParseInt(referenceId, 10, 0)
		if err != nil {
			return c.SendStatus(400)
		}
	}

	if referenceIdInt != 0 {
		// When there is a ReferenceId, tell HTMX to add the comment to its replies instead of the main video comments
		c.Set("HX-Retarget", "#replies_"+referenceId)
	}
	cmnt, err := deps.Clients.Videos.CreateComment(c.Context(), &proto.CreateCommentRequest{Session: c.Cookies("session"), VideoId: idInt, Content: comment, ReferenceId: referenceIdInt})
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return Render(c, frontend.Comment(cmnt, cmnt.UserId))
}
func listComments(c *fiber.Ctx) error {
	videoId := c.Params("id")
	videoIdInt, err := strconv.ParseInt(videoId, 10, 0)
	if err != nil {
		return c.SendStatus(400)
	}

	page := c.QueryInt("page", 0)
	referenceId := c.QueryInt("reference_id", 0)
	sortOrder := c.QueryInt("sort_order", 1)

	comments, err := deps.Clients.Videos.ListComments(c.Context(), &proto.ListCommentsRequest{VideoId: videoIdInt, Session: c.Cookies("session"), Page: int32(page), ReferenceId: int64(referenceId), SortOrder: proto.SortOrder(sortOrder)})
	if err != nil {
		return c.Status(500).SendString(status.Convert(err).Message())
	}
	var uid int64
	if u := getUser(c); u != nil {
		uid = u.Id
	}
	return Render(c, frontend.CommentList(comments, videoIdInt, int32(page), uid, int32(sortOrder)))
}

func getRecommendations(c *fiber.Ctx) error {
	page := c.QueryInt("page", 0)
	videos, err := deps.Clients.Users.ListRecommendations(c.Context(), &proto.ListRecommendationsRequest{Session: c.Cookies("session"), Page: int32(page)})
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return Render(c, frontend.VideoList(videos.Videos, page+1))
}

func editVideo(c *fiber.Ctx) error {
	videoId, err := c.ParamsInt("id")
	if err != nil {
		return c.SendStatus(400)
	}

	user := getUser(c)
	if user == nil {
		return c.SendStatus(401)
	}

	video, err := deps.Clients.Videos.Get(c.Context(), &proto.GetVideoRequest{Session: c.Cookies("session"), Id: int64(videoId)})
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	if video.UserId != user.Id {
		return c.SendStatus(fiber.StatusForbidden)
	}

	return Render(c, frontend.EditVideo(user, video))
}

func saveVideoChanges(c *fiber.Ctx) error {
	videoId, err := c.ParamsInt("id")
	if err != nil {
		return c.SendStatus(400)
	}

	session := c.Cookies("session")
	if session == "" {
		return c.Status(401).SendString("Please sign in to continue")
	}

	title := formValueToPointer(c, "title")
	description := formValueToPointer(c, "description")
	visibility := formValueToPointer(c, "visibility")
	var visibilityInt *proto.Visibility
	if visibility != nil {
		v, ok := proto.Visibility_value[*visibility]
		if !ok {
			return c.SendStatus(400)
		}
		temp := proto.Visibility(v)
		visibilityInt = &temp
	}
	thumbnailId := formValueToPointer(c, "thumbnail_id")
	var thumbnailIdInt *int64
	if thumbnailId != nil {
		v, err := strconv.ParseInt(*thumbnailId, 10, 64)
		if err != nil {
			return c.SendStatus(400)
		}
		thumbnailIdInt = &v
	}

	_, err = deps.Clients.Videos.Update(c.Context(), &proto.VideosUpdateRequest{Session: c.Cookies("session"), Id: int64(videoId), ThumbnailId: thumbnailIdInt, Title: title, Description: description, Visibility: visibilityInt})
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return Render(c, frontend.SoftSuccess("Saved"))

}
