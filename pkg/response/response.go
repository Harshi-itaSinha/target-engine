package response

import (
	"encoding/json"
	"net/http"

	model "github.com/Harshi-itaSinha/target-engine/internal/models"
)

func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

func Success(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, data)
}

func NoContent(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

func BadRequest(w http.ResponseWriter, message string) {
	JSON(w, http.StatusBadRequest, &model.ErrorResponse{
		Error:   "Bad Request",
		Message: message,
		Code:    http.StatusBadRequest,
	})
}

func InternalServerError(w http.ResponseWriter, message string) {
	JSON(w, http.StatusInternalServerError, &model.ErrorResponse{
		Error:   "Internal Server Error",
		Message: message,
		Code:    http.StatusInternalServerError,
	})
}

func NotFound(w http.ResponseWriter, message string) {
	JSON(w, http.StatusNotFound, &model.ErrorResponse{
		Error:   "Not Found",
		Message: message,
		Code:    http.StatusNotFound,
	})
}

func TooManyRequests(w http.ResponseWriter, message string) {
	JSON(w, http.StatusTooManyRequests, &model.ErrorResponse{
		Error:   "Too Many Requests",
		Message: message,
		Code:    http.StatusTooManyRequests,
	})
}

func Unauthorized(w http.ResponseWriter, message string) {
	JSON(w, http.StatusUnauthorized, &model.ErrorResponse{
		Error:   "Unauthorized",
		Message: message,
		Code:    http.StatusUnauthorized,
	})
}

func Forbidden(w http.ResponseWriter, message string) {
	JSON(w, http.StatusForbidden, &model.ErrorResponse{
		Error:   "Forbidden",
		Message: message,
		Code:    http.StatusForbidden,
	})
}
