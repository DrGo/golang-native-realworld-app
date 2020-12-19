package main

import (
	"net/http"

	"github.com/drgo/realworld/errors"
	"github.com/drgo/realworld/utils"
)

// routes
// GET /api/args

// ServeTags handles "/api/user/*"
func ServeTags(ctx *Ctx) error {
	dx := errors.D(ctx.Req, "ServeTags")
	if dx.Path != "/" {
		return errors.E(dx, http.StatusNotFound)
	}
	switch dx.Method {
	case "GET": // GET /api/tags
		dx := errors.D(ctx.Req, "tagsList")
		json, err := ctx.Store().ListTagsJSON()
		if err != nil {
			return errors.E(dx, err, http.StatusNotFound)
		}
		ctx.Res.Header().Set("Content-Type", "application/json")
		utils.Send(ctx.Res, json) //http.StatusFound,
	default:
		return errors.E(dx, http.StatusMethodNotAllowed)
	}
	return nil
}
