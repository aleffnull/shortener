package app

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aleffnull/shortener/internal/pkg/audit"
	"github.com/aleffnull/shortener/internal/pkg/mocks"
	"github.com/aleffnull/shortener/models"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestRouter_NewMuxHandler(t *testing.T) {
	t.Parallel()

	// Arrange.
	ctrl := gomock.NewController(t)
	mock := mocks.NewMock(ctrl)

	maintenanceHandler := NewMaintenanceHandler(mock.App, mock.Logger)
	simpleAPIHandler := NewSimpleAPIHandler(mock.App, mock.AuditService, mock.Logger)
	apiHandler := NewAPIHandler(mock.App, mock.AuditService, mock.Logger)
	userHandler := NewUserHandler(mock.App, mock.Logger)
	router := NewRouter(
		maintenanceHandler,
		simpleAPIHandler,
		apiHandler,
		userHandler,
		mock.AuthorizationService,
		mock.Logger,
	)

	// Act.
	handler := router.NewMuxHandler()

	// Assert.
	mux, ok := handler.(*chi.Mux)
	require.True(t, ok)
	require.NotEmpty(t, mux.Routes())
}

func TestRouter_ServeHTTP(t *testing.T) {
	t.Parallel()

	const (
		key     = "foo"
		fullURL = "http://bar.buz"
	)

	userID := uuid.New()

	// Arrange.
	ctrl := gomock.NewController(t)
	mock := mocks.NewMock(ctrl)
	mock.AuthorizationService.EXPECT().CreateToken(gomock.Any()).Return("token", nil)
	mock.App.EXPECT().
		GetURL(gomock.Any(), key).
		Return(&models.GetURLResponseItem{
			URL:    fullURL,
			UserID: userID,
		}, nil)
	mock.AuditService.EXPECT().AuditEvent(gomock.Any()).DoAndReturn(func(event *audit.Event) {
		require.LessOrEqual(t, event.Timestamp, time.Now())
		require.Equal(t, audit.ActionFollow, event.Action)
		require.Equal(t, userID, event.UserID)
		require.Equal(t, fullURL, event.URL)
	})
	mock.Logger.EXPECT().Infof(gomock.Any())

	maintenanceHandler := NewMaintenanceHandler(mock.App, mock.Logger)
	simpleAPIHandler := NewSimpleAPIHandler(mock.App, mock.AuditService, mock.Logger)
	apiHandler := NewAPIHandler(mock.App, mock.AuditService, mock.Logger)
	userHandler := NewUserHandler(mock.App, mock.Logger)
	router := NewRouter(
		maintenanceHandler,
		simpleAPIHandler,
		apiHandler,
		userHandler,
		mock.AuthorizationService,
		mock.Logger,
	)

	handler := router.NewMuxHandler()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/"+key, nil)

	// Act.
	handler.ServeHTTP(recorder, request)
}
