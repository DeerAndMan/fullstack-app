package errcode

import "net/http"

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	HTTP    int    `json:"-"`
}

func (e *Error) Error() string {
	return e.Message
}

func New(code int, msg string, httpStatus int) *Error {
	return &Error{Code: code, Message: msg, HTTP: httpStatus}
}

// Common errors
var (
	Success          = New(0, "success", http.StatusOK)
	ErrBadRequest    = New(400, "bad request", http.StatusBadRequest)
	ErrUnauthorized  = New(401, "unauthorized", http.StatusUnauthorized)
	ErrForbidden     = New(403, "forbidden", http.StatusForbidden)
	ErrNotFound      = New(404, "not found", http.StatusNotFound)
	ErrInternal      = New(500, "internal server error", http.StatusInternalServerError)
)

// Auth errors
var (
	ErrInvalidCredentials = New(10001, "invalid username or password", http.StatusUnauthorized)
	ErrTokenExpired       = New(10002, "token expired", http.StatusUnauthorized)
	ErrTokenInvalid       = New(10003, "invalid token", http.StatusUnauthorized)
	ErrUsernameExists     = New(10004, "username already exists", http.StatusBadRequest)
)

// User errors
var (
	ErrUserNotFound = New(20001, "user not found", http.StatusNotFound)
	ErrUserDisabled = New(20002, "user is disabled", http.StatusForbidden)
)

// Role errors
var (
	ErrRoleNotFound   = New(30001, "role not found", http.StatusNotFound)
	ErrRoleNameExists = New(30002, "role name already exists", http.StatusBadRequest)
	ErrRoleCodeExists = New(30003, "role code already exists", http.StatusBadRequest)
)

// Upload errors
var (
	ErrFileTooLarge  = New(40001, "file too large", http.StatusBadRequest)
	ErrFileTypeNotAllowed = New(40002, "file type not allowed", http.StatusBadRequest)
)
