package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/aleffnull/shortener/internal/pkg/authorization"
	"github.com/aleffnull/shortener/internal/pkg/logger"
	"github.com/aleffnull/shortener/internal/pkg/utils"
	"github.com/google/uuid"
)

type UserIDOptions int

const (
	UserIDOptionsNone UserIDOptions = iota
	UserIDOptionsRequireValidToken
)

type tokenStatus int

const (
	tokenStatusUnknown tokenStatus = iota
	tokenStatusEmpty
	tokenStatusInvalid
	tokenStatusValid
)

type contextKey int

const (
	userIDCookieName            = "X-UserID"
	userIDContextKey contextKey = iota
)

func UserIDHandler(
	handlerFunc http.HandlerFunc,
	authorizationService authorization.Service,
	logger logger.Logger,
	options UserIDOptions,
) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		userID, status, err := getUserID(request, authorizationService)
		if err != nil {
			utils.HandleServerError(response, err, logger)
			return
		}

		if status != tokenStatusValid {
			// Нет валидного токена, значит, либо его нет в принципе, либо он невалиден.

			if status == tokenStatusInvalid && options == UserIDOptionsRequireValidToken {
				// Токен есть и невалиден, но хендлер требует валидный, не можем авторизовать пользователя.
				utils.HandleUnauthorized(response, "Invalid token", logger)
				return
			}

			// Либо никакого токена не было, либо он невалиден, но нам не важно.
			userID = uuid.New()
			err = setUserID(userID, response, authorizationService)
			if err != nil {
				utils.HandleServerError(response, err, logger)
				return
			}
		}

		ctx := context.WithValue(request.Context(), userIDContextKey, userID)
		requestWithUserID := request.WithContext(ctx)

		handlerFunc(response, requestWithUserID)
	}
}

func GetUserIDFromContext(ctx context.Context) uuid.UUID {
	value := ctx.Value(userIDContextKey)
	if value == nil {
		return uuid.UUID{}
	}

	return (value).(uuid.UUID)
}

func getUserID(request *http.Request, authorizationService authorization.Service) (uuid.UUID, tokenStatus, error) {
	cookie, err := request.Cookie(userIDCookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return uuid.UUID{}, tokenStatusEmpty, nil
		}

		return uuid.UUID{}, tokenStatusUnknown, fmt.Errorf("getUserID, request.Cookie failed: %w", err)
	}

	userID, err := authorizationService.GetUserIDFromToken(cookie.Value)
	if err != nil {
		if errors.Is(err, authorization.ErrTokenExpired) || errors.Is(err, authorization.ErrTokenInvalid) {
			// Токен просрочился, поэтому невалиден, или сам по себе невалиден.
			return uuid.UUID{}, tokenStatusInvalid, nil
		}

		// Все остальные ошибки считаем внутренней ошибкой сервера.
		return uuid.UUID{}, tokenStatusUnknown, fmt.Errorf("getUserID, jwt.Parse failed: %w", err)
	}

	return userID, tokenStatusValid, nil
}

func setUserID(userID uuid.UUID, response http.ResponseWriter, authorizationService authorization.Service) error {
	tokenString, err := authorizationService.CreateToken(userID)
	if err != nil {
		return fmt.Errorf("setUserID, token.SignedString failed: %w", err)
	}

	http.SetCookie(response, &http.Cookie{
		Name:  userIDCookieName,
		Value: tokenString,
	})

	return nil
}
