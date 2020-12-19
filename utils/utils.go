package utils

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode"
)

// ShiftPath splits off the first component of p, which will be cleaned of
// relative components before processing. head will never contain a slash and
// tail will always be a rooted path without trailing slash.
// eg, "/foo/bar/baz" gives "foo", "/bar/baz".
// modified from source: https://blog.merovius.de/2017/06/18/how-not-to-use-an-http-router.html
func ShiftPath(path string) (head, tail string) {
	// p = path.Clean("/" + p)
	i := strings.Index(path[1:], "/") + 1
	if i <= 0 {
		return path[1:], "/"
	}
	// log.Println("head:", path[1:i], "rest:", path[i:])
	return path[1:i], path[i:]
}

// Send writes to http.ResponseWriter and handle errors
func Send(w http.ResponseWriter, b []byte) {
	n, err := w.Write(b)
	if err != nil {
		log.Printf("write failed (%d chars read): %v", n, err)
		panic(err) //temp
	}
}

// SendJSON writes JSON to http.ResponseWriter and handle errors
func SendJSON(w http.ResponseWriter, status int, b []byte) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err := w.Write(b)
	return err
}

func PrintRequestInfo(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s %s %s \n", r.Method, r.URL, r.Proto)
	//Iterate over all header fields
	fmt.Fprintf(w, "Header fields \n")
	for k, v := range r.Header {
		fmt.Fprintf(w, "%q:", k)
		if len(v) == 1 {
			fmt.Fprintf(w, "%q\n", v[0])
		} else {
			fmt.Fprintf(w, "\n")
			for _, entry := range v {
				fmt.Fprintf(w, "%q\n", entry)
			}
		}
	}
	fmt.Fprintf(w, "Host = %q\n", r.Host)
	fmt.Fprintf(w, "RemoteAddr= %q\n", r.RemoteAddr)
	// //Get value for a specified token
	// fmt.Fprintf(w, "\n\nFinding value of \"Accept\" %q", r.Header["Accept"])
}

func GetUniqueID() int64 {
	return time.Now().UnixNano()
}

//FIXME: add dashes for intermediate spaces
// Slugify returns a string that does not contain white characters,
// punctuation, all letters are lower case.
// It could contain `-` but not at the beginning or end of the text.
// TODO: It should be in range of the MaxLength var if specified
func Slugify(s string) string {
	var sb strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) {
			sb.WriteRune(unicode.ToLower(r))
		}
	}
	return sb.String()
}

func EnvOrDefault(key, alt string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return alt
}
