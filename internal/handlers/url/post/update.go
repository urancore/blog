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
	"blog/internal/models"
	"blog/internal/repository"
	"blog/internal/util/logger"
	"blog/internal/util/logger/sl"
)

type UpdateRequest struct {
	Title    string `json:"title" validate:"required,min=3,max=255"`
	Content  string `json:"content" validate:"required,min=10"`
	AuthorID int64  `json:"author_id" validate:"required,min=1"`
}

type UpdateResponse struct {
	response.BaseResponse
	PostID int64 `json:"post_id,omitempty"`
}

type PostUpdater interface {
	UpdatePost(post *models.Post) error
}

func Update(log logger.Logger, postUpdater PostUpdater) http.HandlerFunc {
	validate := validator.New()
	return func(w http.ResponseWriter, r* http.Request) {
		log := log.With(slog.String("fn", "handlers.url.post.Update"))

		// TODO: check auth user, check user id its correct
		postID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			log.Info("invalid path value", sl.Error(err))
			jsonutil.WriteJSON(w, http.StatusBadRequest,
				response.Error(http.StatusBadRequest, "Bad Request"))
			return
		}
		// if user.id != author.id {its not your post} use jwt tocken

		var req UpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Info("error decoding request", sl.Error(err))
			// TODO: add error handler for jsonutil
			jsonutil.WriteJSON(w, http.StatusBadRequest,
						response.Error(http.StatusBadGateway, "EOF"))
			return
		}

		if err := validate.Struct(req); err != nil {
			validationErrors := err.(validator.ValidationErrors)
			log.Info("error validation request", sl.Error(validationErrors))
			jsonutil.WriteJSON(w, http.StatusBadRequest,
						response.Error(http.StatusBadRequest, "invalid json format"))
			return
		}

		newPost := &models.Post{
			ID: postID,
			Title: req.Title,
			Content: req.Title,
			AuthorID: req.AuthorID,
		}
		err = postUpdater.UpdatePost(newPost)
		if err != nil {
			if errors.Is(err, repository.ErrNotExists) {
				log.Info("post is not exists", sl.Error(err), slog.Int64("post_id", postID))
				jsonutil.WriteJSON(w, http.StatusNotFound,
							response.Error(http.StatusNotFound, "Not Found"))
				return
			}
			log.Error("error update post", sl.Error(err), slog.Int64("post_id", postID))
			jsonutil.WriteJSON(w, http.StatusInternalServerError,
						response.Error(http.StatusInternalServerError, "Internal Server Error"))
			return
		}

		resp := UpdateResponse{
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
