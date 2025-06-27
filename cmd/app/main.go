package main

import (
	"fmt"
	"net/http"
	"os"

	"blog/internal/config"
	"blog/internal/middlewares/auth"
	logmd "blog/internal/middlewares/log_md"
	requestid "blog/internal/middlewares/request_id"

	"blog/internal/handlers/url/post"
	"blog/internal/handlers/url/user"
	"blog/internal/repository/sqliterepo"
	"blog/internal/util/logger"
	"blog/internal/util/logger/sl"
)

const minSecretKeySize = 32

func main() {
	err := os.Setenv("SECRET_KEY", "01234567890123456789012345678901") // TODO: add to config
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	secretKey := os.Getenv("SECRET_KEY")
	if minSecretKeySize < len(secretKey) {
		fmt.Printf("SECRET_KEY must be at least %d chars\n", minSecretKeySize)
		os.Exit(1)
	}

	cfg, err := config.Load("config/local.yaml")
	if err != nil {
		fmt.Println(err)
	}

	log := logger.NewLogger(cfg)
	log.Info("logger inited")
	db, err := sqliterepo.New(cfg.SQLite.Path, log)
	if err != nil {
		log.Error("init database error", sl.Error(err))
		os.Exit(1)
	}

	repo := sqliterepo.NewSQLiteRepository(log, db)
	postRepo := repo.Post()
	userRepo := repo.User()
	postRepo.InitPostDatabase()
	userRepo.InitUserDatabase()

	mux := http.NewServeMux()

	// TODO: separate routers using a mux. (userMux, postMux, etc...)
	mux.Handle("POST /post/", auth.AuthMiddleware(post.Create(log, postRepo)))
	mux.Handle("PATCH /post/{id}", auth.AuthMiddleware(post.Update(log, postRepo)))
	mux.Handle("DELETE /post/{id}", auth.AuthMiddleware(post.Delete(log, postRepo)))
	mux.HandleFunc("GET /post/{id}", post.Read(log, postRepo, userRepo))
	mux.HandleFunc("GET /posts", post.GetList(log, postRepo, userRepo))

	// User
	mux.HandleFunc("POST /user/signup", user.SignUpHandler(log, userRepo))
	mux.HandleFunc("POST /user/signin", user.SignInHandler(log, userRepo))

	handle_mux := requestid.RequestID(logmd.MiddlewareLogger(log)(mux))

	server := &http.Server{
		Addr: fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler: handle_mux,
		ReadTimeout: cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout: cfg.Server.IdleTimeout,
	}

	server.ListenAndServe()
}
