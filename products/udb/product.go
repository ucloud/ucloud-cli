package udb

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/products/udb/internal/mysql"
)

type product struct{}

// New returns the udb product (registered via hack/gen-products).
func New() cli.Product { return product{} }

func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "udb", Owners: []string{"episkey"}, Commands: []string{"mysql"}}
}

func (product) NewCommand(ctx *cli.Context) *cobra.Command { return mysql.NewCommand(ctx) }
