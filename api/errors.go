package api

import (
	"errors"
	"fmt"
	"log"
	"net/http"
)

var (
	ErrCouldNotRead             = errors.New("could not read")
	ErrCouldNotReadURLParameter = fmt.Errorf("%w a URL parameter", ErrCouldNotRead)
	ErrCouldNotReadCtxValue     = fmt.Errorf("%w a context value", ErrCouldNotRead)
	ErrCouldNotReadPathVar      = fmt.Errorf("%w a path variable", ErrCouldNotRead)

	ErrNotFound        = errors.New("not found")
	ErrUserNotFound    = fmt.Errorf("the user %w", ErrNotFound)
	ErrDrawingNotFound = fmt.Errorf("the drawing %w", ErrNotFound)
	ErrPointNotFound   = fmt.Errorf("the point %w", ErrNotFound)

	ErrAlreadyExist       = errors.New("already exist")
	ErrValueIsNotSettable = errors.New("the value is not settable")
	ErrWrongValueKind     = errors.New("wrong value kind")

	ErrBadRequestData     = errors.New("bad request data")
	ErrEmptyRequestBody   = fmt.Errorf("%w: empty request body", ErrBadRequestData)
	ErrBadLoginOrPassword = fmt.Errorf("%w: specified wrong or empty login/password", ErrBadRequestData)

	ErrOperationNotAllowed = fmt.Errorf("operation is not allowed")
)

// writeError writes error, if it's not equal nil, into http.ResponseWriter and log.Logger, and then returns true.
// Returns false if err equals nil.
func writeError(w http.ResponseWriter, err error) bool {
	multiTargetErrIs := func(err error, targets ...error) bool {
		for _, t := range targets {
			if errors.Is(err, t) {
				return true
			}
		}
		return false
	}
	respMsg, respStatus, printPanic := "", 0, false
	if err == nil {
		return false
	} else if multiTargetErrIs(err, ErrCouldNotReadPathVar, ErrCouldNotReadURLParameter, ErrAlreadyExist, ErrBadRequestData) {
		respStatus, respMsg = http.StatusBadRequest, "bad request: "+err.Error()
	} else if multiTargetErrIs(err, ErrNotFound) {
		respStatus, respMsg = http.StatusNotFound, "not found: "+err.Error()
	} else if multiTargetErrIs(err, ErrOperationNotAllowed) {
		respStatus, respMsg = http.StatusForbidden, "forbidden: "+err.Error()
	} else {
		printPanic, respStatus, respMsg = true, http.StatusInternalServerError, "internal server error"
	}
	http.Error(w, respMsg, respStatus)
	if printPanic {
		log.Panic(err)
	}
	return true
}
