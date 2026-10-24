package httpx

import "net/http"

// Error writes a plain-text error response with the given status code.
func Error(w http.ResponseWriter, message string, status int) {
	http.Error(w, message, status)
}

// BadRequest writes a 400 error.
func BadRequest(w http.ResponseWriter, message string) {
	Error(w, message, http.StatusBadRequest)
}

// InternalError writes a 500 error.
func InternalError(w http.ResponseWriter, message string) {
	Error(w, message, http.StatusInternalServerError)
}

// MethodNotAllowed writes a 405 error.
func MethodNotAllowed(w http.ResponseWriter) {
	Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}
