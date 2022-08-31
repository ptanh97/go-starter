package handlers

import (
	"go-starter/dto"
	"go-starter/entities"
	"go-starter/enums"
	"go-starter/errors"
	"go-starter/lib"
	"go-starter/models"
	"go-starter/repositories"
	"go-starter/response"
	"go-starter/utils"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	userRepository repositories.IUserRepository
	env            lib.Env
}

func NewAuthHandler(userRepository repositories.IUserRepository, env lib.Env) AuthHandler {
	return AuthHandler{
		userRepository,
		env,
	}
}

// @Tags    auth
// @Summary Login
// @Param   body               body   dto.LoginBody true  " "
// @Param   locale             query  string        false " " enums(en,vi)
// @Success 201                object response.Response{data=models.LoginResponse}
// @Router  /api/v1/auth/login [POST]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	body := dto.LoginBody{}
	if _, err := utils.ValidateRequestBody(w, r, &body); err != nil {
		return
	}

	user, ok := h.userRepository.FindOne(w, r, entities.User{Username: body.Username})
	if !ok {
		return
	}
	if err := bcrypt.
		CompareHashAndPassword(
			[]byte(user.HashedPassword),
			[]byte(body.Password),
		); err != nil {
		switch err {
		case bcrypt.ErrMismatchedHashAndPassword:
			errors.BadRequestException(w, r, enums.Error.InvalidPassword)
		default:
			errors.BadRequestException(w, r, err)
		}
		return
	}

	now := time.Now()
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256,
		models.CurrentUser{
			ID:        user.ID,
			Username:  user.Username,
			Role:      user.Role,
			IssuedAt:  now.Unix(),
			ExpiresAt: now.Add(h.env.JWT_EXPIRES_AT * time.Second).Unix(),
		},
	).SignedString(h.env.JWT_SECRET)

	response.WriteJSON(w, r, response.Response{
		Data: models.LoginResponse{
			AccessToken: token,
		},
	})
}
