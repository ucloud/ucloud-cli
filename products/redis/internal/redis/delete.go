package redis

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/umem"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

type deleteParams struct {
	region    string
	zone      string
	projectID string
}

// newDelete returns ucloud redis delete.
func newDelete(ctx *cli.Context) *cobra.Command {
	var idNames []string
	var p deleteParams
	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete redis instances",
		Long:    "Delete redis instances",
		Example: "ucloud redis delete --umem-id uredis-rl5xuxx/testcli1,uredis-xsdfa/testcli2",
		Run: func(c *cobra.Command, args []string) {
			results := []cli.OpResultRow{}
			for _, idname := range idNames {
				id := ctx.PickResourceID(idname)
				if strings.HasPrefix(id, "uredis") || strings.HasPrefix(id, "uhredis") || strings.HasPrefix(id, "uregionredis") {
					if deleteMasterReplica(ctx, &p, id) {
						results = append(results, cli.OpResultRow{ResourceID: id, Action: "delete", Status: "Deleted"})
					}
				} else if strings.HasPrefix(id, "udredis") {
					if deleteDistributed(ctx, &p, id) {
						results = append(results, cli.OpResultRow{ResourceID: id, Action: "delete", Status: "Deleted"})
					}
				} else {
					fmt.Fprintf(ctx.ProgressWriter(), "redis[%s] unknown id prefix, skip\n", idname)
				}
			}
			ctx.EmitResult(results...)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&idNames, "umem-id", nil, "Required. Resource ID of redis instances to delete")
	flags.StringVar(&p.region, "region", ctx.DefaultRegion(), "Optional. Override default region for this command invocation, see 'ucloud region'")
	flags.StringVar(&p.zone, "zone", ctx.DefaultZone(), "Optional. Override default availability zone for this command invocation, see 'ucloud region'")
	flags.StringVar(&p.projectID, "project-id", ctx.DefaultProjectID(), "Optional. Override default project-id for this command invocation, see 'ucloud project list'")
	command.SetCompletion(cmd, "region", ctx.RegionList)
	command.SetCompletion(cmd, "zone", func() []string { return ctx.ZoneList(p.region) })
	command.SetCompletion(cmd, "project-id", ctx.ProjectList)

	cmd.MarkFlagRequired("umem-id")

	command.SetCompletion(cmd, "umem-id", func() []string {
		return getIDList(ctx, p.projectID, p.region)
	})

	return cmd
}

func deleteMasterReplica(ctx *cli.Context, p *deleteParams, id string) bool {
	client := cli.NewServiceClient(ctx, umem.NewClient)
	req := client.NewDeleteURedisGroupRequest()
	req.Region = &p.region
	req.ProjectId = &p.projectID
	req.GroupId = &id
	_, err := client.DeleteURedisGroup(req)
	if err != nil {
		ctx.HandleError(err)
		return false
	}
	fmt.Fprintf(ctx.ProgressWriter(), "redis[%s] deleted\n", id)
	return true
}

func deleteDistributed(ctx *cli.Context, p *deleteParams, id string) bool {
	client := cli.NewServiceClient(ctx, umem.NewClient)
	req := client.NewDeleteUMemSpaceRequest()
	req.Region = &p.region
	req.Zone = &p.zone
	req.ProjectId = &p.projectID
	req.SpaceId = &id
	_, err := client.DeleteUMemSpace(req)
	if err != nil {
		ctx.HandleError(err)
		return false
	}
	fmt.Fprintf(ctx.ProgressWriter(), "redis[%s] deleted\n", id)
	return true
}
