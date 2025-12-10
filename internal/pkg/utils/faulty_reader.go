package utils

import "github.com/stretchr/testify/assert"

type FaultyReader func(p []byte) (n int, err error)

func (r FaultyReader) Read(p []byte) (n int, err error) {
	return r(p)
}

var AFaultyReader = FaultyReader(func(p []byte) (n int, err error) {
	return 0, assert.AnError
})
