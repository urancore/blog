package logmd

import (
	"log/slog"
	"net/http"
	"time"

	requestid "blog/internal/middlewares/request_id"
	"blog/internal/util/logger"
)

type ResponseRecorder struct {
	http.ResponseWriter
	StatusCode int
	BodySize   int
}

func (r *ResponseRecorder) WriteHeader(statusCode int) {
	r.StatusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *ResponseRecorder) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.BodySize += size
	return size, err
}

func MiddlewareLogger(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := log.With(slog.String("fn", "middlewares.log_md.MiddlewareLogger"))
			start := time.Now()
			reqID := requestid.Get(r.Context())
			log.Info("====== request started ======")
			rRef := r.Referer()
			if rRef == "" {
				rRef = "none"
			}
			// TODO: add ip addr
			log = log.With("handling request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("request_id", reqID),
				slog.String("referer", rRef),
				slog.String("user_agent", r.UserAgent()),
				slog.String("protocol", r.Proto),
				slog.String("ip", r.RemoteAddr), // FIXME: reqrite to real ip
				slog.String("host", r.Host),
			)

			wRecorder := &ResponseRecorder{
				ResponseWriter: w,
				StatusCode:     http.StatusOK, // default
			}
			defer func() {
				duration := time.Since(start)
				log = log.With(slog.Int("status", wRecorder.StatusCode),
					slog.Int("response_size", wRecorder.BodySize),
					slog.Duration("duration", duration),
					slog.String("duaration_human", duration.String()),
				)
				log.Info("====== request completed ======")
			}()
			next.ServeHTTP(wRecorder, r)
		})
	}
}
