package comment

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

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

type CreateRequest struct {
	Content string `json:"content" validate:"required,min=10,max=1024"`
	PostID  int64  `json:"post_id"`
}

type CreateResponse struct {
	response.BaseResponse
	CommentID int64 `json:"comment_id,omitempty"`
	PostID    int64 `json:"post_id,omitempty"`
	AuthorID  int64 `json:"author_id,omitempty"`
}

type commentCreator interface {
	CreateComment(comment *models.Comment) (int64, error)
}

func Create(log logger.Logger, commentCreator commentCreator) http.HandlerFunc {
	validate := validator.New()
	return func(w http.ResponseWriter, r *http.Request) {
		log := log.With("fn", "handlers.url.comment.Create")
		if r.Method != http.MethodPost {
			util.ErrorResponse(w, http.StatusMethodNotAllowed, "Method Not Allowed")
			return
		}

		authorID, ok := r.Context().Value(auth.UserIDCtxKey).(int64)
		if !ok {
			util.ErrorResponse(w, http.StatusInternalServerError, "Internal Server Error")
			return
		}

		var req CreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Info("error decoding request", sl.Error(err))
			util.ErrorResponse(w, http.StatusBadRequest, "EOF")
			return
		}

		if err := validate.Struct(req); err != nil {
			validationErrors := err.(validator.ValidationErrors)
			log.Info("error validation request", sl.Error(validationErrors))
			util.ErrorResponse(w, http.StatusBadRequest, "invalid json format")
			return
		}

		comment := &models.Comment{
			Content:  req.Content,
			PostID:   req.PostID,
			AuthorID: authorID,
		}

		commentID, err := commentCreator.CreateComment(comment)
		if err != nil {
			if errors.Is(err, repository.ErrForeignKeyFailed) {
				log.Error("author or post does not exist", sl.Error(err))
				util.ErrorResponse(w, http.StatusBadRequest, "author or post does not exist")
				return
			}
			log.Error("failed creating comment", sl.Error(err))
			util.ErrorResponse(w, http.StatusBadRequest, "internal server error")
			return
		}

		resp := CreateResponse{
			BaseResponse: response.BaseResponse{
				Status: http.StatusAccepted,
			},
			CommentID: commentID,
			PostID:    req.PostID,
			AuthorID:  authorID,
		}

		log.Info("created comment", slog.Int64("post_id", req.PostID))
		err = jsonutil.WriteJSON(w, http.StatusCreated, resp)
		if err != nil {
			log.Error("json writer error", sl.Error(err))
			http.Error(w, "Internal Server Error", 500)
			return
		}
	}
}
