package comment

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"blog/internal/api/jsonutil"
	"blog/internal/api/response"
	"blog/internal/middlewares/auth"
	"blog/internal/models"
	"blog/internal/repository"
	"blog/internal/util"
	"blog/internal/util/logger"
	"blog/internal/util/logger/sl"

	"github.com/go-playground/validator/v10"
)

type updateRequest struct {
	PostID  int64  `json:"post_id" validate:"required"`
	Content string `json:"content" validate:"required,min=10,max=1024"`
}

type updateResponse struct {
	response.BaseResponse
	CommentID int64 `json:"comment_id,omitempty"`
	AuthorID  int64 `json:"author_id"`
	PostID    int64 `json:"post_id"`
}

type commentUpdater interface {
	UpdateComment(comment *models.Comment) error
	GetCommentAuthorID(commentID int64) (int64, error)
}

func Update(log logger.Logger, commentUpdater commentUpdater) http.HandlerFunc {
	validate := validator.New()
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.url.comment.Update"

		log := log.With(slog.String("fn", fn))

		authorID, ok := r.Context().Value(auth.UserIDCtxKey).(int64)
		if !ok {
			util.ErrorResponse(w, http.StatusInternalServerError, "Internal Server Error")
			return
		}

		commentID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			log.Info("invalid path value", sl.Error(err))
			util.ErrorResponse(w, http.StatusBadRequest, "Bad Request")
			return
		}

		var req updateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Info("error decoding request", sl.Error(err))
			util.ErrorResponse(w, http.StatusBadRequest, "EOF")
			return
		}

		if err := validate.Struct(req); err != nil {
			validationErrors := err.(validator.ValidationErrors)
			log.Info("error validation request", sl.Error(validationErrors))
			util.ErrorResponse(w, http.StatusBadRequest, "Invalid Json Format")
			return
		}

		commentAuthorID, err := commentUpdater.GetCommentAuthorID(commentID)
		if err != nil {
			if errors.Is(err, repository.ErrNotExists) {
				log.Info("comment author not found", sl.Error(err), slog.Int64("comment_id", commentID))
				util.ErrorResponse(w, http.StatusNotFound, "Not Found")
			}
			log.Error("error get author id", sl.Error(err))
			util.ErrorResponse(w, http.StatusInternalServerError, "Internal Server Error")
			return
		}

		if commentAuthorID != authorID {
			log.Info("forbidden: user is not the author",
				slog.Int64("comment_id", commentID),
				slog.Int64("comment_author_id", commentAuthorID),
				slog.Int64("author_id", authorID))

			util.ErrorResponse(w, http.StatusForbidden, "Forbidden")
			return
		}

		newComment := &models.Comment{
			ID:       commentID,
			Content:  req.Content,
			PostID:   req.PostID,
			AuthorID: commentAuthorID,
		}

		err = commentUpdater.UpdateComment(newComment)
		if err != nil {
			if errors.Is(err, repository.ErrNotExists) {
				log.Info("post is not exists", sl.Error(err),
					slog.Int64("post_id", req.PostID),
					slog.Int64("author_id", authorID),
					slog.Int64("comment_author_id", commentAuthorID))

				util.ErrorResponse(w, http.StatusNotFound, "Not Found")
				return
			}

			log.Error("error update post", sl.Error(err),
				slog.Int64("post_id", req.PostID),
				slog.Int64("author_id", authorID),
				slog.Int64("comment_author_id", commentAuthorID))

			util.ErrorResponse(w, http.StatusInternalServerError, "Internal Server Error")
			return
		}

		resp := updateResponse{
			BaseResponse: response.BaseResponse{
				Status: http.StatusAccepted,
			},
			CommentID: commentID,
			AuthorID:  authorID,
			PostID:    req.PostID,
		}

		log.Info("comment updated",
			slog.Int64("post_id", req.PostID),
			slog.Int64("comment_id", commentID))

		err = jsonutil.WriteJSON(w, http.StatusAccepted, resp)
		if err != nil {
			log.Error("json writer error", sl.Error(err))
			http.Error(w, "Internal Server Error", 500)
			return
		}
	}
}
