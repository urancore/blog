package post

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"

	"blog/internal/api/jsonutil"
	"blog/internal/api/response"
	"blog/internal/middlewares/auth"
	"blog/internal/models"
	"blog/internal/repository"
	"blog/internal/util"
	"blog/internal/util/logger"
	"blog/internal/util/logger/sl"
)

type updateRequest struct {
	Title   string `json:"title" validate:"required,min=3,max=255"`
	Content string `json:"content" validate:"required,min=10"`
}

type updateResponse struct {
	response.BaseResponse
	PostID int64 `json:"post_id,omitempty"`
}

type postUpdater interface {
	GetPostByID(id int64) (*models.Post, error)
	UpdatePost(post *models.Post) error
}

func Update(log logger.Logger, postUpdater postUpdater) http.HandlerFunc {
	validate := validator.New()
	return func(w http.ResponseWriter, r *http.Request) {
		log := log.With(slog.String("fn", "handlers.url.post.Update"))

		authorID, ok := r.Context().Value(auth.UserIDCtxKey).(int64)
		if !ok {
			util.ErrorResponse(w, http.StatusInternalServerError, "Internal Server Error")
			return
		}

		postID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			log.Info("invalid path value", sl.Error(err))
			util.ErrorResponse(w, http.StatusBadRequest, "Bad Request")
			return
		}

		var req updateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Info("error decoding request", sl.Error(err))
			// TODO: add error handler for jsonutil
			util.ErrorResponse(w, http.StatusBadRequest, "EOF")
			return
		}

		if err := validate.Struct(req); err != nil {
			validationErrors := err.(validator.ValidationErrors)
			log.Info("error validation request", sl.Error(validationErrors))
			util.ErrorResponse(w, http.StatusBadRequest, "Invalid Json Format")
			return
		}
		post, err := postUpdater.GetPostByID(postID)
		if err != nil {
			if errors.Is(err, repository.ErrNotExists) {
				log.Info("post is not exists", sl.Error(err), slog.Int64("post_id", postID))
				util.ErrorResponse(w, http.StatusNotFound, "Not Found")
				return
			}
		}

		if post.AuthorID != authorID {
			log.Info("forbidden: user is not the author",
				slog.Int64("post_id", postID),
				slog.Int64("post_author_id", post.AuthorID),
				slog.Int64("author_id", authorID))

			util.ErrorResponse(w, http.StatusForbidden, "Forbidden")
			return
		}

		newPost := &models.Post{
			ID:       postID,
			Title:    req.Title,
			Content:  req.Title,
			AuthorID: authorID,
		}

		err = postUpdater.UpdatePost(newPost)
		if err != nil {
			if errors.Is(err, repository.ErrNotExists) {
				log.Info("post is not exists", sl.Error(err), slog.Int64("post_id", postID))
				util.ErrorResponse(w, http.StatusNotFound, "Not Found")
				return
			}
			log.Error("error update post", sl.Error(err), slog.Int64("post_id", postID))
			util.ErrorResponse(w, http.StatusInternalServerError, "Internal Server Error")
			return
		}

		resp := updateResponse{
			BaseResponse: response.BaseResponse{
				Status: http.StatusAccepted,
			},
			PostID: postID,
		}
		log.Info("post updated", slog.Int64("post_id", postID))
		err = jsonutil.WriteJSON(w, http.StatusAccepted, resp)
		if err != nil {
			log.Error("json writer error", sl.Error(err))
			http.Error(w, "Internal Server Error", 500)
			return
		}
	}
}
