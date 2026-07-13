package uhadoop

import (
	"github.com/spf13/cobra"

	uhadoopsdk "github.com/ucloud/ucloud-sdk-go/services/uhadoop"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newDelete ucloud uhadoop delete
func newDelete(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, uhadoopsdk.NewClient)
	req := client.NewDeleteUHadoopInstanceRequest()
	cmd := &cobra.Command{
		Use:   "delete <instance-id>",
		Short: "Delete a UHadoop cluster",
		Long:  `Delete a UHadoop cluster by instance ID`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			req.InstanceId = sdkStr(args[0])
			resp, err := client.DeleteUHadoopInstance(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			ctx.PrintJSON(resp)
		},
	}
	cmd.Flags().SortFlags = false

	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)
	req.ReleaseEIP = cmd.Flags().Bool("release-eip", false, "Optional. Release bound EIP after deletion")

	command.SetFlagValues(cmd, "release-eip", "true", "false")

	return cmd
}
