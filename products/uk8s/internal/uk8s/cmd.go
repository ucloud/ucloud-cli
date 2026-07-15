package uk8s

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// NewCommand builds the `uk8s` root command. Per the platform spec (§2.2
// aggregator role), this file only constructs the top-level command and wires
// the verb constructors via AddCommand — no business logic, no helpers.
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uk8s",
		Short: "Read and manipulate UK8S (UCloud Kubernetes Service) clusters",
		Long:  "Read and manipulate UK8S (UCloud Kubernetes Service) clusters",
	}
	cmd.AddCommand(newCreate(ctx))
	cmd.AddCommand(newDelete(ctx))
	cmd.AddCommand(newList(ctx))
	cmd.AddCommand(newDescribe(ctx))
	cmd.AddCommand(newGetConfig(ctx))
	cmd.AddCommand(newNodeGroup(ctx))
	cmd.AddCommand(newNode(ctx))
	cmd.AddCommand(newImage(ctx))
	cmd.AddCommand(newVersion(ctx))
	return cmd
}
