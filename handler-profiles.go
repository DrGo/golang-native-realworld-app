package main

import (
	"net/http"

	"github.com/drgo/realworld/errors"
	"github.com/drgo/realworld/utils"
)

// handles routes
// POST /api/profiles/:username

// ServeProfiles handles "/api/profiles/*"
func ServeProfiles(ctx *Ctx) error {
	dx := errors.D(ctx.Req, "ServeProfiles")
	var userName, cmd string
	_ = errors.Debug && errors.Logln("ServeProfiles: initial path", dx.Path)
	userName, ctx.Req.URL.Path = utils.ShiftPath(dx.Path)
	if userName == "" { //userName required
		return errors.E(dx, http.StatusNotFound)
	}
	cmd, ctx.Req.URL.Path = utils.ShiftPath(ctx.Req.URL.Path)
	switch cmd {
	case "":
		// GET /api/profiles/:username
		json, err := ctx.Store().GetUserProfileJSON(userName)
		if err != nil {
			return errors.E(dx, err, http.StatusInternalServerError)
		}
		return utils.SendJSON(ctx.Res, http.StatusOK, json)
	case "follow":
		// handle follow
		//	default:
	}
	return errors.E(dx, http.StatusNotFound)
	// if dx.Method == "POST" {
	// 	return usersLogin(ctx)
	// }
	// return errors.E(dx, http.StatusMethodNotAllowed)
}
