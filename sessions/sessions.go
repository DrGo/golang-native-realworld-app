package sessions

import (
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/drgo/realworld/errors"
)

//FIXME: remove old/expired sessions

//Sessions is in-memory key-value session store
type Sessions struct {
	sync.RWMutex
	lastCleaned time.Time
	//max life of a cookie in seconds
	MaxLifeTime int
	CookieName  string
	store       map[string]*Session
	ticker      *time.Ticker
	done        chan interface{}
}

func NewSessionManager(cookieName string, maxLifeTime int) *Sessions {
	ss := &Sessions{
		CookieName:  cookieName,
		MaxLifeTime: maxLifeTime,
		store:       make(map[string]*Session),
		lastCleaned: time.Now(),
	}
	// schedule pruning of expired sessions
	ss.ticker = time.NewTicker(1000 * time.Millisecond)
	ss.done = make(chan interface{})
	go func() {
		for {
			select {
			case <-ss.done:
				return
			case _ = <-ss.ticker.C:
				ss.Prune(int64(ss.MaxLifeTime))
			}
		}
	}()
	return ss
}

// Finalize stops the sessions' garbage collector
// eg, defer ss.Finalize() after calling ss:=NewSessionManager(..)
func (ss *Sessions) Finalize() {
	ss.ticker.Stop()
	ss.done <- true
}

//GenSessionID returns a unique session ID
func (ss *Sessions) GenSessionID() string {
	p1 := time.Now().UnixNano()
	p2 := rand.Int63()
	return strconv.FormatInt(p1, 16) + strconv.FormatInt(p2, 16)
	/*
		// one possibility using a string (also uuid)
		//source: https://astaxie.gitbooks.io/build-web-application-with-golang/content/en/06.2.html
		b := make([]byte, 32)
		if _, err := io.ReadFull(rand.Reader, b); err != nil {
			return ""
		}
		return base64.URLEncoding.EncodeToString(b)
	*/
}

func (ss *Sessions) Add(uid int64) *Session {
	s := &Session{ID: ss.GenSessionID(),
		UserID:     uid,
		LastActive: time.Now(),
		Sessions:   ss,
	}
	ss.Lock()
	ss.store[s.ID] = s
	ss.Unlock()
	return s
}

func (ss *Sessions) GetExisting(sessionID string) *Session {
	ss.RLock()
	s, ok := ss.store[sessionID]
	s.LastActive = time.Now()
	ss.RUnlock()
	if ok {
		return s
	}
	return nil
}

// Authenticate get an existing session or create a new one if none exists
func (ss *Sessions) Authenticate(w http.ResponseWriter, r *http.Request) (*Session, error) {
	// does the request carry a session cookie
	c, err := r.Cookie(ss.CookieName)
	if err != nil {
		return nil, errors.E(err)
	}
	// do we have a valid active session with this id
	s := ss.GetExisting(c.Value)
	if s == nil { // no valid session
		return nil, errors.Errorf("no such session")
	}
	//FIXME: test
	s.LastActive = time.Now()
	// logged in
	return s, nil
}

// func (ss *Sessions) GetOrStart(w http.ResponseWriter, r *http.Request) *Session {
// 	// does the request carry a session cookie
// 	c, err := r.Cookie(ss.CookieName)
// 	if err != nil || c.Value == "" {
// 		return nil
// 	}
// 	// do we have a valid active session with this id
// 	s := ss.GetExisting(c.Value)
// 	if s != nil {
// 		return s
// 	}
// 	// not logged in, so create a session
// 	s = ss.Add()
// 	// add cookie with this session ID
// 	c = &http.Cookie{
// 		Name:  ss.CookieName,
// 		Value: s.ID,
// 		Path:  "/", //otherwise it defaults to r.URL
// 		// Secure : true,
// 		HttpOnly: true, //do not allow JS code to access it
// 		MaxAge:   int(ss.MaxLifeTime),
// 	}
// 	http.SetCookie(w, c)
// 	return s
// }

// user session
type Session struct {
	ID         string
	LastActive time.Time
	UserID     int64
	Sessions   *Sessions
}

func (s *Session) NewCookie() *http.Cookie {
	c := &http.Cookie{
		Name:  s.Sessions.CookieName,
		Value: s.ID,
		Path:  "/", //otherwise it defaults to dx
		// Secure:   true,
		HttpOnly: true, //do not allow JS code to access it
		MaxAge:   s.Sessions.MaxLifeTime,
	}
	// uncomment if compatability with IE is needed
	// if c.MaxAge > 0 {
	// 	c.Expires = time.Now().Add(time.Duration(c.MaxAge) * time.Second)
	// } else if c.MaxAge < 0 {
	// 	// Set it to the past to expire now.
	// 	c.Expires = time.Unix(1, 0)
	// }
	return c
}

// Prune deletes expired sessions by deleting entries in the sessions store that
// have expired (time.now > time last active + MaxLifeTime)
func (ss *Sessions) Prune(maxLifeTime int64) {
	ss.Lock()
	defer ss.Unlock()
	for sid, s := range ss.store {
		if time.Now().Unix() > (s.LastActive.Unix() + maxLifeTime) {
			delete(ss.store, sid)
		}
	}
}

//TODO: refresh session in response to /refresh
// func Refresh(w http.ResponseWriter, r *http.Request) {
// 	// (BEGIN) The code uptil this point is the same as the first part of the `Welcome` route
// 	  c, err := r.Cookie("session_token")
// 	  if err != nil {
// 		  if err == http.ErrNoCookie {
// 			  w.WriteHeader(http.StatusUnauthorized)
// 			  return
// 		  }
// 		  w.WriteHeader(http.StatusBadRequest)
// 		  return
// 	  }
// 	  sessionToken := c.Value

// 	  response, err := cache.Do("GET", sessionToken)
// 	  if err != nil {
// 		  w.WriteHeader(http.StatusInternalServerError)
// 		  return
// 	  }
// 	  if response == nil {
// 		  w.WriteHeader(http.StatusUnauthorized)
// 		  return
// 	  }
// 	  // (END) The code uptil this point is the same as the first part of the `Welcome` route

// 	  // Now, create a new session token for the current user
// 	  newSessionToken := uuid.NewV4().String()
// 	  _, err = cache.Do("SETEX", newSessionToken, "120", fmt.Sprintf("%s",response))
// 	  if err != nil {
// 		  w.WriteHeader(http.StatusInternalServerError)
// 		  return
// 	  }

// 	  // Delete the older session token
// 	  _, err = cache.Do("DEL", sessionToken)
// 	  if err != nil {
// 		  w.WriteHeader(http.StatusInternalServerError)
// 		  return
// 	  }

// 	  // Set the new token as the users `session_token` cookie
// 	  http.SetCookie(w, &http.Cookie{
// 		  Name:    "session_token",
// 		  Value:   newSessionToken,
// 		  Expires: time.Now().Add(120 * time.Second),
// 	  })
//   }
