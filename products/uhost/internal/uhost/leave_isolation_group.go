package uhost

import (
	"fmt"

	"github.com/spf13/cobra"

	uhostsdk "github.com/ucloud/ucloud-sdk-go/services/uhost"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newLeaveIsolationGroup ucloud uhost leave-isolation-group
func newLeaveIsolationGroup(ctx *cli.Context) *cobra.Command {
	var uhostIds []string
	client := cli.NewServiceClient(ctx, uhostsdk.NewClient)
	req := client.NewLeaveIsolationGroupRequest()
	cmd := &cobra.Command{
		Use:   "leave-isolation-group",
		Short: "Detach uhost from its isolation group",
		Run: func(c *cobra.Command, args []string) {
			results := []cli.OpResultRow{}
			for _, idname := range uhostIds {
				id := ctx.PickResourceID(idname)
				any, err := describeUHostByID(ctx, *req.ProjectId, *req.Region, *req.Zone)(id, nil)
				if err != nil {
					ctx.LogError(fmt.Sprintf("fetch uhost %s failed: %v", idname, err))
					continue
				}
				ins, ok := any.(*uhostsdk.UHostInstanceSet)
				if !ok {
					ctx.LogError(fmt.Sprintf("uhost %s may not exist", idname))
					continue
				}
				if ins.IsolationGroup == "" {
					fmt.Fprintf(ctx.ProgressWriter(), "uhost %s doesn't attached any isolation group\n", idname)
					continue
				}
				req.GroupId = sdk.String(ins.IsolationGroup)
				req.UHostId = &id
				_, err = client.LeaveIsolationGroup(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				fmt.Fprintf(ctx.ProgressWriter(), "uhost %s detached from isolation group %s\n", idname, ins.IsolationGroup)
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "leave-isolation-group", Status: "Detached"})
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringSliceVar(&uhostIds, "uhost-id", nil, "Required. Resource ID of uhosts to be detech from its isolation group")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	ctx.BindZone(cmd, req)
	cmd.MarkFlagRequired("uhost-id")
	command.SetCompletion(cmd, "uhost-id", func() []string {
		return getUhostList(ctx, nil, *req.ProjectId, *req.Region, *req.Zone)
	})
	return cmd
}
