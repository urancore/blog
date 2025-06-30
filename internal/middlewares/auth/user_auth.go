package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"blog/internal/handlers/auth"
	"blog/internal/util"
)

const (
	authHeader   = "Authorization"
	UserIDCtxKey = "user_id"
)

func AuthMiddleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get(authHeader)
		if header == " " {
			util.ErrorResponse(w, http.StatusUnauthorized, "Auth Header Empty")
			return
		}

		headerParts := strings.Split(header, " ")
		if len(headerParts) != 2 {
			util.ErrorResponse(w, http.StatusUnauthorized, "Invalid Auth Header")
			return
		}

		userID, err := auth.ParseToken(headerParts[1])
		if err != nil {
			fmt.Println(err)
			util.ErrorResponse(w, http.StatusUnauthorized, "Invalid Auth Token")
			return
		}

		userID = int64(userID)
		ctx := context.WithValue(r.Context(), UserIDCtxKey, userID)
		ctxR := r.WithContext(ctx)
		next.ServeHTTP(w, ctxR)
	}
}
