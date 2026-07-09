package gssh

import (
	"fmt"

	"github.com/spf13/cobra"

	pathxsdk "github.com/ucloud/ucloud-sdk-go/services/pathx"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newModify ucloud gssh update
func newModify(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, pathxsdk.NewClient)
	gsshModifyPortReq := client.NewModifyGlobalSSHPortRequest()
	gsshModifyRemarkReq := client.NewModifyGlobalSSHRemarkRequest()
	project := ctx.DefaultProjectID()
	gsshIDs := []string{}
	cmd := &cobra.Command{
		Use:     "update",
		Short:   "Update GlobalSSH instance",
		Long:    "Update GlobalSSH instance, including port and remark attribute",
		Example: "ucloud gssh update --gssh-id uga-xxx --port 22",
		Run: func(cmd *cobra.Command, args []string) {
			gsshModifyPortReq.ProjectId = sdk.String(ctx.PickResourceID(project))
			gsshModifyRemarkReq.ProjectId = sdk.String(ctx.PickResourceID(project))
			if *gsshModifyPortReq.Port == 0 && *gsshModifyRemarkReq.Remark == "" {
				fmt.Fprintln(ctx.ProgressWriter(), "Error, port or remark required")
			}
			results := []cli.OpResultRow{}
			if *gsshModifyPortReq.Port != 0 {
				port := *gsshModifyPortReq.Port
				if port <= 1 || port >= 65535 || port == 80 || port == 443 || port == 65123 {
					fmt.Fprintln(ctx.ProgressWriter(), "The port number should be between 1 and 65535, and cannot be equal to 80, 443 or 65123")
					return
				}
				for _, idname := range gsshIDs {
					id := ctx.PickResourceID(idname)
					gsshModifyPortReq.InstanceId = sdk.String(id)
					_, err := client.ModifyGlobalSSHPort(gsshModifyPortReq)
					if err != nil {
						ctx.HandleError(err)
						continue
					}
					fmt.Fprintf(ctx.ProgressWriter(), "gssh[%s]'s port updated\n", id)
					results = append(results, cli.OpResultRow{ResourceID: id, Action: "update-port", Status: "Updated"})
				}
			}
			if *gsshModifyRemarkReq.Remark != "" {
				for _, idname := range gsshIDs {
					id := ctx.PickResourceID(idname)
					gsshModifyRemarkReq.InstanceId = sdk.String(id)
					_, err := client.ModifyGlobalSSHRemark(gsshModifyRemarkReq)
					if err != nil {
						ctx.HandleError(err)
						continue
					}
					fmt.Fprintf(ctx.ProgressWriter(), "gssh[%s]'s remark updated\n", id)
					results = append(results, cli.OpResultRow{ResourceID: id, Action: "update-remark", Status: "Updated"})
				}
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&gsshIDs, "gssh-id", nil, "Required. ResourceID of your GlobalSSH instances")
	flags.StringVar(&project, "project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	gsshModifyPortReq.Port = flags.Int("port", 0, "Optional. Port of SSH service.")
	gsshModifyRemarkReq.Remark = flags.String("remark", "", "Optional. Remark of your GlobalSSH.")
	cmd.MarkFlagRequired("gssh-id")
	ctx.SetCompletion(cmd, "project-id", ctx.ProjectList)
	ctx.SetCompletion(cmd, "gssh-id", func() []string {
		return getAllGsshIDNames(ctx, project)
	})
	return cmd
}
