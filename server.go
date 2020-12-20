package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/drgo/realworld/errors"
	"github.com/drgo/realworld/sessions"
	"github.com/drgo/realworld/utils"
)

type ServerOptions struct {
	CookieName   string
	MaxLifeTime  int
	DatabaseName string
	Addr         string
}

type server struct {
	Store    *Store
	Sessions *sessions.Sessions
	mux      *http.ServeMux
	srv      *http.Server
}

func NewServer(opts *ServerOptions) *server {
	s := &server{
		Store:    mustNewStore(opts.DatabaseName),
		Sessions: sessions.NewSessionManager(opts.CookieName, opts.MaxLifeTime),
		srv: &http.Server{
			Addr: opts.Addr,
			// Handler: http.NewServeMux(),
			//FIXME: redirect server logs to app logs?
			//ErrorLog:     logger,
			// ReadTimeout:  5 * time.Second,
			// WriteTimeout: 10 * time.Second,
			// IdleTimeout:  15 * time.Second,
		},
	}
	// replace srv.Handler.HandleFunc... if s.sev.Handler is initialized
	http.HandleFunc("/", s.ServeHTTP)
	return s
}

func (s *server) Finalize() {
	s.Sessions.Finalize()
}

func (s *server) Start() {
	// monitor for interruptions
	stop := make(chan os.Signal, 1) //for os.signals
	signal.Notify(stop, os.Interrupt)
	done := make(chan interface{}) // to ensure that we do not exist before s.Shutdown() is done
	go func() {
		<-stop // blocks until it receives an interrupt signal
		errors.Logf("\nserver stopping...\n")
		// allow time for all goroutines to finish
		ctxWait, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.srv.SetKeepAlivesEnabled(false) //disable keepAlive
		if err := s.srv.Shutdown(ctxWait); err != nil {
			errors.Fatal(err)
		}
		close(done)
	}()
	errors.Logf("running on %s ...\n", s.srv.Addr)
	if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		errors.Fatal(err)
	}
	<-done //block until shutdown is complete
	errors.Logf("\nserver stopped...\n")
}

//FIXME: allow intentional shutdown eg for testing?
// func (s *server) Shutdown(ctx context.Context) {
// }

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
	_ = errors.Debug && errors.Logln(rwErr.Errors.Body)
}

// func (s *server) NewContext(w http.ResponseWriter, r *http.Request) *Ctx {
// 	return &Ctx{Res: w, Req: r, Server: s}
// }

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer errors.Recover(w)      //setup graceful error recovery
	if utils.CORSHandled(w, r) { // handle CORS
		return
	}
	dx := errors.D(r, "ServeHTTP")
	err := errors.E(dx, http.StatusNotFound)
	var head string
	head, r.URL.Path = utils.ShiftPath(r.URL.Path)
	// fmt.Println("debug:", errors.Debug)
	if head == apiRoot {
		head, r.URL.Path = utils.ShiftPath(r.URL.Path)
		_ = errors.Debug && errors.Logln("serveHTTP", head, r.URL.Path)
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

// ServeTest handles "/api/test"
func ServeTest(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")
	//w.Write automatically calls w.WriteHeader(http.StatusOK) sending the contents of the Header
	utils.PrintRequestInfo(w, r)
	return nil
	// Send(w, []byte(`{"message": "working!"}`))
}
