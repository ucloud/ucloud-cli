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

func TestCheckConsistency_EnabledProductMissingDir_Warn(t *testing.T) {
	products := []Product{
		{Name: "udb", Dir: "products/udb", Enabled: true},
	}
	dirs := []string{} // udb dir does not exist

	violations, warnings := checkConsistency(products, dirs)

	// Missing-dir for enabled product must be a WARNING, not a violation.
	if len(violations) != 0 {
		t.Errorf("expected no violations, got: %v", violations)
	}
	if len(warnings) == 0 {
		t.Error("expected warning for missing enabled-product dir, got none")
	}
	if !strings.Contains(warnings[0], "warn") {
		t.Errorf("expected warn prefix, got: %v", warnings[0])
	}
}

func TestCheckConsistency_UnknownDir_Violation(t *testing.T) {
	products := []Product{
		{Name: "udb", Dir: "products/udb", Enabled: true},
	}
	dirs := []string{"udb", "mystery"} // "mystery" has no products.yaml entry

	violations, _ := checkConsistency(products, dirs)

	found := false
	for _, v := range violations {
		if strings.Contains(v, "mystery") && strings.Contains(v, "rule5") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected rule5 violation for unknown directory 'mystery', got: %v", violations)
	}
}

func TestCheckConsistency_AllMatch_Clean(t *testing.T) {
	products := []Product{
		{Name: "udb", Dir: "products/udb", Enabled: true},
	}
	dirs := []string{"udb"}

	violations, warnings := checkConsistency(products, dirs)
	if len(violations) != 0 {
		t.Errorf("expected no violations, got: %v", violations)
	}
	if len(warnings) != 0 {
		t.Errorf("expected no warnings, got: %v", warnings)
	}
}

func TestCheckConsistency_DisabledProductMissingDir_NoWarn(t *testing.T) {
	// A disabled product whose dir is absent should produce no warning/violation.
	products := []Product{
		{Name: "vpc", Dir: "products/vpc", Enabled: false},
	}
	dirs := []string{}

	violations, warnings := checkConsistency(products, dirs)
	if len(violations) != 0 {
		t.Errorf("expected no violations, got: %v", violations)
	}
	if len(warnings) != 0 {
		t.Errorf("expected no warnings for disabled product, got: %v", warnings)
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
