package resetter

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_GenerateStructReset(t *testing.T) {
	t.Parallel()

	type want struct {
		pkg    string
		source string
	}

	tests := []struct {
		name      string
		source    string
		want      want
		wantError bool
	}{
		{
			name: "WHEN not a type THEN error",
			source: `
						package resetter

						// generate:reset
						var Answer int = 42
					`,
			wantError: true,
		},
		{
			name: "WHEN not a struct THEN error",
			source: `
						package resetter

						// generate:reset
						type Integer int
					`,
			wantError: true,
		},
		{
			name: "WHEN no comment THEN no empty response",
			source: `
				package resetter

				type NotResetableStruct struct {
					i int
				}
			`,
		},
		{
			name: "WHEN valid code THEN ok",
			source: `
						package resetter

						// generate:reset
						type ResetableStruct struct {
							i     int
							str   string
							strP  *string
							s     []int
							m     map[string]string
							child *ResetableStruct
						}
					`,
			want: want{
				pkg: "resetter",
				source: `func (rs *ResetableStruct) Reset() {
	if rs == nil {
		return
	}

	rs.i = 0
	rs.str = ""
	if rs.strP != nil {
		*rs.strP = ""
	}
	rs.s = rs.s[:0]
	clear(rs.m)
	if resetter, ok := any(rs.child).(interface{ Reset() }); ok && rs.child != nil {
		resetter.Reset()
	}
}`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, "", tt.source, parser.AllErrors|parser.ParseComments)
			require.NoError(t, err)

			// Act.
			pkg, source, err := GenerateStructReset(f)

			// Assert.
			require.Equal(t, tt.want.pkg, pkg)
			require.Equal(t, tt.want.source, source)
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
