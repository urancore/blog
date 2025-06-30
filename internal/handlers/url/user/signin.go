package user

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"

	"blog/internal/api/jsonutil"
	"blog/internal/api/response"
	"blog/internal/handlers/auth"
	"blog/internal/models"
	"blog/internal/repository"
	"blog/internal/util"
	"blog/internal/util/logger"
	"blog/internal/util/logger/sl"
)

// Структуры для авторизации
type SignInRequest struct {
	Email    string `json:"email" validate:"email,required"`
	Password string `json:"password" validate:"password,required"`
}

type SignInResponse struct {
	response.BaseResponse
	AuthToken string `json:"auth_token"`
}

type UserGetter interface {
	GetUserByEmail(email string) (*models.User, error)
}

func SignInHandler(log logger.Logger, userGetter UserGetter) http.HandlerFunc {
	cstValidator := util.NewCustomValidator()
	return func(w http.ResponseWriter, r *http.Request) {
		log := log.With(slog.String("fn", "handlers.url.user.SignIn"))

		if r.Method != http.MethodPost {
			util.ErrorResponse(w, http.StatusMethodNotAllowed, "Method Not Allowed")
			return
		}

		var req SignInRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Info("decoding error", sl.Error(err))
			util.ErrorResponse(w, http.StatusBadRequest, "Invalid Request")
			return
		}

		if err := cstValidator.Struct(req); err != nil {
			validationErrors := err.(validator.ValidationErrors)
			log.Info("validation failed", sl.Error(validationErrors))
			util.ErrorResponse(w, http.StatusBadRequest, "Validation Error")
			return
		}

		user, err := userGetter.GetUserByEmail(req.Email)
		if err != nil {
			if errors.Is(err, repository.ErrNotExists) {
				log.Info("user not found", slog.String("email", req.Email))
				util.ErrorResponse(w, http.StatusNotFound, "User Not Found")
				return
			}

			log.Error("get user error", sl.Error(err))
			util.ErrorResponse(w, http.StatusInternalServerError, "Internal Server Error")
			return
		}

		if !util.CheckPasswordHash(req.Password, user.Password) {
			log.Info("invalid password", slog.String("email", req.Email))
			util.ErrorResponse(w, http.StatusUnauthorized, "Invalid Credentials")
			return
		}

		token, err := auth.GenerateToken(user.ID, user.Username, user.Password)
		if err != nil {
			log.Error("error generation token", sl.Error(err), slog.Int64("user_id", user.ID))
			util.ErrorResponse(w, http.StatusInternalServerError, "Internal Server Error")
			return
		}

		resp := SignInResponse{
			BaseResponse: response.BaseResponse{
				Status: http.StatusAccepted,
			},
			AuthToken: token,
		}

		log.Info("user authentificate", slog.Int64("user_id", user.ID))
		jsonutil.WriteJSON(w, http.StatusAccepted, resp)
	}
}
