package memcache

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/umem"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newDelete returns ucloud memcache delete.
func newDelete(ctx *cli.Context) *cobra.Command {
	var idNames []string
	client := cli.NewServiceClient(ctx, umem.NewClient)
	req := client.NewDeleteUMemcacheGroupRequest()
	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete memcache instances",
		Long:    "Delete memcache instances",
		Example: "ucloud memcache delete --umem-id umemcache-rl5xuxx/testcli1,umemcache-xsdfa/testcli2",
		Run: func(c *cobra.Command, args []string) {
			results := []cli.OpResultRow{}
			for _, idname := range idNames {
				id := ctx.PickResourceID(idname)
				req.GroupId = &id
				_, err := client.DeleteUMemcacheGroup(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				fmt.Fprintf(ctx.ProgressWriter(), "memcache[%s] deleted\n", idname)
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "delete", Status: "Deleted"})
			}
			ctx.EmitResult(results...)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&idNames, "umem-id", nil, "Required. Resource ID of memcache intances to delete")
	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	ctx.BindZoneEmpty(cmd, req)

	cmd.MarkFlagRequired("umem-id")

	command.SetCompletion(cmd, "umem-id", func() []string {
		return getIDList(ctx, *req.ProjectId, *req.Region)
	})

	return cmd
}
