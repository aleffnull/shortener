package middleware

import (
	"fmt"
	"net"
	"net/http"

	"github.com/aleffnull/shortener/internal/config"
	"github.com/aleffnull/shortener/internal/pkg/logger"
	"github.com/aleffnull/shortener/internal/pkg/utils"
	"github.com/go-http-utils/headers"
)

func TrustedSubnetHandler(
	next http.HandlerFunc,
	configuration *config.Configuration,
	logger logger.Logger,
) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if len(configuration.TrustedSubnet) == 0 {
			utils.HandleForbidden(writer, "Trusted subnet is not configured", logger)
			return
		}

		realIP := request.Header.Get(headers.XRealIP)
		if len(realIP) == 0 {
			utils.HandleForbidden(writer, "No read IP in request headers", logger)
			return
		}

		ok, err := isInCIDR(realIP, configuration.TrustedSubnet)
		if err != nil {
			utils.HandleServerError(writer, err, logger)
			return
		}

		if !ok {
			utils.HandleForbidden(writer, "Request from untrusted network", logger)
			return
		}

		next.ServeHTTP(writer, request)
	}
}

func isInCIDR(ipStr string, cidrStr string) (bool, error) {
	_, ipnet, err := net.ParseCIDR(cidrStr)
	if err != nil {
		return false, fmt.Errorf("failed to parse CIDR: %w", err)
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false, fmt.Errorf("failed to parse IP: %w", err)
	}

	return ipnet.Contains(ip), nil
}
