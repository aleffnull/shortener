package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/aleffnull/shortener/internal/pkg/resetter"
)

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("Start directory is required")
		return
	}

	startDirectory := os.Args[1]
	fmt.Printf("Starting from directory %v\n", startDirectory)

	err := filepath.Walk(startDirectory, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() || (strings.HasPrefix(path, ".") && path != ".") {
			return nil
		}

		if err := processDirectory(path); err != nil {
			return fmt.Errorf("directory processing error: %w", err)
		}

		return nil
	})
	if err != nil {
		fmt.Printf("Processing error: %v\n", err)
	}
}

func processDirectory(path string) error {
	fmt.Printf("Processing directory %v\n", path)

	files, err := filepath.Glob(filepath.Join(path, "*.go"))
	if err != nil {
		return fmt.Errorf("failed to get files list: %w", err)
	}

	sb := &strings.Builder{}
	pkg := ""
	for _, file := range files {
		fmt.Printf("\t%v\n", file)

		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, file, nil, parser.AllErrors|parser.ParseComments)
		if err != nil {
			return fmt.Errorf("file parsing error: %v", err)
		}

		filePackage, source, err := resetter.GenerateStructReset(f)
		if err != nil {
			return fmt.Errorf("code generation error: %v", err)
		}

		if len(source) > 0 {
			if pkg != "" && pkg != filePackage {
				return fmt.Errorf("files with different packages in directory %v", path)
			}

			pkg = filePackage
			_, _ = sb.WriteString(source)
			_, _ = sb.WriteString("\n")
		}
	}

	if sb.Len() > 0 {
		if err := writeGeneratedFile(path, pkg, sb); err != nil {
			return fmt.Errorf("failed to generate resetting code: %w", err)
		}

		fmt.Println("+++ Reset code generated")
	}

	return nil
}

func writeGeneratedFile(path string, pkg string, sb *strings.Builder) error {
	file, err := os.Create(filepath.Join(path, "reset.gen.go"))
	if err != nil {
		return fmt.Errorf("file creation error: %w", err)
	}

	defer file.Close()

	if _, err := file.WriteString(fmt.Sprintf("package %v\n\n", pkg)); err != nil {
		return fmt.Errorf("failed to write package name to file: %w", err)
	}

	if _, err := file.WriteString(sb.String()); err != nil {
		return fmt.Errorf("failed to write code to file: %w", err)
	}

	return nil
}
