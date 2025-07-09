package store

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"

	"github.com/aleffnull/shortener/internal/config"
)

const (
	alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

type saverFunc func(ctx context.Context, key, value string) (bool, error)

type keyStore struct {
	configuration *config.KeyStoreConfiguration
}

func (s *keyStore) saveWithUniqueKey(ctx context.Context, value string, saver saverFunc) (string, error) {
	length := s.configuration.KeyLength
	i := 0

	for length <= s.configuration.KeyMaxLength {
		key := randomString(length)
		exists, err := saver(ctx, key, value)
		if err != nil {
			return "", fmt.Errorf("saveWithUniqueKey, saver failed: %w", err)
		}

		if !exists {
			return key, nil
		}

		i++
		if i >= s.configuration.KeyMaxIterations {
			length *= 2
			i = 0
		}
	}

	return "", errors.New("failed to generate unique key")
}

func randomString(length int) string {
	var arr = make([]byte, length)
	for i := range arr {
		arr[i] = alphabet[rand.IntN(len(alphabet))]
	}

	return string(arr)
}
