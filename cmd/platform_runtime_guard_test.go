package cmd

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

func TestProductionCodeDoesNotUseAggregateBaseClient(t *testing.T) {
	repoRoot := ".."
	var violations []string
	err := filepath.WalkDir(repoRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "docs", "vendor":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		rel, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}
		if strings.HasPrefix(rel, "ux/") {
			return nil
		}

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		inBasePackage := strings.HasPrefix(rel, "base/")
		ast.Inspect(file, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.SelectorExpr:
				if ident, ok := x.X.(*ast.Ident); ok && ident.Name == "base" && x.Sel.Name == "BizClient" {
					violations = append(violations, fset.Position(x.Pos()).String()+": base.BizClient is forbidden")
				}
			case *ast.TypeSpec:
				if inBasePackage && x.Name.Name == "Client" {
					violations = append(violations, fset.Position(x.Pos()).String()+": base.Client aggregate type is forbidden")
				}
			case *ast.ValueSpec:
				if inBasePackage {
					for _, name := range x.Names {
						if name.Name == "BizClient" {
							violations = append(violations, fset.Position(name.Pos()).String()+": base.BizClient global is forbidden")
						}
					}
				}
			case *ast.FuncDecl:
				if inBasePackage && (x.Name.Name == "NewClient" || x.Name.Name == "GetBizClient") {
					violations = append(violations, fset.Position(x.Pos()).String()+": aggregate client constructor is forbidden")
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) > 0 {
		t.Fatalf("aggregate base client usage remains:\n%s", strings.Join(violations, "\n"))
	}
}
