package analysis

import (
	"path"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func Test_panicAnalyzer(t *testing.T) {
	analysistest.Run(t, path.Join(analysistest.TestData(), "panic_pkg"), PanicAnalyzer, "./...")
}
