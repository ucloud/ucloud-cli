package onboarding

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/udb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newList implements `example list`.
//
// Platform APIs exercised: cli.NewServiceClient, ctx.BindCommonParams (the
// aggregate binder), ctx.PickResourceID, ctx.PrintList, ctx.HandleError,
// command.SetCompletion.
func newList(ctx *cli.Context) *cobra.Command {
	// One authed SDK client per command, built from the constructor. The Run
	// func only needs this to type-check; the example is never executed.
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewDescribeUDBInstanceRequest()

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List example instances",
		Long:  "List example instances in the active region/zone/project.",
		Run: func(c *cobra.Command, args []string) {
			if req.DBId != nil && *req.DBId != "" {
				*req.DBId = ctx.PickResourceID(*req.DBId)
			}
			resp, err := client.DescribeUDBInstance(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			rows := make([]instanceRow, 0, len(resp.DataSet))
			for _, ins := range resp.DataSet {
				rows = append(rows, instanceRow{
					ResourceID: ins.DBId,
					Name:       ins.Name,
					Zone:       ins.Zone,
					Mode:       ins.InstanceMode,
					Spec:       fmt.Sprintf("%s|%dMB|%dGB", ins.DBTypeId, ins.MemoryLimit, ins.DiskSpace),
					Status:     ins.State,
				})
			}
			ctx.PrintList(rows)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	// Optional resource-id filter, named after the product.
	req.DBId = flags.String(resourceIDFlag, "", "Optional. List only the specified instance.")
	req.ClassType = sdk.String("sql")

	// One call binds region/zone/project plus --limit/--offset (present on this
	// request) with the per-invocation defaults and the injected completion
	// providers. This is the primary, preferred binder.
	ctx.BindCommonParams(cmd, req)

	// Dynamic completion for the resource-id flag.
	command.SetCompletion(cmd, resourceIDFlag, func() []string {
		return listResourceIDs(ctx, nil, derefStr(req.Region), derefStr(req.Zone), derefStr(req.ProjectId))
	})

	return cmd
}
