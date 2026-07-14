package redis

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/umem"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

type flushParams struct {
	flushType string
	dbNum     int
	region    string
	zone      string
	projectID string
}

// newFlush returns ucloud redis flush.
func newFlush(ctx *cli.Context) *cobra.Command {
	var idNames []string
	var p flushParams
	cmd := &cobra.Command{
		Use:     "flush",
		Short:   "Clear data of redis instances",
		Long:    "Clear data of redis instances. Master-replica instances call FlushallURedisGroup, distributed instances call RemoveUDRedisData",
		Example: "ucloud redis flush --umem-id uredis-rl5xuxx/testcli1 --flush-type FlushAll",
		Run: func(c *cobra.Command, args []string) {
			results := []cli.OpResultRow{}
			for _, idname := range idNames {
				id := ctx.PickResourceID(idname)
				mode, err := describeRedisMode(ctx, id)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				switch mode {
				case redisModeMasterReplica:
					if flushMasterReplica(ctx, &p, id) {
						results = append(results, cli.OpResultRow{ResourceID: id, Action: "flush", Status: "Flushed"})
					}
				case redisModeDistributed:
					if flushDistributed(ctx, &p, id) {
						results = append(results, cli.OpResultRow{ResourceID: id, Action: "flush", Status: "Flushed"})
					}
				default:
					fmt.Fprintf(ctx.ProgressWriter(), "redis[%s] unknown resource type, skip\n", idname)
				}
			}
			ctx.EmitResult(results...)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&idNames, "umem-id", nil, "Required. Resource ID of redis instances to flush data")
	flags.StringVar(&p.flushType, "flush-type", "FlushAll", "Optional. FlushType of redis flush. Only for master-replica instances. Accept values: 'FlushAll', 'FlushDb'")
	flags.IntVar(&p.dbNum, "db-num", 0, "Optional. DbNum to flush. Only used when flush-type is FlushDb for master-replica instances")
	flags.StringVar(&p.region, "region", ctx.DefaultRegion(), "Optional. Override default region for this command invocation, see 'ucloud region'")
	flags.StringVar(&p.zone, "zone", ctx.DefaultZone(), "Optional. Override default availability zone for this command invocation, see 'ucloud region'")
	flags.StringVar(&p.projectID, "project-id", ctx.DefaultProjectID(), "Optional. Override default project-id for this command invocation, see 'ucloud project list'")

	command.SetCompletion(cmd, "region", ctx.RegionList)
	command.SetCompletion(cmd, "zone", func() []string { return ctx.ZoneList(p.region) })
	command.SetCompletion(cmd, "project-id", ctx.ProjectList)
	command.SetCompletion(cmd, "umem-id", func() []string {
		return getIDList(ctx, p.projectID, p.region)
	})

	cmd.MarkFlagRequired("umem-id")

	return cmd
}

func flushMasterReplica(ctx *cli.Context, p *flushParams, id string) bool {
	client := cli.NewServiceClient(ctx, umem.NewClient)
	req := client.NewFlushallURedisGroupRequest()
	req.Region = &p.region
	req.Zone = &p.zone
	req.ProjectId = &p.projectID
	req.GroupId = &id
	req.FlushType = &p.flushType
	if p.flushType == "FlushDb" {
		req.DbNum = sdk.Int(p.dbNum)
	}
	_, err := client.FlushallURedisGroup(req)
	if err != nil {
		ctx.HandleError(err)
		return false
	}
	fmt.Fprintf(ctx.ProgressWriter(), "redis[%s] data flushed\n", id)
	return true
}

func flushDistributed(ctx *cli.Context, p *flushParams, id string) bool {
	client := cli.NewServiceClient(ctx, umem.NewClient)
	req := client.NewRemoveUDRedisDataRequest()
	req.Region = &p.region
	req.Zone = &p.zone
	req.ProjectId = &p.projectID
	req.SpaceId = &id
	_, err := client.RemoveUDRedisData(req)
	if err != nil {
		ctx.HandleError(err)
		return false
	}
	fmt.Fprintf(ctx.ProgressWriter(), "redis[%s] data flushed\n", id)
	return true
}
