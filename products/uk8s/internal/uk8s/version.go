package uk8s

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func newVersion(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{Use: "version", Short: "Inspect UK8S versions"}
	cmd.AddCommand(newVersionList(ctx))
	return cmd
}
