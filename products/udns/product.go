package udns

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	internaludns "github.com/ucloud/ucloud-cli/products/udns/internal/udns"
)

type udns struct{}

func New() cli.Product {
	return udns{}
}

func (u udns) Metadata() cli.Metadata {
	return cli.Metadata{
		Name:     "udns",
		Commands: []string{"udns"},
	}
}

func (u udns) NewCommand(ctx *cli.Context) []*cobra.Command {
	return []*cobra.Command{internaludns.NewCommand(ctx)}
}

var _ cli.Product = udns{}
