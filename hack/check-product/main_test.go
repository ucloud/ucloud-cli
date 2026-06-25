package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeFile creates a file at dir/name with content, creating parent dirs.
func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return path
}

// --------------------------------------------------------------------------
// checkFile tests
// --------------------------------------------------------------------------

func TestCheckFile_Rule1_CrossProductImport(t *testing.T) {
	dir := t.TempDir()
	src := `package cmd

import (
	"github.com/ucloud/ucloud-cli/products/vpc"
)

var _ = vpc.Foo
`
	path := writeFile(t, dir, "udb/cmd.go", src)
	got := checkFile(path, "udb")
	if len(got) == 0 {
		t.Fatal("expected violation for cross-product import, got none")
	}
	if !strings.Contains(got[0], "rule1") {
		t.Errorf("expected rule1 in violation, got: %v", got)
	}
}

func TestCheckFile_Rule1_SameProductImport_Clean(t *testing.T) {
	dir := t.TempDir()
	// Importing within the same product is allowed.
	src := `package cmd

import (
	_ "github.com/ucloud/ucloud-cli/products/udb/internal/helper"
)
`
	path := writeFile(t, dir, "udb/cmd.go", src)
	got := checkFile(path, "udb")
	for _, v := range got {
		if strings.Contains(v, "rule1") {
			t.Errorf("unexpected rule1 violation for same-product import: %v", v)
		}
	}
}

func TestCheckFile_Rule2_CmdImport(t *testing.T) {
	dir := t.TempDir()
	src := `package cmd

import (
	"github.com/ucloud/ucloud-cli/cmd"
)

var _ = cmd.Root
`
	path := writeFile(t, dir, "udb/cmd.go", src)
	got := checkFile(path, "udb")
	if len(got) == 0 {
		t.Fatal("expected violation for cmd import, got none")
	}
	if !strings.Contains(got[0], "rule2") {
		t.Errorf("expected rule2 in violation, got: %v", got)
	}
}

func TestCheckFile_Rule2_BaseImport(t *testing.T) {
	dir := t.TempDir()
	src := `package cmd

import (
	"github.com/ucloud/ucloud-cli/base"
)

var _ = base.Foo
`
	path := writeFile(t, dir, "udb/cmd.go", src)
	got := checkFile(path, "udb")
	if len(got) == 0 {
		t.Fatal("expected violation for base import, got none")
	}
	if !strings.Contains(got[0], "rule2") {
		t.Errorf("expected rule2 in violation, got: %v", got)
	}
}

func TestCheckFile_Rule3_BareNewClient(t *testing.T) {
	dir := t.TempDir()
	src := `package cmd

import "github.com/ucloud/ucloud-sdk-go/services/udb"

func setup() {
	client := udb.NewClient(nil)
	_ = client
}
`
	path := writeFile(t, dir, "udb/cmd.go", src)
	got := checkFile(path, "udb")
	found := false
	for _, v := range got {
		if strings.Contains(v, "rule3") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected rule3 violation for udb.NewClient(), got: %v", got)
	}
}

func TestCheckFile_Rule3_CliNewServiceClient_Clean(t *testing.T) {
	dir := t.TempDir()
	// cli.NewServiceClient is explicitly allowed.
	src := `package cmd

import (
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-sdk-go/services/udb"
)

func setup(ctx *cli.Context) {
	client := cli.NewServiceClient(ctx, udb.NewClient)
	_ = client
}
`
	path := writeFile(t, dir, "udb/cmd.go", src)
	got := checkFile(path, "udb")
	for _, v := range got {
		if strings.Contains(v, "rule3") {
			t.Errorf("unexpected rule3 violation: %v", v)
		}
	}
}

func TestCheckFile_Rule4_SetFlagValuesFunc(t *testing.T) {
	dir := t.TempDir()
	src := `package cmd

func setup(f someFlag) {
	f.SetFlagValuesFunc(func() []string { return nil })
}
`
	path := writeFile(t, dir, "udb/cmd.go", src)
	got := checkFile(path, "udb")
	found := false
	for _, v := range got {
		if strings.Contains(v, "rule4") && strings.Contains(v, "SetFlagValuesFunc") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected rule4 violation for SetFlagValuesFunc, got: %v", got)
	}
}

func TestCheckFile_Rule4_GetFlagValues(t *testing.T) {
	dir := t.TempDir()
	src := `package cmd

func setup(f someFlag) []string {
	return f.GetFlagValues()
}
`
	path := writeFile(t, dir, "udb/cmd.go", src)
	got := checkFile(path, "udb")
	found := false
	for _, v := range got {
		if strings.Contains(v, "rule4") && strings.Contains(v, "GetFlagValues") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected rule4 violation for GetFlagValues, got: %v", got)
	}
}

func TestCheckFile_Rule4_SetFlagValues_CommandReceiver_Clean(t *testing.T) {
	dir := t.TempDir()
	// command.SetFlagValues is the allowed wrapper — must not be flagged.
	src := `package cmd

import "github.com/ucloud/ucloud-cli/pkg/command"

func setup(f someFlag) {
	command.SetFlagValues(f, []string{"a", "b"})
}
`
	path := writeFile(t, dir, "udb/cmd.go", src)
	got := checkFile(path, "udb")
	for _, v := range got {
		if strings.Contains(v, "rule4") {
			t.Errorf("unexpected rule4 violation for command.SetFlagValues: %v", v)
		}
	}
}

func TestCheckFile_Clean(t *testing.T) {
	dir := t.TempDir()
	src := `package cmd

import (
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
	"github.com/ucloud/ucloud-sdk-go/services/udb"
)

func setup(ctx *cli.Context, f someFlag) {
	client := cli.NewServiceClient(ctx, udb.NewClient)
	command.SetFlagValues(f, []string{"a", "b"})
	_ = client
}
`
	path := writeFile(t, dir, "udb/cmd.go", src)
	got := checkFile(path, "udb")
	if len(got) != 0 {
		t.Errorf("expected no violations for clean file, got: %v", got)
	}
}

// --------------------------------------------------------------------------
// checkConsistency tests
// --------------------------------------------------------------------------

func TestCheckConsistency_DirWithoutYAML_Violation(t *testing.T) {
	products := []Product{{Name: "udb", Dir: "products/udb", Enabled: true}}
	dirs := []string{"udb", "mystery"} // mystery 无 product.yaml
	violations, _ := checkConsistency(products, dirs)
	found := false
	for _, v := range violations {
		if strings.Contains(v, "mystery") && strings.Contains(v, "rule5") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected rule5 violation for dir without product.yaml, got: %v", violations)
	}
}

func TestCheckConsistency_AllHaveYAML_Clean(t *testing.T) {
	products := []Product{{Name: "udb", Dir: "products/udb", Enabled: true}}
	dirs := []string{"udb"}
	violations, _ := checkConsistency(products, dirs)
	if len(violations) != 0 {
		t.Errorf("expected no violations, got: %v", violations)
	}
}

func TestLoadProducts(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "products/udb/product.yaml", "name: udb\nowners: [Episkey-G]\ncommands: [mysql]\nenabled: true\n")
	t.Chdir(dir)
	got, err := loadProducts()
	if err != nil {
		t.Fatalf("loadProducts: %v", err)
	}
	if len(got) != 1 || got[0].Name != "udb" || got[0].Dir != "products/udb" || len(got[0].Commands) != 1 {
		t.Fatalf("unexpected: %+v", got)
	}
}

// --------------------------------------------------------------------------
// checkReservedCommands tests
// --------------------------------------------------------------------------

func TestCheckReservedCommands_ReservedName_Violation(t *testing.T) {
	// A product declaring the platform-reserved "config" command must violate.
	products := []Product{
		{Name: "rogue", Dir: "products/rogue", Commands: []string{"config"}, Enabled: true},
	}

	violations := checkReservedCommands(products)

	found := false
	for _, v := range violations {
		if strings.Contains(v, "rule6") &&
			strings.Contains(v, "rogue") &&
			strings.Contains(v, "config") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected rule6 violation for reserved command 'config', got: %v", violations)
	}
}

func TestCheckReservedCommands_RealRegistry_Clean(t *testing.T) {
	// The real registry (udb → mysql) declares no reserved name → clean.
	products := []Product{
		{Name: "udb", Dir: "products/udb", Commands: []string{"mysql"}, Enabled: true},
	}

	violations := checkReservedCommands(products)
	if len(violations) != 0 {
		t.Errorf("expected no violations for clean registry, got: %v", violations)
	}
}

// --------------------------------------------------------------------------
// checkCommandCollisions tests (rule7)
// --------------------------------------------------------------------------

func TestCheckCommandCollisions_Duplicate_Violation(t *testing.T) {
	products := []Product{
		{Name: "uhost", Dir: "products/uhost", Commands: []string{"uhost"}, Enabled: true},
		{Name: "compute", Dir: "products/compute", Commands: []string{"uhost"}, Enabled: true},
	}
	violations := checkCommandCollisions(products)
	found := false
	for _, v := range violations {
		if strings.Contains(v, "rule7") && strings.Contains(v, "uhost") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected rule7 violation for duplicate command 'uhost', got: %v", violations)
	}
}

func TestCheckCommandCollisions_UniqueRegistry_Clean(t *testing.T) {
	products := []Product{
		{Name: "udb", Dir: "products/udb", Commands: []string{"mysql"}, Enabled: true},
		{Name: "uhost", Dir: "products/uhost", Commands: []string{"uhost"}, Enabled: true},
	}
	violations := checkCommandCollisions(products)
	if len(violations) != 0 {
		t.Errorf("expected no violations for unique commands, got: %v", violations)
	}
}

func TestCheckCommandCollisions_DisabledIgnored_Clean(t *testing.T) {
	// 被禁用产品即便重名也不算冲突(它不会被注册进命令树)。
	products := []Product{
		{Name: "uhost", Dir: "products/uhost", Commands: []string{"uhost"}, Enabled: true},
		{Name: "legacy", Dir: "products/legacy", Commands: []string{"uhost"}, Enabled: false},
	}
	violations := checkCommandCollisions(products)
	if len(violations) != 0 {
		t.Errorf("expected no violations when duplicate is disabled, got: %v", violations)
	}
}

// --------------------------------------------------------------------------
// rule8: commands consistency (product.go Metadata vs products.yaml)
// --------------------------------------------------------------------------

func TestSameStringSet(t *testing.T) {
	cases := []struct {
		a, b []string
		want bool
	}{
		{[]string{"mysql"}, []string{"mysql"}, true},
		{[]string{"redis", "memcache"}, []string{"memcache", "redis"}, true}, // order-independent
		{[]string{"mysql"}, []string{"mysql", "extra"}, false},
		{[]string{"mysql"}, []string{"redis"}, false},
		{nil, nil, true},
	}
	for i, c := range cases {
		if got := sameStringSet(c.a, c.b); got != c.want {
			t.Errorf("case %d: sameStringSet(%v,%v)=%v want %v", i, c.a, c.b, got, c.want)
		}
	}
}

func TestExtractMetadataCommands(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "product.go", `package udb

import "github.com/ucloud/ucloud-cli/pkg/cli"

type product struct{}

func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "udb", Commands: []string{"mysql"}}
}
`)
	got, err := extractMetadataCommands(dir)
	if err != nil {
		t.Fatalf("extractMetadataCommands: %v", err)
	}
	if len(got) != 1 || got[0] != "mysql" {
		t.Fatalf("expected [mysql], got %v", got)
	}
}

func TestCheckCommandsConsistency_Mismatch_Violation(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "products/udb/product.go", `package udb

import "github.com/ucloud/ucloud-cli/pkg/cli"

type product struct{}

func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "udb", Commands: []string{"mysql", "extra"}}
}
`)
	t.Chdir(dir) // Go 1.24+: chdir for this test, auto-restored
	products := []Product{
		{Name: "udb", Dir: "products/udb", Commands: []string{"mysql"}, Enabled: true},
	}
	violations := checkCommandsConsistency(products)
	found := false
	for _, v := range violations {
		if strings.Contains(v, "rule8") && strings.Contains(v, "udb") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected rule8 mismatch violation, got: %v", violations)
	}
}

func TestCheckCommandsConsistency_Match_Clean(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "products/udb/product.go", `package udb

import "github.com/ucloud/ucloud-cli/pkg/cli"

type product struct{}

func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "udb", Commands: []string{"mysql"}}
}
`)
	t.Chdir(dir)
	products := []Product{
		{Name: "udb", Dir: "products/udb", Commands: []string{"mysql"}, Enabled: true},
	}
	violations := checkCommandsConsistency(products)
	if len(violations) != 0 {
		t.Errorf("expected no violations for matching commands, got: %v", violations)
	}
}
