package handlers

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc/status"

	"videoapp/cmd/web/frontend"
	common "videoapp/internal/errors"
	"videoapp/internal/generated/proto"
)

func profile(c *fiber.Ctx) error {
	session := c.Cookies("session")
	username := c.Params("username")

	old := time.Now()
	user, err := deps.Clients.Users.Get(c.Context(), &proto.UsersGetRequest{Session: session, Username: username})
	fmt.Printf("Users:Get took %v\n", time.Since(old))
	if err != nil {
		err := status.Convert(err)
		if errors.Is(err.Err(), common.ErrUserNotFound) {
			return c.SendStatus(404)
		}
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Error:" + err.Err().Error())
	}
	res, err := deps.Clients.Videos.GetUserVideos(c.Context(), &proto.GetUserVideosRequest{Id: user.User.Id, Session: session})
	if err == nil {
		return Render(c, frontend.Profile(getUser(c), res.Videos, user))
	}

	status, ok := status.FromError(err)
	if ok && errors.Is(status.Err(), common.ErrSessionNotFound) || errors.Is(status.Err(), common.ErrSessionWrongSize) {
		c.ClearCookie("session")
	} else {
		fmt.Printf("failed to get user: %v\n", err)
	}
	return c.Redirect("/sign/in")
}
func getStages(c *fiber.Ctx) error {
	var ids []string
	if err := c.BodyParser(&ids); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}
	idInts := make([]int64, len(ids))
	for i, id := range ids {
		var err error
		idInts[i], err = strconv.ParseInt(id, 10, 0)
		if err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}
	}
	res, err := deps.Clients.Videos.GetStages(c.Context(), &proto.VideosGetStagesRequest{Ids: idInts})
	if shouldReturn := unwrapGrpcError(c, err, 400); shouldReturn {
		return nil
	}
	return c.SendString(res.Result)
}

// Handles both followers and follows
func getFollowsModal(c *fiber.Ctx) error {
	idString := c.Params("id")
	id, err := strconv.ParseInt(idString, 10, 0)
	if err != nil {
		return c.SendString("Invalid id")
	}

	pageString := c.Query("page", "0")
	page, err := strconv.Atoi(pageString)
	if err != nil {
		return c.SendString("Invalid page")
	}

	var result *proto.FollowUsers
	reqType := c.Params("type")
	switch reqType {
	case "followers":
		result, err = deps.Clients.Users.GetFollowers(c.Context(), &proto.GetFollowsRequest{UserId: id, Page: int32(page)})
	case "follows":
		result, err = deps.Clients.Users.GetFollowing(c.Context(), &proto.GetFollowsRequest{UserId: id, Page: int32(page)})
	default:
		return c.Next()
	}

	if err != nil {
		return c.SendString("There was an error: " + status.Convert(err).Message())
	}
	// Use to test infinite scrolling
	// for range 10 {
	// 	result.Users = append(result.Users, &proto.FollowUser{UserId: 1, CreatedAt: timestamppb.Now(), Username: "demo user"})
	// }
	return Render(c, frontend.FollowUserList(result, id, reqType, page+1))
}
