package utils

import (
	"net/http"
)

type ResponseWrapper struct {
	http.ResponseWriter               // Struct embedding promotes all native methods
	StatusCode          int           // Captures the HTTP status code
}

func NewResponseWrapper(w http.ResponseWriter) *ResponseWrapper {
	return &ResponseWrapper{
		ResponseWriter: w,
		StatusCode:     http.StatusOK, // Default status if WriteHeader is never called
	}
}

// WriteHeader intercepts and records the HTTP status code.
func (rw *ResponseWrapper) WriteHeader(statusCode int) {
	rw.StatusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode) // Forward to actual client
}