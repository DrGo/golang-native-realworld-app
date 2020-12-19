package main

import (
	"net/http"

	"github.com/drgo/realworld/errors"
	"github.com/drgo/realworld/sessions"
	"github.com/drgo/realworld/utils"
)

// routes
// GET /api/user
// PUT /api/user

// ServeUser handles "/api/user/*"
func ServeUser(ctx *Ctx) error {
	dx := errors.D(ctx.Req, "serveuser")
	if dx.Path != "/" {
		return errors.E(dx, http.StatusNotFound)
	}
	session, err := ctx.Server.Authenticate(ctx)
	if err != nil {
		return errors.E(dx, err, http.StatusUnauthorized)
	}
	switch dx.Method {
	case "GET": // GET /api/user
		return userGetCurrent(ctx, session)
	case "PUT": // PUT /api/user
		return userUpdate(ctx)
	default:
		return errors.E(dx, http.StatusMethodNotAllowed)
	}
}

// GET /api/user
// Authentication required, returns a User that's the current user
func userGetCurrent(ctx *Ctx, s *sessions.Session) error {
	dx := errors.D(ctx.Req, "getcurrent")
	json, err := ctx.Store().GetUserJSON(s.UserID)
	if err != nil {
		return errors.E(dx, http.StatusNotFound)
	}
	utils.SendJSON(ctx.Res, http.StatusFound, json)
	return nil
}

func userUpdate(ctx *Ctx) error {
	utils.Send(ctx.Res, []byte(`{"message": "userUpdate!"}`))
	return nil
}
