package onboarding

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newCommand assembles the product's command tree: construct the top-level
// command and AddCommand one constructor per verb — this aggregator is the
// ONLY content allowed in cmd.go (§2 file-layout convention: one verb per
// file, named after the subcommand).
func newCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   productName,
		Short: "Greenfield example product (onboarding worked example)",
		Long: "Greenfield example product demonstrating the ucloud-cli platform " +
			"onboarding contract. Not a real product; exists as the onboarding " +
			"worked example and the platform-API compile gate.",
	}

	cmd.AddCommand(newList(ctx))
	cmd.AddCommand(newDescribe(ctx))
	cmd.AddCommand(newCreate(ctx))
	cmd.AddCommand(newDelete(ctx))
	cmd.AddCommand(newStart(ctx))
	cmd.AddCommand(newStop(ctx))
	cmd.AddCommand(newRestart(ctx))

	return cmd
}
