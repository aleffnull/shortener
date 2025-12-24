package analysis

import (
	"path"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func Test_exitMainAnalyzer(t *testing.T) {
	analysistest.Run(t, path.Join(analysistest.TestData(), "exit_main_pkg"), ExitMainAnalyzer, "./...")
}
