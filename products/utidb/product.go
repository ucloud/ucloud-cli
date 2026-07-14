package utidb

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/products/utidb/internal/tidb"
)

type product struct{}

// New returns the utidb product (registered via hack/gen-products).
func New() cli.Product { return product{} }

func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "utidb", Commands: []string{"utidb"}}
}

func (product) NewCommand(ctx *cli.Context) []*cobra.Command {
	return []*cobra.Command{tidb.NewCommand(ctx)}
}
