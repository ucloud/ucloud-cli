package uhost

import (
	"fmt"

	"github.com/spf13/cobra"

	uhostsdk "github.com/ucloud/ucloud-sdk-go/services/uhost"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newIsolationDelete ucloud uhost isolation-group delete
func newIsolationDelete(ctx *cli.Context) *cobra.Command {
	var ids []string
	client := cli.NewServiceClient(ctx, uhostsdk.NewClient)
	req := client.NewDeleteIsolationGroupRequest()
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete isolation group instances",
		Run: func(c *cobra.Command, args []string) {
			results := []cli.OpResultRow{}
			for _, idname := range ids {
				id := ctx.PickResourceID(idname)
				req.GroupId = &id
				_, err := client.DeleteIsolationGroup(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				fmt.Fprintf(ctx.ProgressWriter(), "isolation group %s deleted\n", idname)
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "delete", Status: "Deleted"})
			}
			ctx.EmitResult(results...)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringSliceVar(&ids, "group-id", nil, "Required. Resource ID of isolation groups to be deleted")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)

	cmd.MarkFlagRequired("group-id")
	command.SetCompletion(cmd, "group-id", func() []string {
		return getIsolationGroupList(ctx, *req.ProjectId, *req.Region)
	})

	return cmd
}
