package main

import (
	"github.com/aleffnull/shortener/internal/pkg/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	multichecker.Main(analysis.PanicAnalyzer, analysis.ExitMainAnalyzer)
}
