package globalssh

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	internalgssh "github.com/ucloud/ucloud-cli/products/globalssh/internal/gssh"
)

type product struct{}

func New() cli.Product { return product{} }

func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "globalssh", Commands: []string{"gssh"}}
}

func (product) NewCommand(ctx *cli.Context) []*cobra.Command {
	return []*cobra.Command{internalgssh.NewCommand(ctx)}
}
