package uhadoop

import (
	"fmt"

	"github.com/spf13/cobra"

	uhadoopsdk "github.com/ucloud/ucloud-sdk-go/services/uhadoop"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newDelete ucloud uhadoop delete
func newDelete(ctx *cli.Context) *cobra.Command {
	var yes bool
	client := cli.NewServiceClient(ctx, uhadoopsdk.NewClient)
	req := client.NewDeleteUHadoopInstanceRequest()
	cmd := &cobra.Command{
		Use:          "delete <instance-id>",
		Short:        "Delete a UHadoop cluster",
		Long:         `Delete a UHadoop cluster by instance ID`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			ok, err := ctx.Confirm(yes, fmt.Sprintf("Are you sure you want to delete cluster %s?", id))
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}
			req.InstanceId = sdkStr(id)
			resp, err := client.DeleteUHadoopInstance(req)
			if err != nil {
				return err
			}
			if resp.RetCode != 0 {
				return fmt.Errorf("[%d] %s", resp.RetCode, resp.Message)
			}
			fmt.Fprintf(ctx.Err(), "Cluster %s deleted\n", id)
			ctx.PrintJSON(resp)
			return nil
		},
	}
	cmd.Flags().SortFlags = false

	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Optional. Do not prompt for confirmation.")
	req.ReleaseEIP = cmd.Flags().Bool("release-eip", false, "Optional. Release bound EIP after deletion")

	command.SetFlagValues(cmd, "release-eip", "true", "false")

	cmd.MarkFlagRequired("region")
	cmd.MarkFlagRequired("zone")
	return cmd
}
