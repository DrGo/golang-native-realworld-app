package utils

import "net/http"

// CORSHandled sets CORS headers and returns true if it handled the (OPTIONS) request
func CORSHandled(w http.ResponseWriter, r *http.Request) bool {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
	if r.Method == "OPTIONS" {
		return true
	}
	return false
}
