package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

//https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body

type malformedRequest struct {
	status int
	msg    string
}

func (mr *malformedRequest) Error() string {
	return mr.msg
}

// DecodeJSONBody decodes JSON request
// For server requests, the Request Body is always non-nil
// but will return EOF immediately when no body is present.
// The Server will close the request body. The ServeHTTP
// Handler does not need to.
func DecodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	// TODO: see godoc.header for info on how to extract all content-type strings
	// if r.Header.Get("Content-Type") != "" {
	// 	value, _ := r.Header., "Content-Type")
	// 	if value != "application/json" {
	// 		msg := "Content-Type header is not application/json"
	// 		return &malformedRequest{status: http.StatusUnsupportedMediaType, msg: msg}
	// 	}
	// }
	badRequest := func(msg string) error {
		return &malformedRequest{status: http.StatusBadRequest, msg: msg}
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1_048_576)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(&dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		switch {
		case errors.As(err, &syntaxError):
			return badRequest(fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset))
		case errors.Is(err, io.ErrUnexpectedEOF):
			return badRequest(fmt.Sprintf("Request body contains badly-formed JSON"))
		case errors.As(err, &unmarshalTypeError):
			return badRequest(fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset))
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return badRequest(fmt.Sprintf("Request body contains unknown field %s", fieldName))
		case errors.Is(err, io.EOF):
			return badRequest("Request body must not be empty")
		case err.Error() == "http: request body too large":
			return badRequest("Request body must not be larger than 1MB")
		default:
			return err
		}
	}
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return badRequest("Request body must only contain a single JSON object")
	}
	return nil
}

func JSON(w http.ResponseWriter, status int, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

// Map for adhoc json generation
//JSON(w, s, Map{
//     "field1": someStructData1,
//     "field2": someStructData2,
// })
type Map map[string]interface{}
