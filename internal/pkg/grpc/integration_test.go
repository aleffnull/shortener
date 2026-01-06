package grpc

import (
	"context"
	"net/url"
	"testing"

	"github.com/aleffnull/shortener/internal/pkg/mocks"
	"github.com/aleffnull/shortener/internal/pkg/pb/shortener/api"
	"github.com/aleffnull/shortener/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestGRPC(t *testing.T) {
	t.Parallel()

	const (
		fullURL       = "http://foo.bar"
		jwtSigningKey = "50db3642a43bc2af1635eb0c21edd092"
	)

	ctrl := gomock.NewController(t)
	mock := mocks.NewMock(ctrl)
	mock.AppParameters.EXPECT().GetJWTSigningKey().Return(jwtSigningKey)

	client := getClient(t)
	token := getToken(t, mock)
	ctx := getContext(t, token)

	// Сокращаем URL.
	shortenRequest := &api.URLShortenRequest{
		Url: fullURL,
	}
	shortenResponse, err := client.ShortenURL(ctx, shortenRequest)
	status, ok := status.FromError(err)
	require.True(t, ok)
	if status.Code() == codes.Unavailable {
		// Сервис не запущен, не мешаем остальным тестам.
		return
	}

	require.NoError(t, err)
	require.NotEmpty(t, shortenResponse.GetResult())

	// Получаем полным URL по ключу из сокращенного.
	shortURL, err := url.Parse(shortenResponse.GetResult())
	require.NoError(t, err)
	expandRequest := &api.URLExpandRequest{
		Id: shortURL.Path[1:],
	}
	expandResponse, err := client.ExpandURL(ctx, expandRequest)
	require.NoError(t, err)
	require.Equal(t, fullURL, expandResponse.GetResult())

	// Получаем список URL пользователя.
	listUserURLsResponse, err := client.ListUserURLs(ctx, &emptypb.Empty{})
	require.NoError(t, err)
	require.Len(t, listUserURLsResponse.GetUrl(), 1)
	require.Equal(t, fullURL, listUserURLsResponse.GetUrl()[0].GetOriginalUrl())
	require.Equal(t, shortenResponse.GetResult(), listUserURLsResponse.GetUrl()[0].GetShortUrl())
}

func getClient(t *testing.T) api.ShortenerServiceClient {
	connection, err := grpc.NewClient("localhost:8181", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, connection.Close())
	})

	return api.NewShortenerServiceClient(connection)
}

func getToken(t *testing.T, mock *mocks.Mock) string {
	authorizationService := service.NewAuthorizationService(mock.AppParameters)
	token, err := authorizationService.CreateToken(uuid.New())
	require.NoError(t, err)

	return token
}

func getContext(t *testing.T, token string) context.Context {
	md := metadata.New(map[string]string{"authorization": token})
	return metadata.NewOutgoingContext(t.Context(), md)
}
