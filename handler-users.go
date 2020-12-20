package main

import (
	"net/http"

	"github.com/drgo/realworld/errors"
	"github.com/drgo/realworld/utils"
)

// handles routes
// POST /api/users/login
// POST /api/users

type userModel struct {
	ID       int64   `json:"id"`
	Username string  `json:"username"`
	Email    string  `json:"email"`
	Bio      string  `json:"bio"`
	Image    *string `json:"image"`
	Token    string  `json:"token"`
}

// used to decode payload for login and register
type credentials struct {
	User struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	} `json:"user"`
}

// ServeUsers handles "/api/users/*"
func ServeUsers(ctx *Ctx) error {
	dx := errors.D(ctx.Req, "ServeUsers")
	// POST /api/users
	if dx.Path == "/" {
		if dx.Method == "POST" {
			return usersRegister(ctx)
		}
		return errors.E(dx, http.StatusMethodNotAllowed)
	}
	var head string
	head, ctx.Req.URL.Path = utils.ShiftPath(dx.Path)
	// POST /api/users/login
	if head == "login" {
		if dx.Method == "POST" {
			return usersLogin(ctx)
		}
		return errors.E(dx, http.StatusMethodNotAllowed)
	}
	return errors.E(dx, http.StatusNotFound)
}

// POST /api/users/login
// Example request body:
// {
//   "user":{
//     "email": "jake@jake.jake",
//     "password": "jakejake"
//   }
// }
// No authentication required, returns a User
// Required fields: email, password
func usersLogin(ctx *Ctx) error {
	dx := errors.D(ctx.Req, "login")
	var creds credentials
	if err := utils.DecodeJSONBody(ctx.Res, ctx.Req, &creds); err != nil {
		return errors.E(dx, err, http.StatusBadRequest)
	}
	_ = errors.Debug && errors.Logln("creds:", creds)
	row, err := ctx.Store().SignInByEmailAndPassword(creds.User.Email, creds.User.Password)
	if err != nil {
		return errors.E(dx, err, http.StatusUnauthorized)
	}
	// create session token to store this user id
	s := ctx.Server.Sessions.Add(row["id"].(int64))
	// send token as cookie
	c := s.NewCookie()
	_ = errors.Debug && errors.Logln("cookie", c)
	http.SetCookie(ctx.Res, c)
	json, err := ctx.Store().GetUserJSON(s.UserID)
	if err != nil {
		return errors.E(dx, http.StatusNotFound)
	}
	return utils.SendJSON(ctx.Res, http.StatusOK, json)
}

// POST /api/users
// Example request body:
// {
//   "user":{
//     "username": "Jacob",
//     "email": "jake@jake.jake",
//     "password": "jakejake"
//   }
// }
// No authentication required, returns a User
// Required fields: email, username, password
func usersRegister(ctx *Ctx) error {
	dx := errors.D(ctx.Req, "register")
	var creds credentials
	if err := utils.DecodeJSONBody(ctx.Res, ctx.Req, &creds); err != nil {
		return errors.E(dx, err, http.StatusBadRequest)
	}
	id, err := ctx.Store().CreateUser(&creds)
	if err != nil {
		return errors.E(dx, err, http.StatusInternalServerError)
	}
	json, err := ctx.Store().GetUserJSON(id)
	return utils.SendJSON(ctx.Res, http.StatusFound, json)
}
