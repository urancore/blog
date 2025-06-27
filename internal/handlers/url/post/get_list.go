package post

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

// TODO: reafactor struct
type postInfo struct {
	PostID    int64      `json:"post_id,omitempty"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	AuthorID  int64      `json:"author_id"`
	Username  string     `json:"username"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
}

type postListResponse struct {
	response.BaseResponse
	Data []postInfo `json:"data"`
}

type postsGetter interface {
	ListPosts(limit, offset int) ([]*models.Post, error)
}
type userGetter interface {
	GetUserByID(id int64) (*models.User, error)
}

func GetList(log logger.Logger, postsGetter postsGetter, userGetter userGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := log.With("fn", "handlers.url.post.GetList")

		qlimit := r.URL.Query().Get("limit")
		qoffset := r.URL.Query().Get("offset")

		limit := 10 // default
		if qlimit != "" {
			l, err := strconv.ParseInt(qlimit, 10, 64)
			if err != nil || l < 1 {
				util.ErrorResponse(w, http.StatusBadRequest, "Invalid limit parameter")
				return
			}
			limit = int(l)
		}

		offset := 0 // default
		if qoffset != "" {
			o, err := strconv.ParseInt(qoffset, 10, 64)
			if err != nil || o < 0 {
				util.ErrorResponse(w, http.StatusBadRequest, "Invalid offset parameter")
				return
			}
			offset = int(o)
		}
		// FIXME: add users, err := GetUsersByIDs(authorIDs)
		// user, exists := users[post.AuthorID]

		posts, err := postsGetter.ListPosts(limit, offset)
		if err != nil {
			log.Error("error get posts",
				sl.Error(err),
				slog.Int("limit", limit),
				slog.Int("offset", offset))
			util.ErrorResponse(w, http.StatusInternalServerError, "Internal Server Error")
			return
		}

		// TODO: ADD REDIS for cache
		responseData := make([]postInfo, 0, len(posts))
		for _, post := range posts {
			user, err := userGetter.GetUserByID(post.AuthorID)
			if err != nil {
				if errors.Is(err, repository.ErrNotExists) {
					log.Info("author not found",
						slog.Int64("user_id", post.AuthorID),
						sl.Error(err))
					continue
				}
				log.Error("error getting author",
					slog.Int64("user_id", post.AuthorID),
					sl.Error(err))
				continue
			}
			responseData = append(responseData, postInfo{
				PostID:    post.ID,
				Title:     post.Title,
				Content:   post.Content,
				AuthorID:  post.AuthorID,
				Username:  user.Username,
				CreatedAt: &post.CreatedAt,
			})
		}

		resp := postListResponse{
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
