// hack/check-product enforces product-module boundaries by statically
// analysing the products/ tree.
//
// Run from repo root:
//
//	go run ./hack/check-product
//
// Exit 0 means all rules passed. Exit non-zero means one or more violations
// were printed to stdout.
//
// # Design note
//
// This tool uses the standard library (go/parser + go/ast + go/token) rather
// than golang.org/x/tools/go/packages, so it adds NO new module dependency.
// Import-level rules are fully covered by the AST; type resolution is not
// needed for the patterns we flag.
//
// # Rules
//
//  1. No cross-product imports: a file under products/A/... must not import
//     github.com/ucloud/ucloud-cli/products/B (for any B != A).
//  2. No cmd or base imports: product files must not import
//     github.com/ucloud/ucloud-cli/cmd or .../base.
//  3. No bare SDK NewClient calls (best-effort AST): flag svc.NewClient(...)
//     where svc is not the identifier "cli" (products must use
//     cli.NewServiceClient).
//  4. No raw completion API calls: flag selector calls whose method name is
//     SetFlagValuesFunc, GetFlagValuesFunc, GetFlagValues, or SetFlagValues
//     (when the receiver is not "command").
//  5. products.yaml consistency: every enabled product whose dir is absent on
//     disk emits a WARNING (not a failure). Every directory under products/
//     that has no entry in products.yaml is a VIOLATION.
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

const moduleRoot = "github.com/ucloud/ucloud-cli"

// Product mirrors the products.yaml entry.
type Product struct {
	Name     string   `yaml:"name"`
	Dir      string   `yaml:"dir"`
	Owners   []string `yaml:"owners"`
	Commands []string `yaml:"commands"`
	Enabled  bool     `yaml:"enabled"`
}

// registry is the top-level products.yaml structure.
type registry struct {
	Products []Product `yaml:"products"`
}

// rawCompletion methods that must go through command.*; flagging them is
// best-effort — we match the selector method name.
var rawCompletionMethods = map[string]bool{
	"SetFlagValuesFunc": true,
	"GetFlagValuesFunc": true,
	"GetFlagValues":     true,
	// SetFlagValues is only flagged when the receiver is NOT "command".
	"SetFlagValues": true,
}

// checkFile parses the Go file at path (which lives under
// products/<productName>/…) and returns one string per violation.
//
// productName is the immediate subdirectory name under products/
// (e.g. "udb"), used to distinguish intra-product imports from cross-product
// imports.
func checkFile(path, productName string) []string {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return []string{fmt.Sprintf("%s: parse error: %v", path, err)}
	}

	var violations []string

	pos := func(node ast.Node) string {
		p := fset.Position(node.Pos())
		return fmt.Sprintf("%s:%d", p.Filename, p.Line)
	}

	// ---- Rule 1 & 2: import paths ----------------------------------------
	for _, imp := range f.Imports {
		if imp.Path == nil {
			continue
		}
		// Strip surrounding quotes.
		importPath := strings.Trim(imp.Path.Value, `"`)

		// Rule 2: no cmd or base imports.
		if importPath == moduleRoot+"/cmd" ||
			strings.HasPrefix(importPath, moduleRoot+"/cmd/") {
			violations = append(violations,
				fmt.Sprintf("%s: rule2: product must not import cmd package %q",
					pos(imp.Path), importPath))
		}
		if importPath == moduleRoot+"/base" ||
			strings.HasPrefix(importPath, moduleRoot+"/base/") {
			violations = append(violations,
				fmt.Sprintf("%s: rule2: product must not import base package %q",
					pos(imp.Path), importPath))
		}

		// Rule 1: no cross-product imports.
		productsPrefix := moduleRoot + "/products/"
		if strings.HasPrefix(importPath, productsPrefix) {
			// e.g. "github.com/ucloud/ucloud-cli/products/vpc/something"
			// → rest = "vpc/something"
			rest := strings.TrimPrefix(importPath, productsPrefix)
			// other product = everything before the first "/"
			otherProduct := rest
			if idx := strings.Index(rest, "/"); idx >= 0 {
				otherProduct = rest[:idx]
			}
			if otherProduct != productName {
				violations = append(violations,
					fmt.Sprintf("%s: rule1: product %q must not import sibling product %q (import %q)",
						pos(imp.Path), productName, otherProduct, importPath))
			}
		}
	}

	// ---- Rules 3 & 4: AST call-expression walk ----------------------------
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		recv, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}
		method := sel.Sel.Name
		receiver := recv.Name

		// Rule 3: bare SDK NewClient calls.
		// Flag svc.NewClient(...) unless svc == "cli".
		if method == "NewClient" && receiver != "cli" {
			violations = append(violations,
				fmt.Sprintf("%s: rule3: direct %s.NewClient() call; use cli.NewServiceClient(ctx, %s.NewClient) instead",
					pos(call), receiver, receiver))
		}

		// Rule 4: raw completion API.
		if rawCompletionMethods[method] {
			// SetFlagValues is only forbidden when NOT called as command.SetFlagValues.
			if method == "SetFlagValues" && receiver == "command" {
				return true
			}
			// All other raw-completion methods are forbidden unconditionally.
			violations = append(violations,
				fmt.Sprintf("%s: rule4: raw completion call %s.%s(); use command.* wrappers instead",
					pos(call), receiver, method))
		}

		return true
	})

	return violations
}

// checkConsistency verifies that:
//   - every enabled product in yamlProducts whose dir IS PRESENT on disk has a
//     matching directory under products/ (and vice-versa).
//   - every directory under products/ corresponds to at least one products.yaml
//     entry (any enablement state).
//
// An enabled product whose dir is ABSENT on disk emits a warning (not a
// violation) — this is the expected pre-F state.
//
// dirs is the list of immediate subdirectory names under products/.
//
// Returns (violations, warnings).
func checkConsistency(yamlProducts []Product, dirs []string) (violations, warnings []string) {
	// Build sets.
	dirSet := make(map[string]bool, len(dirs))
	for _, d := range dirs {
		dirSet[d] = true
	}

	// Build a set of dirs claimed by any products.yaml entry (not just enabled).
	yamlDirNames := make(map[string]bool, len(yamlProducts))
	for _, p := range yamlProducts {
		// p.Dir is "products/<name>"; we want the leaf name.
		leaf := filepath.Base(p.Dir)
		yamlDirNames[leaf] = true

		// Only warn (not fail) if an enabled product's dir is missing.
		if p.Enabled && !dirSet[leaf] {
			warnings = append(warnings,
				fmt.Sprintf("warn: enabled product %q dir %q not found on disk (pre-F state, not a failure)",
					p.Name, p.Dir))
		}
	}

	// Every directory under products/ must have a products.yaml entry.
	for _, d := range dirs {
		if !yamlDirNames[d] {
			violations = append(violations,
				fmt.Sprintf("products/%s: rule5: directory has no matching entry in products.yaml", d))
		}
	}

	return violations, warnings
}

func main() {
	var allViolations []string
	var allWarnings []string

	// ---- Load products.yaml -----------------------------------------------
	yamlPath := "products.yaml"
	raw, err := os.ReadFile(yamlPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "check-product: cannot read %s: %v\n", yamlPath, err)
		os.Exit(2)
	}

	var reg registry
	if err := yaml.Unmarshal(raw, &reg); err != nil {
		fmt.Fprintf(os.Stderr, "check-product: parse %s: %v\n", yamlPath, err)
		os.Exit(2)
	}

	// ---- Discover products/ dirs ------------------------------------------
	productsRoot := "products"
	var foundDirs []string

	entries, err := os.ReadDir(productsRoot)
	if err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "check-product: read products/: %v\n", err)
		os.Exit(2)
	}
	// If products/ doesn't exist, entries is nil/empty → graceful.
	for _, e := range entries {
		if e.IsDir() {
			foundDirs = append(foundDirs, e.Name())
		}
	}

	// ---- Rule 5: consistency check ----------------------------------------
	v5, w5 := checkConsistency(reg.Products, foundDirs)
	allViolations = append(allViolations, v5...)
	allWarnings = append(allWarnings, w5...)

	// ---- Rules 1–4: walk every .go file under products/ ------------------
	if _, statErr := os.Stat(productsRoot); statErr == nil {
		walkErr := filepath.Walk(productsRoot, func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if info.IsDir() || !strings.HasSuffix(path, ".go") {
				return nil
			}

			// Determine which product this file belongs to.
			// path is like "products/udb/internal/mysql/cmd.go"
			rel := strings.TrimPrefix(path, productsRoot+string(filepath.Separator))
			parts := strings.SplitN(rel, string(filepath.Separator), 2)
			productName := parts[0]

			allViolations = append(allViolations, checkFile(path, productName)...)
			return nil
		})
		if walkErr != nil {
			fmt.Fprintf(os.Stderr, "check-product: walk products/: %v\n", walkErr)
			os.Exit(2)
		}
	}

	// ---- Report ------------------------------------------------------------
	for _, w := range allWarnings {
		fmt.Println(w)
	}

	if len(allViolations) == 0 {
		fmt.Println("check-product: all boundary rules passed.")
		os.Exit(0)
	}

	for _, v := range allViolations {
		fmt.Println(v)
	}
	fmt.Fprintf(os.Stderr, "check-product: %d violation(s) found.\n", len(allViolations))
	os.Exit(1)
}
