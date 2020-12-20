package main

import (
	"net/http"

	"github.com/drgo/realworld/errors"
	"github.com/drgo/realworld/sessions"
)

// Ctx holds info required by each route handler
type Ctx struct {
	Res     http.ResponseWriter
	Req     *http.Request
	Server  *server
	Session *sessions.Session
}

// Store returns the server store
func (ctx *Ctx) Store() *Store {
	return ctx.Server.Store
}

// Authenticated convenient way to check if a user is authenticated.
// if true it updates the Ctx.Session field to the user's session
// if false it sends an unauthorized error
func (ctx *Ctx) Authenticated(dx errors.Diag) bool {
	var err error
	if ctx.Session, err = ctx.Server.Sessions.Authenticate(ctx.Res, ctx.Req); err != nil {
		errors.Send(ctx.Res, errors.E(dx, err, http.StatusUnauthorized))
		return false
	}
	return true
}

//QueryParams convenient way to extract query parameters from the current request
func (ctx *Ctx) QueryParams(key string) (values []string, n int) {
	if values, ok := ctx.Req.URL.Query()[key]; ok {
		return values, len(values)
	}
	return []string{}, 0
}
