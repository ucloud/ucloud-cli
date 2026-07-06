package onboarding

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/udb"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newDescribe implements `example describe`.
//
// Platform APIs exercised: cli.NewServiceClient, the non-aggregate binders
// (ctx.BindRegion / ctx.BindZone / ctx.BindProjectID — shown here once for the
// case where you want per-field control), cli.DescribeRow for detail rows,
// ctx.PrintList, ctx.PickResourceID, ctx.HandleError, command.SetCompletion.
func newDescribe(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewDescribeUDBInstanceRequest()

	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Show details of one example instance",
		Long:  "Show the full attribute/value detail of a single example instance.",
		Run: func(c *cobra.Command, args []string) {
			*req.DBId = ctx.PickResourceID(*req.DBId)
			resp, err := client.DescribeUDBInstance(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if len(resp.DataSet) == 0 {
				ctx.HandleError(fmt.Errorf("instance %q not found", *req.DBId))
				return
			}
			ins := resp.DataSet[0]

			// cli.DescribeRow renders a single resource as attribute/content
			// rows. In table mode the field names "Attribute" and "Content"
			// become the two column headers.
			rows := []cli.DescribeRow{
				{Attribute: "ResourceID", Content: ins.DBId},
				{Attribute: "Name", Content: ins.Name},
				{Attribute: "Zone", Content: ins.Zone},
				{Attribute: "Mode", Content: ins.InstanceMode},
				{Attribute: "Version", Content: ins.DBTypeId},
				{Attribute: "Memory(MB)", Content: fmt.Sprintf("%d", ins.MemoryLimit)},
				{Attribute: "Disk(GB)", Content: fmt.Sprintf("%d", ins.DiskSpace)},
				{Attribute: "VirtualIP", Content: ins.VirtualIP},
				{Attribute: "Port", Content: fmt.Sprintf("%d", ins.Port)},
				{Attribute: "Status", Content: ins.State},
			}
			ctx.PrintList(rows)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.DBId = flags.String(resourceIDFlag, "", "Required. Resource ID of the instance to describe.")

	// Non-aggregate binding: bind each common flag explicitly. Equivalent to
	// BindCommonParams for region/zone/project, shown here for the case where a
	// command needs to bind them individually (e.g. to interleave custom flags).
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)

	cmd.MarkFlagRequired(resourceIDFlag)
	command.SetCompletion(cmd, resourceIDFlag, func() []string {
		return listResourceIDs(ctx, nil, derefStr(req.Region), derefStr(req.Zone), derefStr(req.ProjectId))
	})

	return cmd
}
