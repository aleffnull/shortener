package resetter

import (
	"bytes"
	"fmt"
	"go/ast"
	"strings"
	"text/template"
)

type structData struct {
	StructName          string
	IntFields           []string
	StringFields        []string
	SliceFields         []string
	MapFields           []string
	StringPointerFields []string
	TypePointerFields   []string
}

const structResetTemplate = `func (rs *{{.StructName}}) Reset() {
	if rs == nil {
		return
	}

{{range .IntFields}}	rs.{{.}} = 0
{{end}}{{range .StringFields}}	rs.{{.}} = ""
{{end}}{{range .StringPointerFields}}	if rs.{{.}} != nil {
		*rs.{{.}} = ""
	}
{{end}}{{range .SliceFields}}	rs.{{.}} = rs.{{.}}[:0]
{{end}}{{range .MapFields}}	clear(rs.{{.}})
{{end}}{{range .TypePointerFields}}	if resetter, ok := any(rs.{{.}}).(interface{ Reset() }); ok && rs.child != nil {
		resetter.Reset()
	}
{{end}}}`

func GenerateStructReset(file *ast.File) (string, string, error) {
	data := []structData{}

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Doc.Text() != "generate:reset\n" {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				return "", "", fmt.Errorf("expected a type, but got %T", spec)
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				return "", "", fmt.Errorf("expected a struct, but got %T", spec)
			}

			datum := structData{
				StructName: typeSpec.Name.Name,
			}

			for _, field := range structType.Fields.List {
				switch t := field.Type.(type) {
				case *ast.Ident:
					for _, ident := range field.Names {
						switch t.Name {
						case "int":
							datum.IntFields = append(datum.IntFields, ident.Name)
						case "string":
							datum.StringFields = append(datum.StringFields, ident.Name)
						}
					}
				case *ast.StarExpr:
					switch t := t.X.(type) {
					case *ast.Ident:
						for _, ident := range field.Names {
							switch t.Name {
							case "string":
								datum.StringPointerFields = append(datum.StringPointerFields, ident.Name)
							default:
								datum.TypePointerFields = append(datum.TypePointerFields, ident.Name)
							}
						}
					}
				case *ast.ArrayType:
					for _, ident := range field.Names {
						datum.SliceFields = append(datum.SliceFields, ident.Name)
					}
				case *ast.MapType:
					for _, ident := range field.Names {
						datum.MapFields = append(datum.MapFields, ident.Name)
					}
				}
			}

			data = append(data, datum)
		}
	}

	if len(data) == 0 {
		return "", "", nil
	}

	tmpl, err := template.New("struct_reset").Parse(structResetTemplate)
	if err != nil {
		return "", "", fmt.Errorf("template parse error: %w", err)
	}

	sb := &strings.Builder{}
	for i := range data {
		buffer := &bytes.Buffer{}
		if err := tmpl.Execute(buffer, data[i]); err != nil {
			return "", "", fmt.Errorf("template executing error: %w", err)
		}

		_, _ = sb.Write(buffer.Bytes())
	}

	return file.Name.Name, sb.String(), nil
}
