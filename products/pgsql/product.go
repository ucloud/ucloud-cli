package pgsql

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	internalpgsql "github.com/ucloud/ucloud-cli/products/pgsql/internal/pgsql"
)

type product struct{}

func New() cli.Product { return product{} }

func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "pgsql", Commands: []string{"pgsql"}}
}

func (product) NewCommand(ctx *cli.Context) []*cobra.Command {
	return []*cobra.Command{internalpgsql.NewCommand(ctx)}
}
