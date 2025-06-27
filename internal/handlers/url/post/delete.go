package post

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"blog/internal/api/jsonutil"
	"blog/internal/api/response"
	"blog/internal/middlewares/auth"
	"blog/internal/models"
	"blog/internal/repository"
	"blog/internal/util/logger"
	"blog/internal/util/logger/sl"
)

type DeleteResponse struct {
	response.BaseResponse
	PostID int64 `json:"post_id,omitempty"`
}

type PostDeleter interface {
	GetPostByID(id int64) (*models.Post, error)
	DeletePost(id int64) error
}

func Delete(log logger.Logger, postDeleter PostDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := log.With("fn", "handlers.url.post.Delete")

		authorID, ok := r.Context().Value(auth.UserIDCtxKey).(int64)
		if !ok {
			jsonutil.WriteJSON(w, http.StatusInternalServerError,
				response.Error(http.StatusInternalServerError, "Internal Server Error"))
			return
		}

		postID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			log.Info("invalid path value", sl.Error(err))
			jsonutil.WriteJSON(w, http.StatusBadRequest,
				response.Error(http.StatusBadRequest, "Bad Request"))
			return
		}

		post, err := postDeleter.GetPostByID(postID)
		if err != nil {
			if errors.Is(err, repository.ErrNotExists) {
				log.Info("post not found", slog.Int64("post_id", postID), sl.Error(err))
				jsonutil.WriteJSON(w, http.StatusNotFound,
					response.Error(http.StatusNotFound, "Post not found"))
				return
			}
			log.Info("post not found",
				slog.Int64("post_id", postID),
				sl.Error(err))

			jsonutil.WriteJSON(w, http.StatusInternalServerError,
				response.Error(http.StatusInternalServerError, "Internal Server Error"))
			return
		}

		postAuthorID := post.AuthorID
		if postAuthorID != authorID {
			log.Info("forbidden: user is not the author",
				slog.Int64("post_id", postID),
				slog.Int64("post_author_id", post.AuthorID),
				slog.Int64("author_id", authorID))

			jsonutil.WriteJSON(w, http.StatusForbidden,
				response.Error(http.StatusForbidden, "Forbidden"))
			return
		}

		err = postDeleter.DeletePost(postID)
		if err != nil {
			if errors.Is(err, repository.ErrNotExists) {
				log.Info("post is not exists", slog.Int64("post_id", postID))
				jsonutil.WriteJSON(w, http.StatusNotFound,
					response.Error(http.StatusNotFound, "Not Found"))
				return
			}
			log.Error("post delete error", slog.Int64("post_id", postID), sl.Error(err))
			jsonutil.WriteJSON(w, http.StatusInternalServerError,
				response.Error(http.StatusInternalServerError, "Internal Server Error"))
			return
		}

		resp := DeleteResponse{
			BaseResponse: response.BaseResponse{
				Status: http.StatusOK,
			},
			PostID: postID,
		}

		log.Info("post deleted", slog.Int64("post_id", postID))
		err = jsonutil.WriteJSON(w, http.StatusAccepted, resp)
		if err != nil {
			log.Error("json writer error", sl.Error(err))
			http.Error(w, "Internal Server Error", 500)
			return
		}
	}
}
