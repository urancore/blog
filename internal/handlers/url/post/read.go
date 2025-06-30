package post

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"blog/internal/api/jsonutil"
	"blog/internal/api/response"
	"blog/internal/models"
	"blog/internal/repository"
	"blog/internal/repository/redisrepo"
	"blog/internal/util"
	"blog/internal/util/logger"
	"blog/internal/util/logger/sl"
)

type readResponse struct {
	response.BaseResponse
	ID       int64  `json:"post_id,omitempty"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	AuthorID int64  `json:"author_id"`
	Username string `json:"username"`
}

type postReader interface {
	GetPostByID(id int64) (*models.Post, error)
}

type userGetterReader interface {
	GetUserByID(id int64) (*models.User, error)
}

func Read(log logger.Logger, postReader postReader, userGetter userGetterReader, rdb *redisrepo.RedisRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := log.With("fn", "handlers.url.post.Read")

		postID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			log.Info("invalid path value", sl.Error(err))
			util.ErrorResponse(w, http.StatusBadRequest, "Bad Request")
			return
		}

		postKey := fmt.Sprintf("post:%d", postID)
		var cachedPost redisrepo.PostModel

		err = rdb.Get(ctx, postKey, &cachedPost)
		if err == nil {
			log.Info("sending cached post", slog.Int64("postID", postID))
			resp := readResponse{
				BaseResponse: response.BaseResponse{Status: http.StatusOK},
				ID:           cachedPost.ID,
				Title:        cachedPost.Title,
				Content:      cachedPost.Content,
				AuthorID:     cachedPost.AuthorID,
				Username:     cachedPost.Username,
			}
			if writeErr := jsonutil.WriteJSON(w, http.StatusOK, resp); writeErr != nil {
				log.Error("json writer error (cached response)", sl.Error(writeErr))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		} else if !errors.Is(err, redisrepo.KeyNotFound) {
			log.Error("redis get error", sl.Error(err), slog.Int64("postID", postID))
		} else {
			log.Info("post not found in cache", slog.Int64("postID", postID))
		}

		post, err := postReader.GetPostByID(postID)
		if err != nil {
			if errors.Is(err, repository.ErrNotExists) {
				log.Info("post not found in DB", sl.Error(err), slog.Int64("postID", postID))
				util.ErrorResponse(w, http.StatusNotFound, "Page Not Found")
				return
			}
			log.Error("error getting post by id from DB", sl.Error(err), slog.Int64("postID", postID))
			util.ErrorResponse(w, http.StatusInternalServerError, "Internal Server Error")
			return
		}

		user, err := userGetter.GetUserByID(post.AuthorID)
		username := ""
		if err != nil {
			if errors.Is(err, repository.ErrNotExists) {
				log.Warn("author not found for post",
					slog.Int64("user_id", post.AuthorID),
					slog.Int64("post_id", post.ID),
					sl.Error(err))
			} else {
				log.Error("error getting author from DB",
					slog.Int64("user_id", post.AuthorID),
					slog.Int64("post_id", post.ID),
					sl.Error(err))
			}
		} else {
			username = user.Username
		}

		resp := readResponse{
			BaseResponse: response.BaseResponse{
				Status: http.StatusOK,
			},
			ID:       postID,
			Title:    post.Title,
			Content:  post.Content,
			AuthorID: post.AuthorID,
			Username: username,
		}

		postToCache := redisrepo.PostModel{
			Post: models.Post{
				ID:       post.ID,
				Title:    post.Title,
				Content:  post.Content,
				AuthorID: post.AuthorID,
			},
			Username: username,
		}

		go func() {
			setErr := rdb.Set(ctx, postKey, postToCache, 12*time.Hour)
			if setErr != nil {
				log.Error("failed to set post to redis cache", sl.Error(setErr), slog.Int64("postID", postID))
			} else {
				log.Info("post cached successfully", slog.Int64("postID", postID))
			}
		}()

		if writeErr := jsonutil.WriteJSON(w, http.StatusOK, resp); writeErr != nil {
			log.Error("json writer error (direct response)", sl.Error(writeErr))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}
