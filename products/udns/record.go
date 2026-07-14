package udns

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newRecordCommand builds the `udns record` subgroup.
func newRecordCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "record",
		Short: "Manage DNS records within a UDNS zone",
		Long:  "Manage DNS records within a UDNS zone",
	}
	cmd.AddCommand(newRecordListCommand(ctx))
	cmd.AddCommand(newRecordCreateCommand(ctx))
	cmd.AddCommand(newRecordModifyCommand(ctx))
	cmd.AddCommand(newRecordDeleteCommand(ctx))
	return cmd
}
