package post

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"blog/internal/api/jsonutil"
	"blog/internal/api/response"
	"blog/internal/models"
	"blog/internal/repository"
	"blog/internal/util"
	"blog/internal/util/logger"
	"blog/internal/util/logger/sl"
)

// TODO: reafactor struct

type readResponse struct {
	response.BaseResponse
	PostID   int64  `json:"post_id,omitempty"`
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

func Read(log logger.Logger, postReader postReader, userGetter userGetterReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := log.With("fn", "handlers.url.post.Read")
		postID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			log.Info("invalid path value", sl.Error(err))
			util.ErrorResponse(w, http.StatusBadRequest, "Bad Request")
			return
		}

		// TODO: add error handler for jsonutil
		post, err := postReader.GetPostByID(postID)
		if err != nil {
			if errors.Is(err, repository.ErrNotExists) {
				log.Info("post not found", sl.Error(err), slog.Int64("postID", postID))
				util.ErrorResponse(w, http.StatusBadRequest, "Page Not Found")
				return
			}
			log.Error("error get post by id", sl.Error(err), slog.Int64("postID", postID))
			util.ErrorResponse(w, http.StatusInternalServerError, "Internal Server Error")
			return
		}
		user, err := userGetter.GetUserByID(post.AuthorID)
		if err != nil {
			if errors.Is(err, repository.ErrNotExists) {
				log.Info("author not found",
					slog.Int64("user_id", post.AuthorID),
					sl.Error(err))
			}
			log.Error("error getting author",
				slog.Int64("user_id", post.AuthorID),
				sl.Error(err))
		}

		// TODO: ADD REDIS for cache
		resp := readResponse{
			BaseResponse: response.BaseResponse{
				Status: http.StatusOK,
			},
			PostID:   postID,
			Title:    post.Title,
			Content:  post.Content,
			AuthorID: post.AuthorID,
			Username: user.Username,
		}
		err = jsonutil.WriteJSON(w, http.StatusCreated, resp)
		if err != nil {
			log.Error("json writer error", sl.Error(err))
			http.Error(w, "Internal Server Error", 500)
			return
		}
	}
}
