package pool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestObjectWithState_Reset(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		hookBefore func() *ObjectWithState
		hookAfter  func(obj *ObjectWithState)
	}{
		{
			name: "WHEN nil THEN ok",
			hookBefore: func() *ObjectWithState {
				return nil
			},
			hookAfter: func(obj *ObjectWithState) {},
		},
		{
			name: "WHEN not nil THEN state reset",
			hookBefore: func() *ObjectWithState {
				return &ObjectWithState{
					State: 42,
				}
			},
			hookAfter: func(obj *ObjectWithState) {
				require.Zero(t, obj.State)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			obj := tt.hookBefore()

			// Act.
			obj.Reset()

			// Assert.
			tt.hookAfter(obj)
		})
	}
}
