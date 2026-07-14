package pathx

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// NewCommand builds the `pathx` root command.
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pathx",
		Short: "Manipulate uga and upath instances",
		Long:  "Manipulate uga and upath instances",
	}
	cmd.AddCommand(newUGA(ctx))
	cmd.AddCommand(newUpath(ctx))
	cmd.AddCommand(newCreate(ctx))
	cmd.AddCommand(newDelete(ctx))
	cmd.AddCommand(newModify(ctx))
	cmd.AddCommand(newList(ctx))
	cmd.AddCommand(newPrice(ctx))
	cmd.AddCommand(newArea(ctx))
	return cmd
}
