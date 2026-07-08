package gssh

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// NewCommand builds the `gssh` root command.
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gssh",
		Short: "Create,list,update and delete globalssh instance",
		Long:  "Create,list,update and delete globalssh instance",
	}
	cmd.AddCommand(newList(ctx))
	cmd.AddCommand(newCreate(ctx))
	cmd.AddCommand(newDelete(ctx))
	cmd.AddCommand(newModify(ctx))
	cmd.AddCommand(newArea(ctx))
	return cmd
}
