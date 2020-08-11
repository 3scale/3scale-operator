package client

import (
	"fmt"
	"net/http"
)

type ApiErr struct {
	code int
	err  string
}

func (e ApiErr) Error() string {
	return fmt.Sprintf("error calling 3scale system - reason: %s - code: %d", e.err, e.code)
}

func (e ApiErr) Code() int {
	return e.code
}

// codeForError returns the HTTP status for a particular error.
func codeForError(err error) int {
	switch t := err.(type) {
	case ApiErr:
		return t.Code()
	}
	// Unknown
	return -1
}

// IsNotFound returns true if the specified error was created by NewNotFound.
func IsNotFound(err error) bool {
	return codeForError(err) == http.StatusNotFound
}

// IsBadRequest determines if err is an error which indicates that the request is invalid.
func IsBadRequest(err error) bool {
	return codeForError(err) == http.StatusBadRequest
}

// IsUnauthorized determines if err is an error which indicates that the request is unauthorized and
// requires authentication by the user.
func IsUnauthorized(err error) bool {
	return codeForError(err) == http.StatusUnauthorized
}

// IsForbidden determines if err is an error which indicates that the request is forbidden and cannot
// be completed as requested.
func IsForbidden(err error) bool {
	return codeForError(err) == http.StatusForbidden
}
