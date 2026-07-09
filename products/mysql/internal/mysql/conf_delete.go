package mysql

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/udb"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newUDBConfDelete ucloud udb conf delete
func newUDBConfDelete(ctx *cli.Context) *cobra.Command {
	var confID string
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewDeleteUDBParamGroupRequest()
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete configuration of udb by conf-id",
		Long:  "Delete configuration of udb by conf-id",
		Run: func(c *cobra.Command, args []string) {
			id, err := strconv.Atoi(ctx.PickResourceID(confID))
			if err != nil {
				ctx.HandleError(err)
				return
			}
			req.GroupId = &id
			_, err = client.DeleteUDBParamGroup(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "conf[%s] deleted\n", confID)
			ctx.EmitResult(cli.OpResultRow{ResourceID: strconv.Itoa(id), Action: "delete", Status: "Deleted"})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&confID, "conf-id", "", "Required. ConfID of the configuration to delete")
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)

	cmd.MarkFlagRequired("conf-id")
	command.SetCompletion(cmd, "conf-id", func() []string {
		return getModifiableConfIDList(ctx, "", *req.ProjectId, *req.Region, *req.Zone)
	})
	return cmd
}
