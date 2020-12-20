package errors

// inspired by https://commandcenter.blogspot.com/2017/12/error-handling-in-upspin.html

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"
)

type (
	UserID string
	// Op describes an operation, usually as the package and method,
	// such as "user.Lookup".
	Op   string
	Kind int
)

// Diag holds diagnostic info
type Diag struct {
	// Path is the path name of the .
	Path   string `json:"path,omitempty"`
	Method string
	// Op is the operation being performed, usually the name of the method
	// being invoked (Get, Put, etc.). It should not contain an at sign @.
	Op Op `json:"op,omitempty"`
}

func D(r *http.Request, op string) Diag {
	return Diag{
		Path:   r.URL.Path,
		Method: r.Method,
		Op:     Op(op),
	}
}

var std = os.Stderr

// JSON encodes v as JSON and write it to http.ResponseWriter
func JSON(w http.ResponseWriter, status int, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

// Send responds with an error
// TODO: hide internal error details
func Send(w http.ResponseWriter, err error) {
	e, ok := err.(*Error)
	if !ok {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	JSON(w, e.Status, e)
	// for the moment,log the error here
	_ = Debug && Logln(e.Error())
}

// Error is the type that implements the error interface.
// It contains a number of fields, each of different type.
// An Error value may leave some values unset.
type Error struct {
	Diag `json:"diag,omitempty"`
	// User is the ID of the user attempting the operation.
	User UserID `json:"user,omitempty"`
	// Kind is the class of error, such as permission failure,
	// or "Other" if its class is unknown or irrelevant.
	Kind Kind `json:"kind,omitempty"`
	// additional details
	Detail string `json:"detail,omitempty"`
	// The underlying error that triggered this one, if any.
	Err error `json:"error,omitempty"`
	// Status is HTTP status codes
	Status int `json:"status,omitempty"`
}

func (e *Error) isZero() bool {
	return e.Path == "" && e.User == "" && e.Op == "" && e.Kind == 0 && e.Err == nil
}

// Error implements the error interface
func (e *Error) Error() string {
	b := new(bytes.Buffer)
	// e.printStack(b)
	if e.Op != "" {
		pad(b, ": ")
		b.WriteString(string(e.Op))
	}
	if e.Path != "" {
		pad(b, ": ")
		b.WriteString(string(e.Path))
	}
	if e.User != "" {
		if e.Path == "" {
			pad(b, ": ")
		} else {
			pad(b, ", ")
		}
		b.WriteString("user ")
		b.WriteString(string(e.User))
	}
	// if e.Kind != 0 {
	// 	pad(b, ": ")
	// 	// b.WriteString(e.Kind.String())
	// }
	if e.Status != 0 {
		pad(b, ": ")
		b.WriteString(http.StatusText(e.Status))
	}
	if e.Err != nil {
		// Indent on new line if we are cascading non-empty errors.
		if prevErr, ok := e.Err.(*Error); ok {
			if !prevErr.isZero() {
				pad(b, Separator)
				b.WriteString(e.Err.Error())
			}
		} else {
			pad(b, ": ")
			b.WriteString(e.Err.Error())
		}
	}
	if e.Detail != "" {
		pad(b, ": ")
		b.WriteString(e.Detail)
	}
	if b.Len() == 0 {
		return "no error"
	}
	return b.String()
}

// MarshalJSON implements MarshalJSON interface to serialize the Err field as string and add a time stamp
func (e *Error) MarshalJSON() ([]byte, error) {
	type alias Error // alias to avoid recursive call of MarshalJSON
	type augmented struct {
		alias
		Err       string    `json:"error"`
		TimeStamp time.Time `json:"time_stamp"`
	}
	return json.Marshal(&augmented{
		alias(*e),      //marshall the Error struct
		e.getErrText(), // overrides the e.Err field
		time.Now(),     // append a time stamp
	})
}

func (e *Error) getErrText() string {
	txt := e.Detail
	if txt == "" {
		txt = http.StatusText(e.Status)
	}
	if txt != "" {
		txt = txt + ", "
	}
	if e.Err != nil {
		txt = txt + e.Err.Error()
	}
	return txt
}

// Separator is the string used to separate nested errors.
var Separator = ":\n\t"

// E builds an error value from its arguments.
// There must be at least one argument or E panics.
// The type of each argument determines its meaning.
// If more than one argument of a given type is presented,
// only the last one is recorded.
// If the error is printed, only those items that have been
// set to non-zero values will appear in the result.
//
// If Kind is not specified or Other, we set it to the Kind of
// the underlying error.
//
func E(args ...interface{}) error {
	if len(args) == 0 {
		panic("call to errors.E with no arguments")
	}
	e := &Error{}
	for _, arg := range args {
		switch arg := arg.(type) {
		case Diag:
			e.Diag = arg
		case UserID:
			e.User = arg
		case int:
			// must be status
			e.Status = arg
		case string:
			//must be detail
			e.Detail = arg
		case Kind:
			e.Kind = arg
		case *Error:
			// Make a copy
			copy := *arg
			e.Err = &copy
		case error:
			e.Err = arg
		default:
			_, file, line, _ := runtime.Caller(1)
			_ = Debug && Logf("errors.E: bad call from %s:%d: %v", file, line, args)
			return Errorf("unknown type %T, value %v in error call", arg, arg)
		}
	}
	return e
}

// pad appends str to the buffer if the buffer already has some data.
func pad(b *bytes.Buffer, str string) {
	if b.Len() == 0 {
		return
	}
	b.WriteString(str)
}

// Recreate the errors.New functionality of the standard Go errors package
// so we can create simple text errors when needed.

// Str returns an error that formats as the given text. It is intended to
// be used as the error-typed argument to the E function.
func Str(text string) error {
	return &errorString{text}
}

// errorString is a trivial implementation of error.
type errorString struct {
	s string
}

func (e *errorString) Error() string {
	return e.s
}

// Errorf is equivalent to fmt.Errorf, but allows clients to import only this
// package for all error handling.
func Errorf(format string, args ...interface{}) error {
	return &errorString{fmt.Sprintf(format, args...)}
}

// Match compares its two error arguments. It can be used to check
// for expected errors in tests. Both arguments must have underlying
// type *Error or Match will return false. Otherwise it returns true
// iff every non-zero element of the first error is equal to the
// corresponding element of the second.
// If the Err field is a *Error, Match recurs on that field;
// otherwise it compares the strings returned by the Error methods.
// Elements that are in the second argument but not present in
// the first are ignored.
//
// // For example,
// //	Match(errors.E(UserName("joe@schmoe.com"), errors.Permission), err)
// // tests whether err is an Error with Kind=Permission and User=joe@schmoe.com.
// func Match(err1, err2 error) bool {
// 	e1, ok := err1.(*Error)
// 	if !ok {
// 		return false
// 	}
// 	e2, ok := err2.(*Error)
// 	if !ok {
// 		return false
// 	}
// 	if e1.Path != "" && e2.Path != e1.Path {
// 		return false
// 	}
// 	if e1.User != "" && e2.User != e1.User {
// 		return false
// 	}
// 	if e1.Op != "" && e2.Op != e1.Op {
// 		return false
// 	}
// 	// if e1.Kind != Other && e2.Kind != e1.Kind {
// 	// 	return false
// 	// }
// 	if e1.Err != nil {
// 		if _, ok := e1.Err.(*Error); ok {
// 			return Match(e1.Err, e2.Err)
// 		}
// 		if e2.Err == nil || e2.Err.Error() != e1.Err.Error() {
// 			return false
// 		}
// 	}
// 	return true
// }

// Is reports whether err is an *Error of the given Kind.
// If err is nil then Is returns false.
func Is(kind Kind, err error) bool {
	e, ok := err.(*Error)
	if !ok {
		return false
	}
	// if e.Kind != Other {
	// 	return e.Kind == kind
	// }
	if e.Err != nil {
		return Is(kind, e.Err)
	}
	return false
}

// Recover recovers from panic and send an internalservererror to client
func Recover(w http.ResponseWriter) {
	err := recover()
	if err != nil {
		_ = Debug && Logln(err)
		//TODO: sendMeMail(err)
		http.Error(w, "", http.StatusInternalServerError)
	}
}

// Fatal panics if err != nil
//FIXME: add more details
func Fatal(err error) {
	if err != nil {
		panic(err)
	}
}
