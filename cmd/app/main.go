package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"blog/internal/config"
	"blog/internal/handlers/url/comment"
	"blog/internal/handlers/url/post"
	"blog/internal/handlers/url/user"
	"blog/internal/middlewares/auth"
	logmd "blog/internal/middlewares/log_md"
	requestid "blog/internal/middlewares/request_id"
	"blog/internal/repository/redisrepo"
	"blog/internal/repository/sqliterepo"
	"blog/internal/util/logger"
	"blog/internal/util/logger/sl"
)

const (
	minSecretKeySize       = 32
	shutdownTimeoutSeconds = 5 // время на завершение работы сервера
)

func main() {
	log := logger.NewLogger(nil)

	log.Info("Application is starting...")

	log.Info("Loading configuration...")
	cfg, err := config.Load("config/local.yaml")
	if err != nil {
		log.Error("Failed to load configuration", sl.Error(err))
		os.Exit(1)
	}
	log.Info("Configuration loaded successfully.", slog.String("env", cfg.Environment))

	log = logger.NewLogger(cfg)
	log.Info("Logger reinitialized with configuration settings.")

	log.Debug("Checking SECRET_KEY environment variable...")
	if cfg.Auth.SecretKey == "" {
		log.Error("SECRET_KEY is not set in configuration or environment variable.")
		os.Exit(1)
	} else if len(cfg.Auth.SecretKey) < minSecretKeySize {
		log.Error(
			"Invalid SECRET_KEY length",
			slog.Int("required", minSecretKeySize),
			slog.Int("actual", len(cfg.Auth.SecretKey)),
		)
		os.Exit(1)
	}
	log.Info("SECRET_KEY checked successfully.")

	log.Info("Initializing SQLite database...", slog.String("path", cfg.SQLite.Path))
	db, err := sqliterepo.New(cfg.SQLite.Path, log)
	if err != nil {
		log.Error("SQLite database initialization failed", sl.Error(err))
		os.Exit(1)
	}
	log.Info("SQLite database connection established.")

	log.Debug("Initializing SQL repositories...")
	sqlRepo := sqliterepo.NewSQLiteRepository(log, db)
	postRepo := sqlRepo.Post()
	userRepo := sqlRepo.User()
	commentRepo := sqlRepo.Comment()
	log.Info("SQL repositories initialized.")

	var rdb *redisrepo.RedisRepo
	log.Info("Initializing Redis client...", slog.String("addr", cfg.Redis.Addr))
	rdb, err = redisrepo.NewRedisClient(cfg)
	if err != nil {
		log.Error("Failed to connect to Redis", sl.Error(err), slog.String("addr", cfg.Redis.Addr))
		rdb = nil
	} else {
		log.Info("Redis client connected successfully.")
	}

	log.Info("Initializing database tables...")
	if err := postRepo.InitPostDatabase(); err != nil {
		log.Error("Failed to initialize post table", sl.Error(err))
		os.Exit(1)
	}
	if err := userRepo.InitUserDatabase(); err != nil {
		log.Error("Failed to initialize user table", sl.Error(err))
		os.Exit(1)
	}
	if err := commentRepo.InitCommentDatabase(); err != nil {
		log.Warn("Failed to initialize comment table", sl.Error(err))
	}
	log.Info("Database tables initialized successfully.")

	log.Info("Registering HTTP routes...")
	mux := http.NewServeMux()

	// Post handlers
	mux.Handle("POST /api/post", auth.AuthMiddleware(post.Create(log, postRepo)))
	mux.Handle("PATCH api/post/{id}", auth.AuthMiddleware(post.Update(log, postRepo)))
	mux.Handle("DELETE /api/post/{id}", auth.AuthMiddleware(post.Delete(log, postRepo)))
	mux.HandleFunc("GET /api/post/{id}", post.Read(log, postRepo, userRepo, rdb))
	mux.HandleFunc("GET /api/posts", post.GetList(log, postRepo, userRepo))

	// User handlers
	mux.HandleFunc("POST /api/user/signup", user.SignUpHandler(log, userRepo))
	mux.HandleFunc("POST /api/user/signin", user.SignInHandler(log, userRepo))

	// Comment handlers
	mux.HandleFunc("POST /api/comment", auth.AuthMiddleware(comment.Create(log, commentRepo)))
	mux.HandleFunc("DELETE /api/comment/{id}", auth.AuthMiddleware(comment.Delete(log, commentRepo)))
	mux.HandleFunc("PATCH /api/comment/{id}", auth.AuthMiddleware(comment.Update(log, commentRepo)))
	mux.HandleFunc("GET /api/comments", comment.GetList(log, commentRepo, userRepo))

	log.Info("HTTP routes registered.")

	log.Info("Applying middlewares...")
	handler := requestid.RequestID(logmd.MiddlewareLogger(log)(mux))
	log.Info("Middlewares applied: RequestID, Logger.")

	serverAddr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:         serverAddr,
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		log.Info("HTTP server starting...",
			slog.String("address", server.Addr),
			slog.Duration("read_timeout", cfg.Server.ReadTimeout),
			slog.Duration("write_timeout", cfg.Server.WriteTimeout),
			slog.Duration("idle_timeout", cfg.Server.IdleTimeout),
		)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("HTTP server failed to start", sl.Error(err))
			os.Exit(1)
		}
		log.Info("HTTP server stopped gracefully.")
	}()

	// Setting Graceful Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM) // get Ctrl+C and SIGTERM

	sig := <-stop
	log.Info("Received shutdown signal", slog.String("signal", sig.String()))

	log.Info("Attempting graceful shutdown of HTTP server...")
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), shutdownTimeoutSeconds*time.Second)
	defer cancelShutdown()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("HTTP server graceful shutdown failed", sl.Error(err))
		os.Exit(1)
	}

	log.Info("HTTP server gracefully stopped.")
	log.Info("Application shutdown complete. Goodbye!")
	os.Exit(0)
}
