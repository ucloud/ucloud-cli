package cmd

import (
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

type multiCommandProductForTest struct{}

func (p multiCommandProductForTest) Metadata() cli.Metadata {
	return cli.Metadata{Name: "multi", Commands: []string{"alpha", "beta"}}
}

func (p multiCommandProductForTest) NewCommand(ctx *cli.Context) []*cobra.Command {
	return []*cobra.Command{
		{Use: "alpha"},
		{Use: "beta"},
	}
}

func TestRegisteredProductsUseCommandDirectoryProducts(t *testing.T) {
	products := registeredProducts()

	byName := make(map[string]cli.Product, len(products))
	for _, p := range products {
		byName[p.Metadata().Name] = p
	}

	for _, tt := range []struct {
		product  string
		commands []string
	}{
		{product: "sharedbw", commands: []string{"bw"}},
		{product: "eip", commands: []string{"eip"}},
		{product: "firewall", commands: []string{"firewall"}},
		{product: "globalssh", commands: []string{"gssh"}},
		{product: "image", commands: []string{"image"}},
		{product: "memcache", commands: []string{"memcache"}},
		{product: "mysql", commands: []string{"mysql"}},
		{product: "pathx", commands: []string{"pathx"}},
		{product: "redis", commands: []string{"redis"}},
		{product: "subnet", commands: []string{"subnet"}},
		{product: "udisk", commands: []string{"udisk"}},
		{product: "udpn", commands: []string{"udpn"}},
		{product: "uhost", commands: []string{"uhost"}},
		{product: "ulb", commands: []string{"ulb"}},
		{product: "umodelverse", commands: []string{"umodelverse"}},
		{product: "uphost", commands: []string{"uphost"}},
		{product: "vpc", commands: []string{"vpc"}},
	} {
		assertProductCommands(t, byName, tt.product, tt.commands)
	}

	for _, removedName := range []string{"bw", "gssh", "udb", "umem", "unet"} {
		if _, ok := byName[removedName]; ok {
			t.Fatalf("registeredProducts includes grouped/stale product %q; want existing top-level commands in independent directories", removedName)
		}
	}
}

func TestAddProductCommandsRegistersAllProductCommands(t *testing.T) {
	root := &cobra.Command{Use: "ucloud"}

	addProductCommands(root, []cli.Product{multiCommandProductForTest{}}, cli.NewContext(cli.Deps{}))

	for _, name := range []string{"alpha", "beta"} {
		if _, _, err := root.Find([]string{name}); err != nil {
			t.Fatalf("product command %q was not registered: %v", name, err)
		}
	}
}

func TestAddPlatformCommandsExcludesMigratedProductCommands(t *testing.T) {
	src, err := os.ReadFile("root.go")
	if err != nil {
		t.Fatalf("read root.go: %v", err)
	}
	for _, constructor := range []string{
		"NewCmdUDPN(",
		"NewCmdGssh(",
		"NewCmdPathx(",
		"NewCmdBandwidth(",
		"NewCmdRedis(",
		"NewCmdMemcache(",
		"NewCmdULB(",
		"NewCmdSubnet(",
		"NewCmdVpc(",
	} {
		if strings.Contains(string(src), constructor) {
			t.Fatalf("addPlatformCommands must not register %s after product migration", constructor)
		}
	}
}

func assertProductCommands(t *testing.T, products map[string]cli.Product, name string, want []string) {
	t.Helper()

	p, ok := products[name]
	if !ok {
		t.Fatalf("registeredProducts missing product %q", name)
	}

	got := append([]string(nil), p.Metadata().Commands...)
	sort.Strings(got)
	sort.Strings(want)
	if len(got) != len(want) {
		t.Fatalf("product %q commands = %v, want %v", name, got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("product %q commands = %v, want %v", name, got, want)
		}
	}
}
