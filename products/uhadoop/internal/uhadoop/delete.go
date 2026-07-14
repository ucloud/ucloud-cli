package uhadoop

import (
	"fmt"

	"github.com/spf13/cobra"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	uhadoopsdk "github.com/ucloud/ucloud-sdk-go/services/uhadoop"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

func newDelete(ctx *cli.Context) *cobra.Command {
	var yes bool
	client := cli.NewServiceClient(ctx, uhadoopsdk.NewClient)
	req := client.NewDeleteUHadoopInstanceRequest()
	cmd := &cobra.Command{
		Use:   "delete <instance-id>",
		Short: "Delete a UHadoop cluster",
		Long:  `Delete a UHadoop cluster by instance ID`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]
			ok, err := ctx.Confirm(yes, fmt.Sprintf("Are you sure you want to delete cluster %s?", id))
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if !ok {
				return
			}
			w := ctx.ProgressWriter()
			req.InstanceId = sdk.String(id)
			_, err = client.DeleteUHadoopInstance(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(w, "uhadoop[%s] deleted\n", id)
			ctx.EmitResult(cli.OpResultRow{ResourceID: id, Action: "delete", Status: "Deleted"})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	flags.BoolVarP(&yes, "yes", "y", false, "Optional. Do not prompt for confirmation")
	req.ReleaseEIP = flags.Bool("release-eip", false, "Optional. Release bound EIP after deletion")
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)

	command.SetFlagValues(cmd, "release-eip", "true", "false")
	cmd.MarkFlagRequired("region")
	cmd.MarkFlagRequired("zone")

	return cmd
}
