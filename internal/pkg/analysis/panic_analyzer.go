package analysis

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var PanicAnalyzer = &analysis.Analyzer{
	Name: "panicCheck",
	Doc:  "check for panic() usage",
	Run:  runPanicAnalysis,
}

func runPanicAnalysis(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			if call, ok := node.(*ast.CallExpr); ok {
				if ident, ok := call.Fun.(*ast.Ident); ok {
					if ident.Name == "panic" {
						pass.Reportf(ident.Pos(), "panic is prohibited")
						return false
					}
				}
			}
			return true
		})
	}

	return nil, nil
}
