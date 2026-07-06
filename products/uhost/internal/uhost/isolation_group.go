package uhost

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newIsolationGroup ucloud uhost isolation-group
// Mirrors cmd/uhost.go NewCmdIsolation (AddCommand order: list, create, delete).
func newIsolationGroup(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "isolation-group",
		Short: "List and manipulate isolation group of uhost",
		Long:  "List and manipulate isolation group of uhost",
	}
	cmd.AddCommand(newIsolationList(ctx))
	cmd.AddCommand(newIsolationCreate(ctx))
	cmd.AddCommand(newIsolationDelete(ctx))
	return cmd
}
