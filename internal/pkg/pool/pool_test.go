package pool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPool(t *testing.T) {
	t.Parallel()

	objectsCount := 0
	pool := New(func() *ObjectWithState {
		objectsCount++
		return &ObjectWithState{}
	})

	object := pool.Get()
	object.State = 42
	pool.Put(object)

	object = pool.Get()
	require.Equal(t, 0, object.State)
	require.Equal(t, 1, objectsCount)
}
