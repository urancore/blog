package requestid

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
)

type ctxKey string

const requestIDKey ctxKey = "reqID"

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ID (16 байт -> 32 hex символа)
		requestID := generateID()
		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		r = r.WithContext(ctx)
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r)
	})
}

func generateID() string {
	buf := make([]byte, 16) // 128 бит
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("[error] %v", err)
	}
	return hex.EncodeToString(buf)
}

func Get(ctx context.Context) string {
	if val, ok := ctx.Value(requestIDKey).(string); ok {
		return val
	}
	return ""
}
