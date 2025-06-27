package user

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	// "time"

	"github.com/go-playground/validator/v10"

	"blog/internal/api/jsonutil"
	"blog/internal/api/response"
	// "blog/internal/handlers/url/token"
	"blog/internal/models"
	"blog/internal/repository"
	"blog/internal/util"
	"blog/internal/util/logger"
	"blog/internal/util/logger/sl"
)

type SignUpRequest struct {
	Username string `json:"username" validate:"username,required"`
	Email    string `json:"email" validate:"email,required"`
	Password string `json:"password" validate:"password,required"`
}

type UserRes struct {
	UserID   int64  `json:"user_id"`
}

type SignUpResponse struct {
	response.BaseResponse
	// AccessToken string  `json:"access_token"`
	User        UserRes `json:"user"`
}

type UserCreator interface {
	CreateUser(user *models.User) (int64, error)
}

func CreateUser(user *models.User, userCreator UserCreator) (int64, error) {
	hash := util.GeneratePasswordHash(user.Password)
	user.Password = hash

	userID, err := userCreator.CreateUser(user)
	if err != nil {
		if errors.Is(err, repository.ErrEmailAlreadyExists) {
			return 0, err
		} else if errors.Is(err, repository.ErrUsernameAlreadyExists) {
			return 0, err
		}
		return 0, err
	}

	return userID, nil
}

func SignUpHandler(log logger.Logger, userCreator UserCreator) http.HandlerFunc {
	cstValidator := util.NewCustomValidator()
	return func(w http.ResponseWriter, r *http.Request) {
		log := log.With(slog.String("fn", "handlers.url.user.SignUp"))

		if r.Method != http.MethodPost {
			jsonutil.WriteJSON(w, http.StatusMethodNotAllowed,
				response.Error(http.StatusMethodNotAllowed, "method not allowed"))
			return
		}

		var req SignUpRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Info("error decoding request", sl.Error(err))
			jsonutil.WriteJSON(w, http.StatusBadRequest,
				response.Error(http.StatusBadRequest, "invalid request"))
			return
		}

		if err := cstValidator.Struct(req); err != nil {
			validationErrors := err.(validator.ValidationErrors)
			log.Info("validation failed", sl.Error(validationErrors))
			jsonutil.WriteJSON(w, http.StatusBadRequest,
				response.Error(http.StatusBadRequest, "validation error"))
			return
		}

		user := &models.User{
			Username: req.Username,
			Email:    req.Email,
			Password: req.Password,
		}

		userID, err := CreateUser(user, userCreator)
		if err != nil {
			if errors.Is(err, repository.ErrEmailAlreadyExists) {
				log.Info("email exists", slog.String("email", req.Email))
				jsonutil.WriteJSON(w, http.StatusConflict,
					response.Error(http.StatusConflict, "email already exists"))
				return
			} else if errors.Is(err, repository.ErrUsernameAlreadyExists) {
				log.Info("username exists", slog.String("username", req.Username))
				jsonutil.WriteJSON(w, http.StatusConflict,
					response.Error(http.StatusConflict, "username already exists"))
				return
			}

			log.Error("create user error", sl.Error(err))
			jsonutil.WriteJSON(w, http.StatusInternalServerError,
				response.Error(http.StatusInternalServerError, "internal error"))
			return
		}

		// accessToken, _, err := jwtMaker.CreateToken(userID, req.Email, 15*time.Minute)
		// if err != nil {
		// 	log.Error("token creation failed", sl.Error(err))
		// 	jsonutil.WriteJSON(w, http.StatusInternalServerError,
		// 		response.Error(http.StatusInternalServerError, "token error"))
		// 	return
		// }

		resp := SignUpResponse{
			BaseResponse: response.BaseResponse{
				Status: http.StatusCreated,
			},
			User: UserRes{
				UserID:   userID,
			},
		}

		// w.Header().Set("Authorization", "Bearer "+accessToken)
		log.Info("user created", slog.Int64("user_id", userID))
		jsonutil.WriteJSON(w, http.StatusCreated, resp)
	}
}
