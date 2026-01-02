package resetter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestObjectWithState_Reset(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		hookBefore func() *ResetableStruct
		hookAfter  func(obj *ResetableStruct)
	}{
		{
			name: "WHEN nil THEN ok",
			hookBefore: func() *ResetableStruct {
				return nil
			},
			hookAfter: func(obj *ResetableStruct) {},
		},
		{
			name: "WHEN not nil THEN fields reset",
			hookBefore: func() *ResetableStruct {
				strToPoint := "bar"
				return &ResetableStruct{
					i:    42,
					str:  "foo",
					strP: &strToPoint,
					s:    []int{1, 2, 3},
					m:    map[string]string{"one": "two"},
					child: &ResetableStruct{
						i: 84,
					},
				}
			},
			hookAfter: func(obj *ResetableStruct) {
				require.Zero(t, obj.i)
				require.Len(t, obj.str, 0)
				require.Len(t, *obj.strP, 0)
				require.Len(t, obj.s, 0)
				require.Len(t, obj.m, 0)
				require.Zero(t, obj.child.i)
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
