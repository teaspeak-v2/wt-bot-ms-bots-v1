package apperror

import (
	"encoding/json"
	"errors"
	"net/http"
)

type Code string

const (
	CodeInvalidRequest Code = "invalid_request"
	CodeUnauthorized   Code = "unauthorized"
	CodeForbidden      Code = "forbidden"
	CodeNotFound       Code = "not_found"
	CodeConflict       Code = "conflict"
	CodeInternal       Code = "internal"
)

type Error struct {
	Code    Code
	Message string
	Status  int
	Err     error
}

func (e *Error) Error() string { return e.Message }
func (e *Error) Unwrap() error { return e.Err }

func New(code Code, status int, message string) *Error {
	return &Error{Code: code, Status: status, Message: message}
}

func Wrap(code Code, status int, message string, err error) *Error {
	return &Error{Code: code, Status: status, Message: message, Err: err}
}

func InvalidRequest(message string, err error) *Error {
	return Wrap(CodeInvalidRequest, http.StatusBadRequest, message, err)
}
func Unauthorized(message string, err error) *Error {
	return Wrap(CodeUnauthorized, http.StatusUnauthorized, message, err)
}
func Forbidden(message string, err error) *Error {
	return Wrap(CodeForbidden, http.StatusForbidden, message, err)
}
func NotFound(message string, err error) *Error {
	return Wrap(CodeNotFound, http.StatusNotFound, message, err)
}
func Conflict(message string, err error) *Error {
	return Wrap(CodeConflict, http.StatusConflict, message, err)
}
func Internal(message string, err error) *Error {
	return Wrap(CodeInternal, http.StatusInternalServerError, message, err)
}

func As(err error) *Error {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr
	}
	return Internal("internal server error", err)
}

type response struct {
	Error struct {
		Code    Code   `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func WriteJSON(w http.ResponseWriter, err error) {
	appErr := As(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.Status)
	_ = json.NewEncoder(w).Encode(response{Error: struct {
		Code    Code   `json:"code"`
		Message string `json:"message"`
	}{Code: appErr.Code, Message: appErr.Message}})
}

func StatusCode(err error) int { return As(err).Status }
func CodeOf(err error) Code    { return As(err).Code }
