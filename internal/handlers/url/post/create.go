package post

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"

	"blog/internal/api/jsonutil"
	"blog/internal/api/response"
	"blog/internal/models"
	"blog/internal/repository"
	"blog/internal/util/logger"
	"blog/internal/util/logger/sl"
)

type Request struct {
	Title    string `json:"title" validate:"required,min=3,max=255"`
	Content  string `json:"content" validate:"required,min=10"`
	AuthorID int64  `json:"author_id" validate:"required,min=1"`
}

type CreateResponse struct {
	response.BaseResponse
	PostID int64 `json:"post_id,omitempty"`
}

type PostCreator interface {
	CreatePost(post *models.Post) (int64, error) // post.title, post.content, post.authorID
}

func Create(log logger.Logger, creator PostCreator) http.HandlerFunc {
	validate := validator.New()

	return func(w http.ResponseWriter, r *http.Request) {
		log := log.With(slog.String("fn", "handlers.url.post.create.New"))
		if r.Method != http.MethodPost {

			return
		}

		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Error("error decoding request", sl.Error(err))
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

		req.AuthorID = int64(req.AuthorID)
		post := &models.Post{
			Title:    req.Title,
			Content:  req.Content,
			AuthorID: req.AuthorID,
		}

		postID, err := creator.CreatePost(post)
		if err != nil {
			if errors.Is(err, repository.ErrForeignKeyFailed) {
				log.Error("author does not exist", sl.Error(err))
				jsonutil.WriteJSON(w, http.StatusBadRequest,
							response.Error(http.StatusBadRequest, "author does not exist"))
				return
			}
			log.Error("failed creating post", sl.Error(err))
			jsonutil.WriteJSON(w, http.StatusBadRequest,
						response.Error(http.StatusInternalServerError, "internal server error"))
			return
		}

		resp := CreateResponse{
			BaseResponse: response.BaseResponse{
				Status: http.StatusAccepted,
			},
			PostID: postID,
		}

		err = jsonutil.WriteJSON(w, http.StatusCreated, resp)
		if err != nil {
			log.Error("json writer error", sl.Error(err))
			http.Error(w, "Internal Server Error", 500)
			return
		}
	}
}
