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
//  6. Reserved command names: no products.yaml entry may declare a top-level
//     command name that the platform itself registers (see reservedCommands).
//     A product declaring e.g. "config" would silently shadow the platform
//     command, so it is a VIOLATION.
//  7. Cross-product command uniqueness: no two enabled products may declare the
//     same top-level command name (cobra AddCommand silently shadows duplicates).
//  8. Commands consistency: each enabled product's product.go Metadata().Commands
//     must match its products.yaml `commands` (order-independent).
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
	"init":      true, // NewCmdInit
	"auth":      true, // NewCmdAuth
	"gendoc":    true, // NewCmdDoc (doc-gen command, Use: "gendoc")
	"config":    true, // NewCmdConfig
	"region":    true, // NewCmdRegion
	"project":   true, // NewCmdProject
	// uhost migrated to products/uhost (Part 6) — no longer platform-reserved.
	"subnet":    true, // NewCmdSubnet
	"vpc":       true, // NewCmdVpc
	"bandwidth": true, // NewCmdBandwidth
	"udpn":      true, // NewCmdUDPN
	"ulb":       true, // NewCmdULB
	"gssh":      true, // NewCmdGssh
	"pathx":     true, // NewCmdPathx
	"redis":     true, // NewCmdRedis
	"memcache":  true, // NewCmdMemcache
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
