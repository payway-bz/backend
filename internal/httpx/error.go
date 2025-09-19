package httpx

import (
	"net/http"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

// WriteErr writes a JSON error response with the provided status code and message.
func WriteErr(w http.ResponseWriter, code int, msg string) {
	WriteJSON(w, code, ErrorResponse{Error: msg})
}

// WriteInternalServerError logs to error.log and responds with HTTP 500 using a standard message.
func WriteInternalServerError(w http.ResponseWriter) {
	WriteErr(w, http.StatusInternalServerError, "internal server error")
}

// WriteForbidden logs to error.log and responds with HTTP 403.
func WriteForbidden(w http.ResponseWriter) {
	WriteErr(w, http.StatusForbidden, "forbidden")
}

// WriteBadRequest logs to error.log and responds with HTTP 400.
func WriteBadRequest(w http.ResponseWriter) {
	WriteErr(w, http.StatusBadRequest, "bad request")
}

func WriteUnauthorized(w http.ResponseWriter) {
	WriteErr(w, http.StatusUnauthorized, "unauthorized")
}
