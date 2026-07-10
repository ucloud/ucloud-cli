package cmd

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const guardModuleRoot = "github.com/ucloud/ucloud-cli"

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
		inPlatformPackage := strings.HasPrefix(rel, "cmd/internal/platform/")
		ast.Inspect(file, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.SelectorExpr:
				if ident, ok := x.X.(*ast.Ident); ok && ident.Name == "base" && x.Sel.Name == "BizClient" {
					violations = append(violations, fset.Position(x.Pos()).String()+": platform.BizClient is forbidden")
				}
			case *ast.TypeSpec:
				if inPlatformPackage && x.Name.Name == "Client" {
					violations = append(violations, fset.Position(x.Pos()).String()+": platform.Client aggregate type is forbidden")
				}
			case *ast.ValueSpec:
				if inPlatformPackage {
					for _, name := range x.Names {
						if name.Name == "BizClient" {
							violations = append(violations, fset.Position(name.Pos()).String()+": platform.BizClient global is forbidden")
						}
					}
				}
			case *ast.FuncDecl:
				if inPlatformPackage && (x.Name.Name == "NewClient" || x.Name.Name == "GetBizClient") {
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

func TestProductionCodeDoesNotImportLegacyTopLevelPackages(t *testing.T) {
	repoRoot := ".."
	forbidden := map[string]string{
		guardModuleRoot + "/base": "legacy top-level base package is forbidden",
		guardModuleRoot + "/ux":   "legacy top-level ux package is forbidden",
		guardModuleRoot + "/ansi": "legacy top-level ansi package is forbidden",
	}

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

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imp := range file.Imports {
			importPath, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				return err
			}
			if msg, ok := forbidden[importPath]; ok {
				violations = append(violations, fset.Position(imp.Path.Pos()).String()+": "+msg)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) > 0 {
		t.Fatalf("legacy top-level imports remain:\n%s", strings.Join(violations, "\n"))
	}
}

func TestLegacyTopLevelPackageDirectoriesDoNotExist(t *testing.T) {
	for _, dir := range []string{"../base", "../ux", "../ansi"} {
		if _, err := os.Stat(dir); err == nil {
			t.Fatalf("legacy top-level package directory still exists: %s", dir)
		}
	}
}
