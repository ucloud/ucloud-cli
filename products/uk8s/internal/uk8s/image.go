package uk8s

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func newImage(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{Use: "image", Short: "Inspect UK8S images"}
	cmd.AddCommand(newImageList(ctx))
	return cmd
}
