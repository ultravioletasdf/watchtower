package handlers

import (
	"fmt"
	"time"
	"videoapp/clients/htmx/frontend"
	"videoapp/proto"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc/status"
)

func signIn(c *fiber.Ctx) error {
	return Render(c, frontend.Sign(frontend.SignText{
		Endpoint:            "/sign/in",
		Title:               "Sign in to WatchTower",
		Heading:             "Welcome Back",
		Description:         "Sign in to continue to WatchTower",
		Button:              "Sign In",
		AskForUsername:      false,
		AskIfForgotPassword: true,
		Alternative: frontend.SignTextAlternative{
			Description: "Don't have an account?",
			Link: frontend.Link{
				Text: "Sign Up",
				Href: "/sign/up",
			},
		},
	}))
}
func postSignIn(c *fiber.Ctx) error {
	email := c.FormValue("email")
	password := c.FormValue("password")

	session, err := deps.Clients.Sessions.Create(ctx, &proto.Crededentials{Email: email, Password: password})
	if err != nil {
		status, grpcError := status.FromError(err)
		if grpcError {
			return c.SendString(status.Message())
		}
		return c.SendString(fmt.Sprintf("Unknown Error: %s", err.Error()))
	}
	yearAhead := time.Now().Add(time.Hour * 24 * 365)
	c.Cookie(&fiber.Cookie{Name: "session", Value: session.Token, Path: "/", HTTPOnly: true, Secure: true, Expires: yearAhead})
	c.Set("HX-Redirect", "/")
	return nil
}
func signUp(c *fiber.Ctx) error {
	return Render(c, frontend.Sign(frontend.SignText{
		Endpoint:    "/sign/up",
		Title:       "Create a WatchTower account",
		Heading:     "Create an account",
		Description: "Create an account to use WatchTower", Button: "Continue",
		AskForUsername: true,
		Alternative: frontend.SignTextAlternative{
			Description: "Already have an account?",
			Link: frontend.Link{
				Text: "Sign In",
				Href: "/sign/in",
			},
		},
	}))
}
func postSignUp(c *fiber.Ctx) error {
	email := c.FormValue("email")
	username := c.FormValue("username")
	password := c.FormValue("password")

	_, err := deps.Clients.Users.Create(ctx, &proto.CreateRequest{Email: email, Username: username, Password: password})
	if err != nil {
		status, grpcError := status.FromError(err)
		if grpcError {
			return c.SendString(status.Message())
		}
		return c.SendString(fmt.Sprintf("Unknown Error: %s", err.Error()))
	}
	session, err := deps.Clients.Sessions.Create(ctx, &proto.Crededentials{Email: email, Password: password})
	if err != nil {
		status, grpcError := status.FromError(err)
		if grpcError {
			return c.SendString(status.Message())
		}
		return c.SendString(fmt.Sprintf("Unknown Error: %s", err.Error()))
	}

	yearAhead := time.Now().Add(time.Hour * 24 * 365)
	c.Cookie(&fiber.Cookie{Name: "session", Value: session.Token, Path: "/", HTTPOnly: true, Secure: true, Expires: yearAhead})
	c.Set("HX-Redirect", "/")
	return nil
}
func signOut(c *fiber.Ctx) error {
	_, err := deps.Clients.Sessions.Delete(ctx, &proto.Session{Token: c.Cookies("session")})
	if err != nil {
		fmt.Printf("Failed to delete session: %v\n", err)
	}
	c.ClearCookie("session")
	return c.Redirect("/")
}
