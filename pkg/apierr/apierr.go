package apierr

import (
	"encoding/json"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Body is the JSON error envelope used by HTTP handlers.
type Body struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Reason  string `json:"reason,omitempty"`
}

// Write writes a JSON error response with the given HTTP status.
func Write(w http.ResponseWriter, httpStatus int, reason, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(Body{
		Code:    httpStatus,
		Message: message,
		Reason:  reason,
	})
}

// Unauthorized writes 401 with reason (e.g. NO_SESSION, SESSION_EXPIRED).
func Unauthorized(w http.ResponseWriter, reason string) {
	if reason == "" {
		reason = "UNAUTHORIZED"
	}
	Write(w, http.StatusUnauthorized, reason, "unauthorized")
}

// BadRequest writes 400.
func BadRequest(w http.ResponseWriter, reason, message string) {
	if message == "" {
		message = "bad request"
	}
	Write(w, http.StatusBadRequest, reason, message)
}

// NotFound writes 404.
func NotFound(w http.ResponseWriter, reason, message string) {
	if message == "" {
		message = "not found"
	}
	Write(w, http.StatusNotFound, reason, message)
}

// Internal writes 500.
func Internal(w http.ResponseWriter, reason, message string) {
	if message == "" {
		message = "internal error"
	}
	Write(w, http.StatusInternalServerError, reason, message)
}

// GRPC helpers — thin wrappers so services share message style.

func Unauthenticated(msg string) error {
	if msg == "" {
		msg = "not authenticated"
	}
	return status.Error(codes.Unauthenticated, msg)
}

func InvalidArgument(msg string) error {
	return status.Error(codes.InvalidArgument, msg)
}

func NotFoundGRPC(msg string) error {
	if msg == "" {
		msg = "not found"
	}
	return status.Error(codes.NotFound, msg)
}

func AlreadyExists(msg string) error {
	return status.Error(codes.AlreadyExists, msg)
}

func InternalGRPC(msg string) error {
	if msg == "" {
		msg = "internal error"
	}
	return status.Error(codes.Internal, msg)
}
