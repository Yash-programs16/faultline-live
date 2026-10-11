package response

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
}

type SuccessResponse struct {
	Data interface{} `json:"data"`
	Meta interface{} `json:"meta,omitempty"`
}

func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func Success(w http.ResponseWriter, data interface{}, meta ...interface{}) {
	resp := SuccessResponse{Data: data}
	if len(meta) > 0 {
		resp.Meta = meta[0]
	}
	JSON(w, http.StatusOK, resp)
}

func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, SuccessResponse{Data: data})
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func Error(w http.ResponseWriter, status int, errMsg, code string) {
	JSON(w, status, ErrorResponse{Error: errMsg, Code: code})
}

func BadRequest(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, message, "BAD_REQUEST")
}

func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnauthorized, message, "UNAUTHORIZED")
}

func Forbidden(w http.ResponseWriter, message string) {
	Error(w, http.StatusForbidden, message, "FORBIDDEN")
}

func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, message, "NOT_FOUND")
}

func InternalError(w http.ResponseWriter, message string) {
	Error(w, http.StatusInternalServerError, message, "INTERNAL_ERROR")
}

func ServiceUnavailable(w http.ResponseWriter, message string) {
	Error(w, http.StatusServiceUnavailable, message, "SERVICE_UNAVAILABLE")
}

func TooManyRequests(w http.ResponseWriter, message string) {
	Error(w, http.StatusTooManyRequests, message, "TOO_MANY_REQUESTS")
}
