package sqlserver

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	internalsqlserver "github.com/ucloud/ucloud-cli/products/sqlserver/internal/sqlserver"
)

type product struct{}

func New() cli.Product { return product{} }

func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "sqlserver", Commands: []string{"sqlserver"}}
}

func (product) NewCommand(ctx *cli.Context) []*cobra.Command {
	return []*cobra.Command{internalsqlserver.NewCommand(ctx)}
}
