package main

import (
	"fmt"
	"net/http"
	"os"

	"blog/internal/config"
	// "blog/internal/models"
	"blog/internal/handlers/url/post"
	"blog/internal/repository/sqliterepo"
	"blog/internal/util/logger"
	"blog/internal/util/logger/sl"
	// "blog/internal/util/logger/sl"
)

func main() {
	cfg, err := config.Load("config/local.yaml")
	if err != nil {
		fmt.Println(err)
	}

	_ = cfg

	log := logger.NewLogger(cfg)
	log.Info("logger inited")
	db, err := sqliterepo.New(cfg.SQLite.Path, log)
	if err != nil {
		log.Error("init database error", sl.Error(err))
		os.Exit(1)
	}

	repo := sqliterepo.NewSQLiteRepository(log, db)
	postRepo := repo.Post()

	mux := http.NewServeMux()

	// TODO: write CRUD api
	mux.HandleFunc("POST /post/create", post.Create(log, postRepo)) // TODO: add auth
	mux.HandleFunc("GET /post/{id}", post.Read(log, postRepo))
	mux.HandleFunc("PATCH /post/update/{id}", post.Update(log, postRepo))
	mux.HandleFunc("DELETE /post/delete/{id}", post.Delete(log, postRepo))

	server := &http.Server{
		Addr: fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler: mux,
		ReadTimeout: cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout: cfg.Server.IdleTimeout,
	}

	server.ListenAndServe()
}
