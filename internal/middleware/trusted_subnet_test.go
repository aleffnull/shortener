package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aleffnull/shortener/internal/config"
	"github.com/aleffnull/shortener/internal/pkg/mocks"
	"github.com/go-http-utils/headers"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestTrustedSubnetHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		want       int
		hookBefore func(mock *mocks.Mock) (*config.Configuration, *http.Request)
	}{
		{
			name: "WHEN trusted subnet not configured THEN forbidden",
			want: http.StatusForbidden,
			hookBefore: func(mock *mocks.Mock) (*config.Configuration, *http.Request) {
				mock.Logger.EXPECT().Warnf(gomock.Any(), gomock.Any())
				configuration := &config.Configuration{}
				request := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
				return configuration, request
			},
		},
		{
			name: "WHEN no real ip in request headers THEN forbidden",
			want: http.StatusForbidden,
			hookBefore: func(mock *mocks.Mock) (*config.Configuration, *http.Request) {
				mock.Logger.EXPECT().Warnf(gomock.Any(), gomock.Any())
				configuration := &config.Configuration{
					TrustedSubnet: "192.168.1.0/24",
				}
				request := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
				return configuration, request
			},
		},
		{
			name: "WHEN real ip parse error THEN server error",
			want: http.StatusInternalServerError,
			hookBefore: func(mock *mocks.Mock) (*config.Configuration, *http.Request) {
				mock.Logger.EXPECT().Errorf(gomock.Any(), gomock.Any())
				configuration := &config.Configuration{
					TrustedSubnet: "192.168.1.0/24",
				}
				request := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
				request.Header.Add(headers.XRealIP, "foo")
				return configuration, request
			},
		},
		{
			name: "WHEN not trusted subnet THEN forbidden",
			want: http.StatusForbidden,
			hookBefore: func(mock *mocks.Mock) (*config.Configuration, *http.Request) {
				mock.Logger.EXPECT().Warnf(gomock.Any(), gomock.Any())
				configuration := &config.Configuration{
					TrustedSubnet: "192.168.1.0/24",
				}
				request := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
				request.Header.Add(headers.XRealIP, "10.10.10.15")
				return configuration, request
			},
		},
		{
			name: "WHEN trusted subnet THEN handler called",
			want: http.StatusOK,
			hookBefore: func(mock *mocks.Mock) (*config.Configuration, *http.Request) {
				configuration := &config.Configuration{
					TrustedSubnet: "192.168.1.0/24",
				}
				request := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
				request.Header.Add(headers.XRealIP, "192.168.1.42")
				return configuration, request
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			ctrl := gomock.NewController(t)
			mock := mocks.NewMock(ctrl)
			configuration, request := tt.hookBefore(mock)
			trustedSubnetHandler := TrustedSubnetHandler(
				func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				},
				configuration,
				mock.Logger,
			)
			recorder := httptest.NewRecorder()

			// Act.
			trustedSubnetHandler(recorder, request)

			// Assert.
			require.Equal(t, tt.want, recorder.Code)
		})
	}
}

func TestTrustedSubnetHandler_isInCIDR(t *testing.T) {
	t.Parallel()

	type args struct {
		ip   string
		cidr string
	}

	tests := []struct {
		name      string
		args      *args
		want      bool
		wantError bool
	}{
		{
			name: "WHEN cidr parse error THEN error",
			args: &args{
				cidr: "foo",
			},
			wantError: true,
		},
		{
			name: "WHEN ip parse error THEN error",
			args: &args{
				cidr: "192.168.1.0/24",
				ip:   "foo",
			},
			wantError: true,
		},
		{
			name: "WHEN ip in not in cidr THEN false",
			args: &args{
				cidr: "192.168.1.0/24",
				ip:   "10.10.10.15",
			},
			want: false,
		},
		{
			name: "WHEN ip in in cidr THEN true",
			args: &args{
				cidr: "192.168.1.0/24",
				ip:   "192.168.1.42",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := isInCIDR(tt.args.ip, tt.args.cidr)
			require.Equal(t, tt.want, got)
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
