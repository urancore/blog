package comment

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"blog/internal/api/jsonutil"
	"blog/internal/api/response"
	"blog/internal/models"
	"blog/internal/repository"
	"blog/internal/util"
	"blog/internal/util/logger"
	"blog/internal/util/logger/sl"
)

type commentInfo struct {
	CommentID int64      `json:"comment_id,omitempty"`
	Content   string     `json:"content"`
	PostID    int64      `json:"post_id"`
	AuthorID  int64      `json:"author_id"`
	Username  string     `json:"username"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
}

type commenttListResponse struct {
	response.BaseResponse
	Data []commentInfo `json:"data"`
}

type commentGetter interface {
	ListComments(limit int, offset int, postID int64) ([]*models.Comment, error)
}

type userGetter interface {
	GetUserByID(id int64) (*models.User, error)
}

func GetList(log logger.Logger, commentGetter commentGetter, userGetter userGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := log.With("fn", "handlers.url.comment.GetList")

		qlimit := r.URL.Query().Get("limit")
		qoffset := r.URL.Query().Get("offset")
		qpostID := r.URL.Query().Get("post_id")

		limit := 10 // default
		if qlimit != "" {
			l, err := strconv.ParseInt(qlimit, 10, 64)
			if err != nil || l < 1 {
				log.Info("error parsing limit query", sl.Error(err))
				util.ErrorResponse(w, http.StatusBadRequest, "Invalid limit parameter")
				return
			}
			limit = int(l)
		}

		offset := 0 // default
		if qoffset != "" {
			o, err := strconv.ParseInt(qoffset, 10, 64)
			if err != nil || o < 0 {
				log.Info("error parsing offset query", sl.Error(err))
				util.ErrorResponse(w, http.StatusBadRequest, "Invalid offset parameter")
				return
			}
			offset = int(o)
		}
		var postID int64
		if qpostID != "" {
			o, err := strconv.ParseInt(qpostID, 10, 64)
			if err != nil || o < 0 {
				log.Info("error parsing post_id query", sl.Error(err))
				util.ErrorResponse(w, http.StatusBadRequest, "Invalid post id parametet")
				return
			}
			postID = o
		}
		// TODO: пж почини этот говнокод, n+1 запросы к бд это 90iq
		comments, err := commentGetter.ListComments(int(limit), int(offset), postID)
		if err != nil {
			log.Error("error get comments",
				sl.Error(err),
				slog.Int("limit", limit),
				slog.Int("offset", offset))
			util.ErrorResponse(w, http.StatusInternalServerError, "Internal Server Error")
			return
		}
		// TODO: пж почини этот говнокод, n+1 запросы к бд это 90iq
		responseData := make([]commentInfo, 0, len(comments))
		for _, comment := range comments {
			user, err := userGetter.GetUserByID(comment.AuthorID)
			if err != nil {
				if errors.Is(err, repository.ErrNotExists) {
					log.Info("author not found",
						slog.Int64("user_id", comment.AuthorID),
						sl.Error(err))
					continue
				}
				log.Error("error getting author",
					slog.Int64("user_id", comment.AuthorID),
					sl.Error(err))
				continue
			}
			responseData = append(responseData, commentInfo{
				CommentID: comment.ID,
				Content:   comment.Content,
				PostID:    comment.PostID,
				AuthorID:  comment.AuthorID,
				Username:  user.Username,
				CreatedAt: &comment.CreatedAt,
			})
		}

		resp := commenttListResponse{
			BaseResponse: response.BaseResponse{
				Status: http.StatusOK,
			},
			Data: responseData,
		}
		if err := jsonutil.WriteJSON(w, http.StatusOK, resp); err != nil {
			log.Error("failed to write JSON response", sl.Error(err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}

	}
}
