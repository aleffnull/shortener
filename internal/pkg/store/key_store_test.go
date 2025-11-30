package store

import (
	"math/rand/v2"
	"strings"
	"testing"
)

func BenchmarkRandomString(b *testing.B) {
	const length = 1000

	b.Run("by array", func(b *testing.B) {
		for b.Loop() {
			_ = randomStringByBuffer(length)
		}
	})

	b.Run("by concatenation", func(b *testing.B) {
		for b.Loop() {
			_ = randomStringByConcatenation(length)
		}
	})

	b.Run("by string builder", func(b *testing.B) {
		for b.Loop() {
			_ = randomStringByStringBuilder(length)
		}
	})
}

func randomStringByBuffer(length int) string {
	var arr = make([]byte, length)
	for i := range arr {
		arr[i] = alphabet[rand.IntN(len(alphabet))]
	}

	return string(arr)
}

func randomStringByConcatenation(length int) string {
	result := ""
	for range length {
		result += string(alphabet[rand.IntN(len(alphabet))])
	}

	return result
}

func randomStringByStringBuilder(length int) string {
	sb := strings.Builder{}
	for range length {
		sb.WriteByte(alphabet[rand.IntN(len(alphabet))])
	}

	return sb.String()
}
