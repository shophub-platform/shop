package response

import (
	"encoding/json"
	"net/http"
)

// APIResponse je standardni wrapper za sve API odgovore.
// Kao ResponseEntity<T> u Spring-u.
type APIResponse[T any] struct {
	Success bool   `json:"success"`
	Data    T      `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

// JSON šalje JSON odgovor sa zadatim status kodom.
// Kao ResponseEntity.ok().body(...) u Spring-u.
func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Ok šalje 200 sa podacima.
func Ok[T any](w http.ResponseWriter, data T) {
	JSON(w, http.StatusOK, APIResponse[T]{Success: true, Data: data})
}

// Created šalje 201 sa podacima.
func Created[T any](w http.ResponseWriter, data T) {
	JSON(w, http.StatusCreated, APIResponse[T]{Success: true, Data: data})
}

// Error šalje grešku sa zadatim status kodom.
func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, APIResponse[any]{Success: false, Error: message})
}

// BadRequest šalje 400.
func BadRequest(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, message)
}

// NotFound šalje 404.
func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, message)
}

// InternalError šalje 500.
func InternalError(w http.ResponseWriter, message string) {
	Error(w, http.StatusInternalServerError, message)
}
