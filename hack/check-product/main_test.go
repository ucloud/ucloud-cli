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
	_ "github.com/ucloud/ucloud-cli/products/mysql/internal/helper"
)
`
	path := writeFile(t, dir, "mysql/cmd.go", src)
	got := checkFile(path, "mysql")
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
	src := "package cmd\n\n" +
		"import (\n" +
		"\t\"" + moduleRoot + "/base\"\n" +
		")\n\n" +
		"var _ = base.Foo\n"

	path := writeFile(t, dir, "udb/cmd.go", src)
	got := checkFile(path, "udb")
	if len(got) == 0 {
		t.Fatal("expected violation for base import, got none")
	}
	if !strings.Contains(got[0], "rule2") {
		t.Errorf("expected rule2 in violation, got: %v", got)
	}
}

func TestCheckFile_Rule2_LegacyPlatformImports(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		{
			name: "ux",
			src:  "package cmd\n\nimport \"" + moduleRoot + "/ux\"\n\nvar _ = ux.Doc\n",
		},
		{
			name: "ansi",
			src:  "package cmd\n\nimport \"" + moduleRoot + "/ansi\"\n\nvar _ = ansi.CursorLeft\n",
		},
		{
			name: "cmd/internal",
			src: `package cmd

import "github.com/ucloud/ucloud-cli/cmd/internal/runtime"

var _ = runtime.Active
`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := writeFile(t, t.TempDir(), "udb/cmd.go", tc.src)
			got := checkFile(path, "udb")
			if len(got) == 0 {
				t.Fatalf("expected violation for %s import, got none", tc.name)
			}
			if !strings.Contains(got[0], "rule2") {
				t.Errorf("expected rule2 in violation, got: %v", got)
			}
		})
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
	products := []Product{{Name: "mysql", Dir: "products/mysql", Enabled: true}}
	dirs := []string{"mysql", "mystery"} // mystery 无 product.yaml
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
	products := []Product{{Name: "mysql", Dir: "products/mysql", Enabled: true}}
	dirs := []string{"mysql"}
	violations, _ := checkConsistency(products, dirs)
	if len(violations) != 0 {
		t.Errorf("expected no violations, got: %v", violations)
	}
}

func TestLoadProducts(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "products/mysql/product.yaml", "name: mysql\nowners: [Episkey-G]\ncommands: [mysql]\nenabled: true\n")
	t.Chdir(dir)
	got, err := loadProducts()
	if err != nil {
		t.Fatalf("loadProducts: %v", err)
	}
	if len(got) != 1 || got[0].Name != "mysql" || got[0].Dir != "products/mysql" || len(got[0].Commands) != 1 {
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
	// The real registry declares no reserved name → clean.
	products := []Product{
		{Name: "mysql", Dir: "products/mysql", Commands: []string{"mysql"}, Enabled: true},
	}

	violations := checkReservedCommands(products)
	if len(violations) != 0 {
		t.Errorf("expected no violations for clean registry, got: %v", violations)
	}
}

func TestReservedCommands_ExcludesStaleBandwidthImplementationName(t *testing.T) {
	if reservedCommands["bandwidth"] {
		t.Fatal("reservedCommands must not include stale implementation name \"bandwidth\"")
	}
}

func TestReservedCommands_ExcludesMigratedUDPNProduct(t *testing.T) {
	if reservedCommands["udpn"] {
		t.Fatal("reservedCommands must not include udpn after it migrates to products/udpn")
	}
}

func TestReservedCommands_ExcludesMigratedGSSHProductCommand(t *testing.T) {
	if reservedCommands["gssh"] {
		t.Fatal("reservedCommands must not include gssh after it migrates to products/globalssh")
	}
}

func TestReservedCommands_ExcludesMigratedPathXProductCommand(t *testing.T) {
	if reservedCommands["pathx"] {
		t.Fatal("reservedCommands must not include pathx after it migrates to products/pathx")
	}
}

func TestReservedCommands_ExcludesMigratedBandwidthProductCommand(t *testing.T) {
	if reservedCommands["bw"] {
		t.Fatal("reservedCommands must not include bw after it migrates to products/sharedbw")
	}
}

func TestReservedCommands_ExcludesMigratedRedisProductCommand(t *testing.T) {
	if reservedCommands["redis"] {
		t.Fatal("reservedCommands must not include redis after it migrates to products/redis")
	}
}

func TestReservedCommands_ExcludesMigratedMemcacheProductCommand(t *testing.T) {
	if reservedCommands["memcache"] {
		t.Fatal("reservedCommands must not include memcache after it migrates to products/memcache")
	}
}

func TestReservedCommands_ExcludesMigratedULBProductCommand(t *testing.T) {
	if reservedCommands["ulb"] {
		t.Fatal("reservedCommands must not include ulb after it migrates to products/ulb")
	}
}

func TestReservedCommands_ExcludesMigratedVPCProductCommand(t *testing.T) {
	if reservedCommands["vpc"] {
		t.Fatal("reservedCommands must not include vpc after it migrates to products/vpc")
	}
}

func TestReservedCommands_ExcludesMigratedSubnetProductCommand(t *testing.T) {
	if reservedCommands["subnet"] {
		t.Fatal("reservedCommands must not include subnet after it migrates to products/subnet")
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
		{Name: "mysql", Dir: "products/mysql", Commands: []string{"mysql"}, Enabled: true},
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
	writeFile(t, dir, "product.go", `package mysql

import "github.com/ucloud/ucloud-cli/pkg/cli"

type product struct{}

func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "mysql", Commands: []string{"mysql"}}
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
	writeFile(t, dir, "products/mysql/product.go", `package mysql

import "github.com/ucloud/ucloud-cli/pkg/cli"

type product struct{}

func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "mysql", Commands: []string{"mysql", "extra"}}
}
`)
	t.Chdir(dir) // Go 1.24+: chdir for this test, auto-restored
	products := []Product{
		{Name: "mysql", Dir: "products/mysql", Commands: []string{"mysql"}, Enabled: true},
	}
	violations := checkCommandsConsistency(products)
	found := false
	for _, v := range violations {
		if strings.Contains(v, "rule8") && strings.Contains(v, "mysql") {
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
	writeFile(t, dir, "products/mysql/product.go", `package mysql

import "github.com/ucloud/ucloud-cli/pkg/cli"

type product struct{}

func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "mysql", Commands: []string{"mysql"}}
}
`)
	t.Chdir(dir)
	products := []Product{
		{Name: "mysql", Dir: "products/mysql", Commands: []string{"mysql"}, Enabled: true},
	}
	violations := checkCommandsConsistency(products)
	if len(violations) != 0 {
		t.Errorf("expected no violations for matching commands, got: %v", violations)
	}
}

// --------------------------------------------------------------------------
// rule9: §6.1 product import whitelist
// --------------------------------------------------------------------------

func TestCheckFile_Rule9_ImportWhitelist(t *testing.T) {
	src := `package p

	import (
		_ "fmt"
		_ "github.com/spf13/cobra"
		_ "github.com/ucloud/ucloud-cli/internal/common"
		_ "github.com/ucloud/ucloud-cli/model/status"
		_ "github.com/ucloud/ucloud-cli/pkg/cli"
		_ "github.com/ucloud/ucloud-cli/products/mysql/internal/mysql"
		_ "github.com/ucloud/ucloud-sdk-go/private/services/uhost"
		_ "github.com/fatih/color"
	)
`
	joined := strings.Join(checkFile(writeFile(t, t.TempDir(), "x.go", src), "mysql"), "\n")

	for _, bad := range []string{
		`rule9: import "github.com/ucloud/ucloud-cli/model/status"`,
		`rule9: import "github.com/fatih/color"`,
	} {
		if !strings.Contains(joined, bad) {
			t.Fatalf("missing violation %q in:\n%s", bad, joined)
		}
	}
	for _, ok := range []string{`"fmt"`, `"github.com/spf13/cobra"`,
		`"github.com/ucloud/ucloud-cli/internal/common"`,
		`"github.com/ucloud/ucloud-cli/pkg/cli"`,
		`"github.com/ucloud/ucloud-cli/products/mysql/internal/mysql"`,
		`"github.com/ucloud/ucloud-sdk-go/private/services/uhost"`} {
		if strings.Contains(joined, "rule9: import "+ok) {
			t.Fatalf("false positive on %s in:\n%s", ok, joined)
		}
	}
}

func TestCheckFile_Rule9_DoesNotDuplicateRules12(t *testing.T) {
	src := "package p\n\n" +
		"import (\n" +
		"\t_ \"" + moduleRoot + "/base\"\n" +
		"\t_ \"" + moduleRoot + "/products/vpc/internal/vpc\"\n" +
		")\n"
	joined := strings.Join(checkFile(writeFile(t, t.TempDir(), "x.go", src), "udb"), "\n")
	if strings.Count(joined, "ucloud-cli/base") != 1 || strings.Count(joined, "products/vpc") != 1 {
		t.Fatalf("rule1/2 territory must be flagged exactly once (by rule1/2, not rule9):\n%s", joined)
	}
}

// --------------------------------------------------------------------------
// rule10: §2 file layout (grab-bag filenames + one cobra constructor per file)
// --------------------------------------------------------------------------

func TestCheckFile_Rule10_OneConstructorPerFile(t *testing.T) {
	dir := t.TempDir()

	// Two cobra constructors in one file (a method counts too) → exactly one
	// rule10 violation naming the file.
	src := `package p

import "github.com/spf13/cobra"

func NewCmdList() *cobra.Command { return &cobra.Command{} }

func (b builder) NewCmdCreate() *cobra.Command { return &cobra.Command{} }
`
	path := writeFile(t, dir, "udb/list.go", src)
	got := checkFile(path, "udb")
	count := 0
	for _, v := range got {
		if strings.Contains(v, "rule10") {
			count++
			if !strings.Contains(v, path) {
				t.Errorf("rule10 violation must name the file %s, got: %v", path, v)
			}
		}
	}
	if count != 1 {
		t.Fatalf("expected exactly one rule10 violation for two constructors, got %d: %v", count, got)
	}

	// One constructor + unrelated funcs (incl. a FuncLit returning
	// *cobra.Command inside a body) → zero rule10.
	clean := `package p

import "github.com/spf13/cobra"

func NewCmdList() *cobra.Command {
	build := func() *cobra.Command { return &cobra.Command{} }
	return build()
}

func rows() []string { return nil }

func (x *thing) status() error { return nil }
`
	cleanPath := writeFile(t, dir, "udb/clean.go", clean)
	for _, v := range checkFile(cleanPath, "udb") {
		if strings.Contains(v, "rule10") {
			t.Errorf("unexpected rule10 violation for single-constructor file: %v", v)
		}
	}

	// _test.go files are exempt from the constructor budget.
	testPath := writeFile(t, dir, "udb/list_test.go", src)
	for _, v := range checkFile(testPath, "udb") {
		if strings.Contains(v, "rule10") {
			t.Errorf("unexpected rule10 violation for _test.go file: %v", v)
		}
	}
}

func TestCheckFilename_Rule10_GrabBagNames(t *testing.T) {
	flagged := []string{
		"products/uhost/helpers.go",
		"products/uhost/internal/uhost/utils.go",
		"products/udisk/util.go",
		"products/eip/common.go",
		"products/image/misc.go",
		"products/uhost/helpers_test.go",
	}
	for _, p := range flagged {
		got := checkFilename(p)
		if len(got) != 1 || !strings.Contains(got[0], "rule10") {
			t.Errorf("expected one rule10 violation for %s, got: %v", p, got)
		}
	}

	clean := []string{
		"products/uhost/list.go",
		"products/uhost/rows.go",
		"products/uhost/describe.go",
		"products/uhost/status.go",
		"products/uhost/x.go",
		"products/uhost/list_test.go",
		"products/uhost/product.yaml", // non-.go files are out of scope
	}
	for _, p := range clean {
		if got := checkFilename(p); len(got) != 0 {
			t.Errorf("unexpected rule10 violation for %s: %v", p, got)
		}
	}
}
