package memcache

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	internalmemcache "github.com/ucloud/ucloud-cli/products/memcache/internal/memcache"
)

type product struct{}

func New() cli.Product { return product{} }

func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "memcache", Commands: []string{"memcache"}}
}

func (product) NewCommand(ctx *cli.Context) []*cobra.Command {
	return []*cobra.Command{internalmemcache.NewCommand(ctx)}
}
