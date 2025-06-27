package util

import (
	"net/http"

	"blog/internal/api/jsonutil"
	"blog/internal/api/response"
)

func ErrorResponse(w http.ResponseWriter, status int, msg string) {
	jsonutil.WriteJSON(w, status,
		response.Error(status, msg))
}
