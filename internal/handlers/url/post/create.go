package post

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
	Title    string `json:"title" validate:"required,min=3,max=255"`
	Content  string `json:"content" validate:"required,min=10"`
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
		log := log.With(slog.String("fn", "handlers.url.post.Create"))
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

		post := &models.Post{
			Title:    req.Title,
			Content:  req.Content,
			AuthorID: authorID,
		}

		postID, err := creator.CreatePost(post)
		if err != nil {
			if errors.Is(err, repository.ErrForeignKeyFailed) {
				log.Error("author does not exist", sl.Error(err))
				util.ErrorResponse(w, http.StatusBadRequest, "author does not exist")
				return
			}
			log.Error("failed creating post", sl.Error(err))
			util.ErrorResponse(w, http.StatusBadRequest,"internal server error")
			return
		}

		resp := CreateResponse{
			BaseResponse: response.BaseResponse{
				Status: http.StatusAccepted,
			},
			PostID: postID,
		}

		log.Info("created post", slog.Int64("post_id", postID), slog.String("title", post.Title))
		err = jsonutil.WriteJSON(w, http.StatusCreated, resp)
		if err != nil {
			log.Error("json writer error", sl.Error(err))
			http.Error(w, "Internal Server Error", 500)
			return
		}
	}
}
