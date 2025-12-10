package app

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/aleffnull/shortener/internal/config"
	"github.com/aleffnull/shortener/internal/domain"
	"github.com/aleffnull/shortener/internal/pkg/mocks"
	"github.com/aleffnull/shortener/internal/pkg/store"
	"github.com/aleffnull/shortener/models"
)

func TestShortenerApp_Init(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		wantError  bool
		hookBefore func(mock *mocks.Mock)
	}{
		{
			name:      "WHEN storage error THEN error",
			wantError: true,
			hookBefore: func(mock *mocks.Mock) {
				mock.Store.EXPECT().Init().Return(assert.AnError)
			},
		},
		{
			name:      "WHEN connection error THEN error",
			wantError: true,
			hookBefore: func(mock *mocks.Mock) {
				mock.Store.EXPECT().Init().Return(nil)
				mock.Connection.EXPECT().Init(gomock.Any()).Return(assert.AnError)
			},
		},
		{
			name:      "WHEN parameters error THEN error",
			wantError: true,
			hookBefore: func(mock *mocks.Mock) {
				mock.Store.EXPECT().Init().Return(nil)
				mock.Connection.EXPECT().Init(gomock.Any()).Return(nil)
				mock.AppParameters.EXPECT().Init(gomock.Any()).Return(assert.AnError)
			},
		},
		{
			name: "WHEN no errors THEN ok",
			hookBefore: func(mock *mocks.Mock) {
				mock.Store.EXPECT().Init().Return(nil)
				mock.Connection.EXPECT().Init(gomock.Any()).Return(nil)
				mock.AppParameters.EXPECT().Init(gomock.Any()).Return(nil)
				mock.AuditService.EXPECT().Init()
				mock.DeleteURLsService.EXPECT().Init()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mock := mocks.NewMock(ctrl)
			tt.hookBefore(mock)
			shortener := NewShortenerApp(
				mock.Connection,
				mock.Store,
				mock.DeleteURLsService,
				mock.AuditService,
				mock.Logger,
				mock.AppParameters,
				&config.Configuration{},
			)

			err := shortener.Init(context.Background())
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestShortenerApp_Shutdown(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := mocks.NewMock(ctrl)
	mock.DeleteURLsService.EXPECT().Shutdown()
	mock.AuditService.EXPECT().Shutdown()
	mock.Connection.EXPECT().Shutdown()

	shortener := NewShortenerApp(
		mock.Connection,
		mock.Store,
		mock.DeleteURLsService,
		mock.AuditService,
		mock.Logger,
		mock.AppParameters,
		&config.Configuration{},
	)
	shortener.Shutdown()
}

func TestShortenerApp_GetURL(t *testing.T) {
	t.Parallel()

	type args struct {
		key string
	}

	tests := []struct {
		name       string
		args       *args
		wantError  bool
		hookBefore func(mocks *mocks.Mock, args *args) *models.GetURLResponseItem
	}{
		{
			name: "WHEN storage error THEN error",
			args: &args{
				key: "foo",
			},
			wantError: true,
			hookBefore: func(mocks *mocks.Mock, args *args) *models.GetURLResponseItem {
				mocks.Store.EXPECT().Load(gomock.Any(), args.key).Return(nil, assert.AnError)
				return nil
			},
		},
		{
			name: "WHEN nil item loaded THEN nil returned",
			args: &args{
				key: "foo",
			},
			hookBefore: func(mocks *mocks.Mock, args *args) *models.GetURLResponseItem {
				mocks.Store.EXPECT().Load(gomock.Any(), args.key).Return(nil, nil)
				return nil
			},
		},
		{
			name: "WHEN item loaded THEN item returned",
			args: &args{
				key: "foo",
			},
			hookBefore: func(mocks *mocks.Mock, args *args) *models.GetURLResponseItem {
				urlItem := &domain.URLItem{
					URL:       "http://localhost/bar",
					UserID:    uuid.New(),
					IsDeleted: false,
				}
				responseItem := &models.GetURLResponseItem{
					URL:       urlItem.URL,
					UserID:    urlItem.UserID,
					IsDeleted: urlItem.IsDeleted,
				}
				mocks.Store.EXPECT().Load(gomock.Any(), args.key).Return(urlItem, nil)
				return responseItem
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			ctrl := gomock.NewController(t)
			mock := mocks.NewMock(ctrl)
			want := tt.hookBefore(mock, tt.args)

			shortener := NewShortenerApp(
				mock.Connection,
				mock.Store,
				mock.DeleteURLsService,
				mock.AuditService,
				mock.Logger,
				mock.AppParameters,
				&config.Configuration{},
			)

			// Act.
			got, err := shortener.GetURL(context.Background(), tt.args.key)

			// Assert.
			require.Equal(t, want, got)
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestShortenerApp_GetUserURLs(t *testing.T) {
	t.Parallel()

	type args struct {
		userID uuid.UUID
	}

	tests := []struct {
		name       string
		args       *args
		wantError  bool
		hookBefore func(mocks *mocks.Mock, args *args) (*config.Configuration, []*models.UserURLsResponseItem)
	}{
		{
			name: "WHEN storage error THEN error",
			args: &args{
				userID: uuid.New(),
			},
			wantError: true,
			hookBefore: func(mocks *mocks.Mock, args *args) (*config.Configuration, []*models.UserURLsResponseItem) {
				mocks.Store.EXPECT().LoadAllByUserID(gomock.Any(), args.userID).Return(nil, assert.AnError)
				return &config.Configuration{}, nil
			},
		},
		{
			name: "WHEN join path error THEN error",
			args: &args{
				userID: uuid.New(),
			},
			wantError: true,
			hookBefore: func(mocks *mocks.Mock, args *args) (*config.Configuration, []*models.UserURLsResponseItem) {
				mocks.Store.EXPECT().LoadAllByUserID(gomock.Any(), args.userID).Return([]*domain.KeyOriginalURLItem{
					{
						URLKey:      "foo",
						OriginalURL: "http://foo.bar",
					},
				}, nil)
				return &config.Configuration{
					BaseURL: ":::\\::",
				}, nil
			},
		},
		{
			name: "WHEN no errors THEN ok",
			args: &args{
				userID: uuid.New(),
			},
			hookBefore: func(mocks *mocks.Mock, args *args) (*config.Configuration, []*models.UserURLsResponseItem) {
				mocks.Store.EXPECT().LoadAllByUserID(gomock.Any(), args.userID).Return([]*domain.KeyOriginalURLItem{
					{
						URLKey:      "foo",
						OriginalURL: "http://foo.bar",
					},
				}, nil)
				configuration := &config.Configuration{
					BaseURL: "http://localhost",
				}
				want := []*models.UserURLsResponseItem{
					{
						ShortURL:    "http://localhost/foo",
						OriginalURL: "http://foo.bar",
					},
				}
				return configuration, want
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			ctrl := gomock.NewController(t)
			mock := mocks.NewMock(ctrl)
			configuration, want := tt.hookBefore(mock, tt.args)

			shortener := NewShortenerApp(
				mock.Connection,
				mock.Store,
				mock.DeleteURLsService,
				mock.AuditService,
				mock.Logger,
				mock.AppParameters,
				configuration,
			)

			// Act.
			got, err := shortener.GetUserURLs(context.Background(), tt.args.userID)

			// Assert.
			require.Equal(t, want, got)
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestShortenerApp_ShortenURL(t *testing.T) {
	t.Parallel()

	type args struct {
		request *models.ShortenRequest
		userID  uuid.UUID
	}

	tests := []struct {
		name       string
		args       *args
		wantError  bool
		hookBefore func(mock *mocks.Mock, args *args) (*config.Configuration, *models.ShortenResponse)
	}{
		{
			name: "WHEN storage error THEN error",
			args: &args{
				request: &models.ShortenRequest{
					URL: "http://foo.bar",
				},
				userID: uuid.New(),
			},
			wantError: true,
			hookBefore: func(mock *mocks.Mock, args *args) (*config.Configuration, *models.ShortenResponse) {
				mock.Store.EXPECT().Save(gomock.Any(), args.request.URL, args.userID).Return("", assert.AnError)
				return nil, nil
			},
		},
		{
			name: "WHEN join path error THEN error",
			args: &args{
				request: &models.ShortenRequest{
					URL: "http://foo.bar",
				},
				userID: uuid.New(),
			},
			wantError: true,
			hookBefore: func(mock *mocks.Mock, args *args) (*config.Configuration, *models.ShortenResponse) {
				mock.Store.EXPECT().Save(gomock.Any(), args.request.URL, args.userID).Return("foo", nil)
				return &config.Configuration{
					BaseURL: ":::\\::",
				}, nil
			},
		},
		{
			name: "GIVEN not duplicate WHEN no errors THEN ok",
			args: &args{
				request: &models.ShortenRequest{
					URL: "http://foo.bar",
				},
				userID: uuid.New(),
			},
			hookBefore: func(mock *mocks.Mock, args *args) (*config.Configuration, *models.ShortenResponse) {
				mock.Store.EXPECT().Save(gomock.Any(), args.request.URL, args.userID).Return("foo", nil)
				configuration := &config.Configuration{
					BaseURL: "http://localhost",
				}
				want := &models.ShortenResponse{
					Result:      "http://localhost/foo",
					IsDuplicate: false,
				}
				return configuration, want
			},
		},
		{
			name: "GIVEN duplicate WHEN no errors THEN ok",
			args: &args{
				request: &models.ShortenRequest{
					URL: "http://foo.bar",
				},
				userID: uuid.New(),
			},
			hookBefore: func(mock *mocks.Mock, args *args) (*config.Configuration, *models.ShortenResponse) {
				err := &store.DuplicateURLError{
					Key: "bar",
				}
				mock.Store.EXPECT().Save(gomock.Any(), args.request.URL, args.userID).Return("", err)
				mock.Logger.EXPECT().Infof(gomock.Any(), err)
				configuration := &config.Configuration{
					BaseURL: "http://localhost",
				}
				want := &models.ShortenResponse{
					Result:      "http://localhost/bar",
					IsDuplicate: true,
				}
				return configuration, want
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mock := mocks.NewMock(ctrl)
			configuration, wantResponse := tt.hookBefore(mock, tt.args)

			shortener := NewShortenerApp(
				mock.Connection,
				mock.Store,
				mock.DeleteURLsService,
				mock.AuditService,
				mock.Logger,
				mock.AppParameters,
				configuration,
			)

			gotResponse, err := shortener.ShortenURL(context.Background(), tt.args.request, tt.args.userID)
			require.Equal(t, wantResponse, gotResponse)
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestShortenerApp_ShortenURLBatch(t *testing.T) {
	t.Parallel()

	type args struct {
		requestItems []*models.ShortenBatchRequestItem
		userID       uuid.UUID
	}

	tests := []struct {
		name       string
		args       *args
		wantError  bool
		hookBefore func(mock *mocks.Mock, args *args) (*config.Configuration, []*models.ShortenBatchResponseItem)
	}{
		{
			name: "WHEN no request items THEN do nothing",
			args: &args{
				requestItems: []*models.ShortenBatchRequestItem{},
			},
			hookBefore: func(_ *mocks.Mock, _ *args) (*config.Configuration, []*models.ShortenBatchResponseItem) {
				return nil, []*models.ShortenBatchResponseItem{}
			},
		},
		{
			name: "WHEN storage error THEN error",
			args: &args{
				requestItems: []*models.ShortenBatchRequestItem{
					{
						CorrelationID: "foo",
						OriginalURL:   "http://foo.bar",
					},
				},
				userID: uuid.New(),
			},
			wantError: true,
			hookBefore: func(mock *mocks.Mock, args *args) (*config.Configuration, []*models.ShortenBatchResponseItem) {
				mock.Store.EXPECT().
					SaveBatch(gomock.Any(), []*domain.BatchRequestItem{
						{
							CorrelationID: args.requestItems[0].CorrelationID,
							OriginalURL:   args.requestItems[0].OriginalURL,
						},
					}, args.userID).
					Return(nil, assert.AnError)
				return nil, nil
			},
		},
		{
			name: "WHEN join path error THEN error",
			args: &args{
				requestItems: []*models.ShortenBatchRequestItem{
					{
						CorrelationID: "42",
						OriginalURL:   "http://foo.bar",
					},
				},
				userID: uuid.New(),
			},
			wantError: true,
			hookBefore: func(mock *mocks.Mock, args *args) (*config.Configuration, []*models.ShortenBatchResponseItem) {
				mock.Store.EXPECT().
					SaveBatch(gomock.Any(), []*domain.BatchRequestItem{
						{
							CorrelationID: args.requestItems[0].CorrelationID,
							OriginalURL:   args.requestItems[0].OriginalURL,
						},
					}, args.userID).
					Return([]*domain.BatchResponseItem{
						{
							CorrelationID: args.requestItems[0].CorrelationID,
							Key:           "foo",
						},
					}, nil)
				return &config.Configuration{
					BaseURL: ":::\\::",
				}, nil
			},
		},
		{
			name: "WHEN no errors THEN ok",
			args: &args{
				requestItems: []*models.ShortenBatchRequestItem{
					{
						CorrelationID: "42",
						OriginalURL:   "http://foo.bar",
					},
				},
				userID: uuid.New(),
			},
			hookBefore: func(mock *mocks.Mock, args *args) (*config.Configuration, []*models.ShortenBatchResponseItem) {
				mock.Store.EXPECT().
					SaveBatch(gomock.Any(), []*domain.BatchRequestItem{
						{
							CorrelationID: args.requestItems[0].CorrelationID,
							OriginalURL:   args.requestItems[0].OriginalURL,
						},
					}, args.userID).
					Return([]*domain.BatchResponseItem{
						{
							CorrelationID: args.requestItems[0].CorrelationID,
							Key:           "foo",
						},
					}, nil)
				configuration := &config.Configuration{
					BaseURL: "http://localhost",
				}
				responseItems := []*models.ShortenBatchResponseItem{
					{
						CorrelationID: args.requestItems[0].CorrelationID,
						ShortURL:      "http://localhost/foo",
					},
				}
				return configuration, responseItems
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mock := mocks.NewMock(ctrl)
			configuration, wantResponse := tt.hookBefore(mock, tt.args)

			shortener := NewShortenerApp(
				mock.Connection,
				mock.Store,
				mock.DeleteURLsService,
				mock.AuditService,
				mock.Logger,
				mock.AppParameters,
				configuration,
			)

			gotResponse, err := shortener.ShortenURLBatch(context.Background(), tt.args.requestItems, tt.args.userID)
			require.Equal(t, wantResponse, gotResponse)
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestShortenerApp_DeleteURLs(t *testing.T) {
	t.Parallel()

	type argsT struct {
		keys   []string
		userID uuid.UUID
	}

	args := &argsT{
		keys:   []string{"foo", "bar"},
		userID: uuid.New(),
	}

	// Arrange.
	ctrl := gomock.NewController(t)
	mock := mocks.NewMock(ctrl)
	mock.DeleteURLsService.EXPECT().Delete(&domain.DeleteURLsRequest{
		Keys:   args.keys,
		UserID: args.userID,
	})
	shortener := NewShortenerApp(
		mock.Connection,
		mock.Store,
		mock.DeleteURLsService,
		mock.AuditService,
		mock.Logger,
		mock.AppParameters,
		nil,
	)

	// Act.
	shortener.DeleteURLs(args.keys, args.userID)
}

func TestShortenerApp_CheckStore(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		wantError  bool
		hookBefore func(mocks *mocks.Mock)
	}{
		{
			name:      "WHEN store error THEN error",
			wantError: true,
			hookBefore: func(mocks *mocks.Mock) {
				mocks.Store.EXPECT().CheckAvailability(gomock.Any()).Return(assert.AnError)
			},
		},
		{
			name: "WHEN no errors THEN ok",
			hookBefore: func(mocks *mocks.Mock) {
				mocks.Store.EXPECT().CheckAvailability(gomock.Any()).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mock := mocks.NewMock(ctrl)
			tt.hookBefore(mock)
			shortener := NewShortenerApp(
				mock.Connection,
				mock.Store,
				mock.DeleteURLsService,
				mock.AuditService,
				mock.Logger,
				mock.AppParameters,
				nil,
			)

			err := shortener.CheckStore(context.Background())
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
