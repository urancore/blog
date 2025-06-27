package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"blog/internal/api/jsonutil"
	"blog/internal/api/response"
	"blog/internal/handlers/auth"
)

// 401 unauthorized

const (
	authHeader = "Authorization"
	UserIDCtxKey = "user_id"
)

func AuthMiddleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get(authHeader)
		if header == " " {
			jsonutil.WriteJSON(w, http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "auth header empty"))
			return
		}

		headerParts := strings.Split(header, " ")
		if len(headerParts) != 2 {
			jsonutil.WriteJSON(w, http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "invalid auth header"))
			return
		}

		userID, err := auth.ParseToken(headerParts[1])
		if err != nil {
			fmt.Println(err)
			jsonutil.WriteJSON(w, http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "invalid auth token"))
			return
		}

		userID = int64(userID)
		ctx := context.WithValue(r.Context(), UserIDCtxKey, userID)
		ctxR := r.WithContext(ctx)
		next.ServeHTTP(w, ctxR)
	}
}
