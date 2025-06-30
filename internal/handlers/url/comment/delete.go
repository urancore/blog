package comment

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"blog/internal/api/jsonutil"
	"blog/internal/api/response"
	"blog/internal/middlewares/auth"
	"blog/internal/repository"
	"blog/internal/util"
	"blog/internal/util/logger"
	"blog/internal/util/logger/sl"
)

type deleteResponse struct {
	response.BaseResponse
	CommentID int64 `json:"comment_id,omitempty"`
	AuthorID  int64 `json:"author_id,omitempty"`
}

type commentDeleter interface {
	DeleteComment(id int64) error
	GetCommentAuthorID(commentID int64) (int64, error)
}

func Delete(log logger.Logger, commentDeleter commentDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := log.With("fn", "handlers.url.post.Delete")

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

		commentAuthorID, err := commentDeleter.GetCommentAuthorID(commentID)
		if err != nil {
			if errors.Is(err, repository.ErrNotExists) {
				log.Info("comment is not exists", slog.Int64("comment_id", commentID))
				util.ErrorResponse(w, http.StatusNotFound, "Not Found")
				return
			}
			log.Error("comment delete error", slog.Int64("comment_id", commentID), sl.Error(err))
			util.ErrorResponse(w, http.StatusInternalServerError, "Internal Server Error")
			return

		} else if commentAuthorID != authorID {
			log.Info("forbidden: user is not the author",
				slog.Int64("author_id", authorID),
				slog.Int64("comment_author_id", commentAuthorID),
				slog.Int64("comment_id", commentID))

			util.ErrorResponse(w, http.StatusForbidden, "Forbidden")
			return
		}

		err = commentDeleter.DeleteComment(commentID)
		if err != nil {
			if errors.Is(err, repository.ErrNotExists) {
				log.Info("comment not found", slog.Int64("comment_id", commentID), sl.Error(err))
				util.ErrorResponse(w, http.StatusNotFound, "Comment Not Found")
				return
			}
			log.Info("comment not found",
				slog.Int64("comment_id", commentID),
				sl.Error(err))

			util.ErrorResponse(w, http.StatusInternalServerError, "Internal Server Error")
			return
		}

		resp := deleteResponse{
			BaseResponse: response.BaseResponse{
				Status: http.StatusOK,
			},
			CommentID: commentID,
			AuthorID:  authorID,
		}

		log.Info("comment deleted", slog.Int64("comment_id", commentID))
		err = jsonutil.WriteJSON(w, http.StatusAccepted, resp)
		if err != nil {
			log.Error("json writer error", sl.Error(err))
			http.Error(w, "Internal Server Error", 500)
			return
		}
	}
}
