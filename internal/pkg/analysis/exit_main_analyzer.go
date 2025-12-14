package analysis

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var ExitMainAnalyzer = &analysis.Analyzer{
	Name: "exitMain",
	Doc:  "check for os.Exit or log.Fatal outside of main.main",
	Run:  runExitMainAnalysis,
}

func runExitMainAnalysis(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			if funcDecl, ok := node.(*ast.FuncDecl); ok {
				// Объявлена какая-то функция.

				if funcDecl.Name.Name == "main" && pass.Pkg.Name() == "main" {
					// Это `main()` из пакета `main`, для нее нет ограничений.
					// Заходить в потомки не нужно.
					return false
				}

				// Поищем в функции проблемные вызовы.
				ast.Inspect(funcDecl, func(node ast.Node) bool {
					if call, ok := node.(*ast.CallExpr); ok {
						if selector, ok := call.Fun.(*ast.SelectorExpr); ok {
							if isOsExit(selector) {
								pass.Reportf(selector.Pos(), "os.Exit is allowed only inside of main.main")
								return false
							}

							if isLogFatal(selector) {
								pass.Reportf(selector.Pos(), "log.Fatal is allowed only inside of main.main")
								return false
							}
						}
					}

					return true
				})
			}

			// Ищем дальше.
			return true
		})
	}

	return nil, nil
}

func isOsExit(selector *ast.SelectorExpr) bool {
	if pkgIdent, ok := selector.X.(*ast.Ident); ok {
		if pkgIdent.Name == "os" && selector.Sel.Name == "Exit" {
			return true
		}
	}

	return false
}

func isLogFatal(selector *ast.SelectorExpr) bool {
	if pkgIdent, ok := selector.X.(*ast.Ident); ok {
		if pkgIdent.Name == "log" && selector.Sel.Name == "Fatal" {
			return true
		}
	}

	return false
}
