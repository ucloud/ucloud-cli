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
//  2. No platform-internal or legacy imports: product files must not import
//     github.com/ucloud/ucloud-cli/cmd, .../base, .../ux, or .../ansi.
//  3. No bare SDK NewClient calls (best-effort AST): flag svc.NewClient(...)
//     where svc is not the identifier "cli" (products must use
//     cli.NewServiceClient).
//  4. No raw completion API calls: flag selector calls whose method name is
//     SetFlagValuesFunc, GetFlagValuesFunc, GetFlagValues, or SetFlagValues
//     (when the receiver is not "command").
//  5. product.yaml consistency: every enabled product whose dir is absent on
//     disk emits a WARNING (not a failure). Every directory under products/
//     that has no product.yaml is a VIOLATION.
//  6. Reserved command names: no product.yaml may declare a top-level
//     command name that the platform itself registers (see reservedCommands).
//     A product declaring e.g. "config" would silently shadow the platform
//     command, so it is a VIOLATION.
//  7. Cross-product command uniqueness: no two enabled products may declare the
//     same top-level command name (cobra AddCommand silently shadows duplicates).
//  8. Commands consistency: each enabled product's product.go Metadata().Commands
//     must match its product.yaml `commands` (order-independent).
//  9. §6.1 import whitelist: product files may import only stdlib,
//     ucloud-sdk-go, spf13/cobra|pflag, pkg/cli|command|ui, internal/common,
//     and their own product subtree. Anything else (model/*, ux/, new
//     third-party deps) is a violation; extending the list is a platform PR.
//  10. §2 file layout: (a) grab-bag basenames (helpers.go, utils.go, util.go,
//     common.go, misc.go, and their _test.go variants) are forbidden under
//     products/ — name files by verb or concern; (b) each non-test .go file
//     may declare at most one top-level function (methods included) returning
//     *cobra.Command: one verb per file.
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

const moduleRoot = "github.com/ucloud/ucloud-cli"

// Product mirrors the products.yaml entry.
type Product struct {
	Name     string   `yaml:"name"`
	Dir      string   `yaml:"-"` // 从 product.yaml 路径推断,不从文件读
	Owners   []string `yaml:"owners"`
	Commands []string `yaml:"commands"`
	Enabled  bool     `yaml:"enabled"`
}

// loadProducts scans products/*/product.yaml and returns the products in
// deterministic (path-sorted) order. Dir is derived from each file's directory.
func loadProducts() ([]Product, error) {
	matches, err := filepath.Glob("products/*/product.yaml")
	if err != nil {
		return nil, err
	}
	sort.Strings(matches)
	var products []Product
	for _, path := range matches {
		raw, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil, fmt.Errorf("read %s: %w", path, readErr)
		}
		var p Product
		if uErr := yaml.Unmarshal(raw, &p); uErr != nil {
			return nil, fmt.Errorf("parse %s: %w", path, uErr)
		}
		p.Dir = filepath.Dir(path)
		products = append(products, p)
	}
	return products, nil
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

// ---- Rule 9: §6.1 product import whitelist -------------------------------
// A product file may import ONLY: the Go standard library; the platform
// contract packages pkg/cli, pkg/command, pkg/ui; internal/common
// (domain-agnostic pure tools, open to products); the UCloud SDK (incl.
// private/); the cobra flag stack; and its own product subtree. Everything
// else — other module-internal packages (model/*, ux/, ...) and any NEW
// third-party dependency — is a violation. Extending this list is a platform
// PR by design: it is the gate that keeps go.mod out of product PRs.
//
// allowedModulePackages: exact import paths — subpackages need their own
// entry (platform PR).
var allowedModulePackages = map[string]bool{
	moduleRoot + "/pkg/cli":         true,
	moduleRoot + "/pkg/command":     true,
	moduleRoot + "/pkg/ui":          true,
	moduleRoot + "/internal/common": true,
}

// allowedThirdParty: prefix match — subpackages allowed.
var allowedThirdParty = []string{
	"github.com/ucloud/ucloud-sdk-go",
	"github.com/spf13/cobra",
	"github.com/spf13/pflag",
}

// importAllowed reports whether importPath is inside the §6.1 whitelist for
// files belonging to productName.
func importAllowed(importPath, productName string) bool {
	first := importPath
	if idx := strings.Index(importPath, "/"); idx >= 0 {
		first = importPath[:idx]
	}
	if !strings.Contains(first, ".") {
		return true // standard library
	}
	if allowedModulePackages[importPath] {
		return true
	}
	self := moduleRoot + "/products/" + productName
	if importPath == self || strings.HasPrefix(importPath, self+"/") {
		return true
	}
	for _, p := range allowedThirdParty {
		if importPath == p || strings.HasPrefix(importPath, p+"/") {
			return true
		}
	}
	return false
}

// ---- Rule 10a: forbidden grab-bag filenames --------------------------------
// §2 names product files by verb or concern (list.go, rows.go, completion.go,
// describe.go/poll.go, status.go, ...). Grab-bag basenames defeat that layout,
// so they are forbidden under products/ — including their _test.go variants
// (a grab-bag test file is the same smell).
var grabBagBasenames = map[string]bool{
	"helpers": true,
	"utils":   true,
	"util":    true,
	"common":  true,
	"misc":    true,
}

// checkFilename enforces rule 10a on a single path under products/. It is a
// pure path check (no parsing) kept separate from checkFile so it can be
// unit-tested with arbitrary paths.
func checkFilename(path string) []string {
	base := filepath.Base(path)
	if !strings.HasSuffix(base, ".go") {
		return nil
	}
	stem := strings.TrimSuffix(base, ".go")
	stem = strings.TrimSuffix(stem, "_test")
	if grabBagBasenames[stem] {
		return []string{fmt.Sprintf(
			"rule10: grab-bag filename %q is forbidden under products/ (name files by verb or concern: <verb>.go, rows.go, completion.go, describe.go/poll.go, status.go)",
			path)}
	}
	return nil
}

// ---- Rule 10b: one cobra constructor per file -------------------------------

// returnsCobraCommand reports whether the function signature's result list
// includes a *cobra.Command. Matching is by AST shape — a StarExpr over a
// SelectorExpr whose Sel is "Command" — regardless of the import alias or
// package qualifier: the only *X.Command pointer type used in this codebase
// is cobra's (verified by grep across cmd/, products/, pkg/, internal/, base/,
// ux/, model/), so selector-name matching is sufficient and alias-proof.
func returnsCobraCommand(ft *ast.FuncType) bool {
	if ft == nil || ft.Results == nil {
		return false
	}
	for _, field := range ft.Results.List {
		star, ok := field.Type.(*ast.StarExpr)
		if !ok {
			continue
		}
		if sel, ok := star.X.(*ast.SelectorExpr); ok && sel.Sel.Name == "Command" {
			return true
		}
	}
	return false
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

	// ---- Rules 1, 2 & 9: import paths --------------------------------------
	productsPrefix := moduleRoot + "/products/"
	for _, imp := range f.Imports {
		if imp.Path == nil {
			continue
		}
		// Strip surrounding quotes.
		importPath := strings.Trim(imp.Path.Value, `"`)

		// Shared territory booleans: rules 1-2 and rule 9's carve-out must
		// stay in sync, so classify each import path exactly once.
		isCmdImport := importPath == moduleRoot+"/cmd" ||
			strings.HasPrefix(importPath, moduleRoot+"/cmd/")
		legacyPlatformPackage := ""
		for _, name := range []string{"base", "ux", "ansi"} {
			prefix := moduleRoot + "/" + name
			if importPath == prefix || strings.HasPrefix(importPath, prefix+"/") {
				legacyPlatformPackage = name
				break
			}
		}
		isProductsImport := strings.HasPrefix(importPath, productsPrefix)

		// Rule 2: no platform-internal or legacy imports.
		if isCmdImport {
			violations = append(violations,
				fmt.Sprintf("%s: rule2: product must not import cmd package %q",
					pos(imp.Path), importPath))
		}
		if legacyPlatformPackage != "" {
			violations = append(violations,
				fmt.Sprintf("%s: rule2: product must not import legacy %s package %q",
					pos(imp.Path), legacyPlatformPackage, importPath))
		}

		// Rule 1: no cross-product imports.
		if isProductsImport {
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

		// Rule 9: §6.1 whitelist. cmd, legacy platform, and products/ prefixes are owned by
		// rules 1-2 above (more specific messages) — rule 9 covers the rest.
		isRule12Territory := isCmdImport || legacyPlatformPackage != "" || isProductsImport
		if !isRule12Territory && !importAllowed(importPath, productName) {
			violations = append(violations,
				fmt.Sprintf("%s: rule9: import %q is outside the §6.1 product import whitelist (allowed: stdlib, ucloud-sdk-go, spf13/cobra|pflag, pkg/cli|command|ui, internal/common, own product); extending the whitelist is a platform PR",
					pos(imp.Path), importPath))
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

	// ---- Rule 10b: at most one cobra constructor per file ------------------
	// §2「one verb per file」: a non-test file may declare at most one
	// top-level function whose results include *cobra.Command. Methods count
	// too (a method returning *cobra.Command is still a constructor); FuncLit
	// closures inside bodies are naturally excluded because only f.Decls is
	// scanned.
	if !strings.HasSuffix(filepath.Base(path), "_test.go") {
		constructors := 0
		for _, decl := range f.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok && returnsCobraCommand(fn.Type) {
				constructors++
			}
		}
		if constructors > 1 {
			violations = append(violations,
				fmt.Sprintf("rule10: %s declares %d cobra-constructor functions; §2 allows at most one per file (one verb per file; move extras to their own <verb>.go)",
					path, constructors))
		}
	}

	return violations
}

// checkConsistency (rule5): every immediate subdirectory of products/ must have
// a product.yaml (i.e. appear in the scanned products list).
func checkConsistency(products []Product, dirs []string) (violations, warnings []string) {
	hasYAML := make(map[string]bool, len(products))
	for _, p := range products {
		hasYAML[filepath.Base(p.Dir)] = true
	}
	for _, d := range dirs {
		if !hasYAML[d] {
			violations = append(violations,
				fmt.Sprintf("products/%s: rule5: directory has no product.yaml", d))
		}
	}
	return violations, warnings
}

// reservedCommands is the set of PLATFORM-RESERVED top-level command names.
// A product must not declare any of these in its products.yaml `commands`,
// because doing so would silently shadow the platform's own command.
//
// MUST track the platform commands registered in cmd/root.go:
// addPlatformCommands (the root.AddCommand(NewCmd*()) calls and newSchemaCmd),
// plus the root-level pseudo-commands wired in NewCmdRoot (completion, signup).
// When a platform command is added/renamed/removed in cmd/root.go, update this
// set to match.
var reservedCommands = map[string]bool{
	// addPlatformCommands (cmd/root.go), in registration order:
	"init":    true, // NewCmdInit
	"auth":    true, // NewCmdAuth
	"gendoc":  true, // NewCmdDoc (doc-gen command, Use: "gendoc")
	"config":  true, // NewCmdConfig
	"region":  true, // NewCmdRegion
	"project": true, // NewCmdProject
	// uhost migrated to products/uhost (Part 6) — no longer platform-reserved.
	"ext":       true, // NewCmdExt
	"api":       true, // NewCmdAPI
	"signature": true, // NewCmdSignature
	"__schema":  true, // newSchemaCmd (hidden)
	// Root-level pseudo-commands wired in NewCmdRoot (cmd/root.go):
	"completion": true, // NewCmdCompletion
	"signup":     true, // NewCmdSignup
}

// checkReservedCommands verifies that no product declares a top-level command
// name that collides with a platform-reserved name (see reservedCommands).
// Returns one violation string per offending (product, command) pair.
func checkReservedCommands(yamlProducts []Product) []string {
	var violations []string
	for _, p := range yamlProducts {
		for _, c := range p.Commands {
			if reservedCommands[c] {
				violations = append(violations,
					fmt.Sprintf("rule6: product %q declares reserved platform command %q", p.Name, c))
			}
		}
	}
	return violations
}

// checkCommandCollisions verifies that no two ENABLED products declare the same
// top-level command name (设计 §6.3「一级命令无冲突」, rule7). cobra's AddCommand
// does not panic on duplicate names — the later registration silently shadows the
// earlier — so this static gate is the only thing that catches a collision once
// multiple products live under products/.
func checkCommandCollisions(yamlProducts []Product) []string {
	// command name -> products declaring it (enabled only)
	owners := make(map[string][]string)
	for _, p := range yamlProducts {
		if !p.Enabled {
			continue
		}
		for _, c := range p.Commands {
			owners[c] = append(owners[c], p.Name)
		}
	}

	// Deterministic output: iterate command names in sorted order.
	cmds := make([]string, 0, len(owners))
	for c := range owners {
		cmds = append(cmds, c)
	}
	sort.Strings(cmds)

	var violations []string
	for _, c := range cmds {
		if len(owners[c]) > 1 {
			violations = append(violations,
				fmt.Sprintf("rule7: top-level command %q declared by multiple products %v (command names must be unique across products)",
					c, owners[c]))
		}
	}
	return violations
}

func main() {
	var allViolations []string
	var allWarnings []string

	// ---- Load products from products/*/product.yaml -----------------------
	products, err := loadProducts()
	if err != nil {
		fmt.Fprintf(os.Stderr, "check-product: load products: %v\n", err)
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
	v5, w5 := checkConsistency(products, foundDirs)
	allViolations = append(allViolations, v5...)
	allWarnings = append(allWarnings, w5...)

	// ---- Rule 6: reserved platform command names --------------------------
	allViolations = append(allViolations, checkReservedCommands(products)...)

	// ---- Rule 7: cross-product command-name uniqueness --------------------
	allViolations = append(allViolations, checkCommandCollisions(products)...)

	// ---- Rule 8: product.go Metadata commands ↔ products.yaml -------------
	allViolations = append(allViolations, checkCommandsConsistency(products)...)

	// ---- Rules 1–4 & 10: walk every .go file under products/ --------------
	if _, statErr := os.Stat(productsRoot); statErr == nil {
		walkErr := filepath.Walk(productsRoot, func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if info.IsDir() || !strings.HasSuffix(path, ".go") {
				return nil
			}

			// Determine which product this file belongs to.
			// path is like "products/mysql/internal/mysql/cmd.go"
			rel := strings.TrimPrefix(path, productsRoot+string(filepath.Separator))
			parts := strings.SplitN(rel, string(filepath.Separator), 2)
			productName := parts[0]

			// Rule 10a is a per-path check; no parsing needed.
			allViolations = append(allViolations, checkFilename(path)...)
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

// extractMetadataCommands statically extracts the string slice passed as the
// Commands field of the cli.Metadata composite literal returned by the
// Metadata() method in <productDir>/product.go. Returns (nil, nil) if there is
// no Metadata method or no Commands field (设计 §6.3「命令声明与 products.yaml 一致」).
func extractMetadataCommands(productDir string) ([]string, error) {
	path := filepath.Join(productDir, "product.go")
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return nil, err
	}

	var cmds []string
	found := false
	ast.Inspect(f, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Recv == nil || fn.Name.Name != "Metadata" || fn.Body == nil {
			return true
		}
		ast.Inspect(fn.Body, func(m ast.Node) bool {
			kv, ok := m.(*ast.KeyValueExpr)
			if !ok {
				return true
			}
			key, ok := kv.Key.(*ast.Ident)
			if !ok || key.Name != "Commands" {
				return true
			}
			lit, ok := kv.Value.(*ast.CompositeLit)
			if !ok {
				return true
			}
			for _, elt := range lit.Elts {
				if bl, ok := elt.(*ast.BasicLit); ok && bl.Kind == token.STRING {
					if s, uerr := strconv.Unquote(bl.Value); uerr == nil {
						cmds = append(cmds, s)
					}
				}
			}
			found = true
			return false
		})
		return true
	})
	if !found {
		return nil, nil
	}
	return cmds, nil
}

// sameStringSet reports whether a and b contain the same elements (order-independent).
func sameStringSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	seen := make(map[string]int, len(a))
	for _, s := range a {
		seen[s]++
	}
	for _, s := range b {
		seen[s]--
	}
	for _, n := range seen {
		if n != 0 {
			return false
		}
	}
	return true
}

// checkCommandsConsistency verifies that each enabled product whose dir exists
// on disk declares the same Commands in product.go's Metadata() as in
// products.yaml (设计 §6.3, rule8). Skips products whose dir is not yet present
// (pre-F state). Order-independent comparison.
func checkCommandsConsistency(yamlProducts []Product) []string {
	var violations []string
	for _, p := range yamlProducts {
		if !p.Enabled {
			continue
		}
		if _, statErr := os.Stat(p.Dir); os.IsNotExist(statErr) {
			continue
		}
		meta, err := extractMetadataCommands(p.Dir)
		if err != nil {
			violations = append(violations,
				fmt.Sprintf("rule8: %s: cannot parse product.go: %v", p.Dir, err))
			continue
		}
		if meta == nil {
			violations = append(violations,
				fmt.Sprintf("rule8: product %q: product.go has no Metadata().Commands to verify against products.yaml", p.Name))
			continue
		}
		if !sameStringSet(p.Commands, meta) {
			violations = append(violations,
				fmt.Sprintf("rule8: product %q commands mismatch: products.yaml=%v vs product.go Metadata()=%v",
					p.Name, p.Commands, meta))
		}
	}
	return violations
}
