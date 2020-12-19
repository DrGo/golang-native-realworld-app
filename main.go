package main

import (
	"log"
	"net/http"

	"github.com/drgo/realworld/errors"
	"github.com/drgo/realworld/sessions"
	"github.com/drgo/realworld/utils"
)

const (
	host        = "localhost"
	port        = "8080"
	apiVersion  = "v1"
	apiRoot     = "api"
	cookieName  = "session"
	maxLifeTime = 10 * 60 //10mins
)

type Ctx struct {
	Res     http.ResponseWriter
	Req     *http.Request
	Server  *server
	Session *sessions.Session
}

func (ctx *Ctx) Store() *Store {
	return ctx.Server.Store
}

func (ctx *Ctx) Authenticated(dx errors.Diag) bool {
	var err error
	if ctx.Session, err = ctx.Server.Sessions.Authenticate(ctx.Res, ctx.Req); err != nil {
		errors.Send(ctx.Res, errors.E(dx, err, http.StatusUnauthorized))
		return false
	}
	return true
}
func (ctx *Ctx) QueryParams(key string) (values []string, n int) {
	if values, ok := ctx.Req.URL.Query()[key]; ok {
		return values, len(values)
	}
	return []string{}, 0
}

type server struct {
	Store    *Store
	Sessions *sessions.Sessions
}

func (s *server) Finalize() {
	s.Sessions.Finalize()
}

// Error translates errors.Error to the format used by the realdworld app
// 	"errors":{
// 	  "body": [
// 		"can't be empty"
// 	  ]
// 	}
func (s *server) Error(w http.ResponseWriter, err error) {
	var rwErr struct {
		Errors struct {
			Body []string `json:"body"`
		} `json:"errors"`
	}
	status := http.StatusInternalServerError //default error code
	e, ok := err.(*errors.Error)
	if ok {
		rwErr.Errors.Body = append(rwErr.Errors.Body, e.Error())
		status = e.Status
	} else {
		//not an errors.Error
		rwErr.Errors.Body = append(rwErr.Errors.Body, err.Error())
	}
	utils.JSON(w, status, rwErr)
	// for the moment,log the error here
	log.Println(rwErr.Errors.Body)
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer errors.Recover(w)      //setup graceful error recovery
	if utils.CORSHandled(w, r) { // handle CORS
		return
	}
	dx := errors.D(r, "ServeHTTP")
	err := errors.E(dx, http.StatusNotFound)
	var head string
	head, r.URL.Path = utils.ShiftPath(r.URL.Path)

	if head == apiRoot {
		head, r.URL.Path = utils.ShiftPath(r.URL.Path)
		log.Println("serveHTTP", head, r.URL.Path)
		ctx := &Ctx{Res: w, Req: r, Server: s}
		switch head {
		case "test":
			err = ServeTest(w, r)
		case "users":
			// /api/users/*
			err = ServeUsers(ctx)
		case "profiles":
			// /api/profiles/*
			err = ServeProfiles(ctx)
		case "user":
			// /api/user/*
			err = ServeUser(ctx)
		case "articles":
			// /api/articles/*
			err = ServeArticles(ctx)
		case "tags":
			// /api/tags/*
			err = ServeTags(ctx)
		}
	}
	if err != nil {
		s.Error(w, err)
	}
}

func (s *server) Authenticate(ctx *Ctx) (*sessions.Session, error) {
	return s.Sessions.Authenticate(ctx.Res, ctx.Req)
}

// ServeTest handles "/api/test/*"
func ServeTest(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")
	//w.Write automatically calls w.WriteHeader(http.StatusOK) sending the contents of the Header
	utils.PrintRequestInfo(w, r)
	// Send(w, []byte(`{"message": "working!"}`))
	return nil
}

func main() {
	// Verbose logging with file name and line number
	log.SetFlags(log.Lshortfile)
	s := &server{
		Store:    mustNewStore("db/rw.db"),
		Sessions: sessions.NewSessionManager(cookieName, maxLifeTime),
	}
	defer s.Finalize()
	http.Handle("/", s)
	addr := host + ":" + port
	log.Printf("running on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil)) // use "localhost:8080" to suppress macos firewall permission
}
