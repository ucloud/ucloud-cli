package redis

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/umem"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

type resizeParams struct {
	region    string
	zone      string
	projectID string
	size      int
	blockID   string
}

// newResize returns ucloud redis resize.
func newResize(ctx *cli.Context) *cobra.Command {
	var idNames []string
	var p resizeParams
	cmd := &cobra.Command{
		Use:     "resize",
		Short:   "Resize redis instances",
		Long:    "Resize redis instances. Master-replica instances call ResizeURedisGroup, distributed instances call ResizeUDRedisBlockSize for the specified block",
		Example: "ucloud redis resize --umem-id uredis-rl5xuxx/testcli1 --size-gb 4",
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
					if resizeMasterReplica(ctx, &p, id) {
						results = append(results, cli.OpResultRow{ResourceID: id, Action: "resize", Status: "Resized"})
					}
				case redisModeDistributed:
					if p.blockID == "" {
						fmt.Fprintf(ctx.ProgressWriter(), "redis[%s] --block-id is required for distributed redis\n", idname)
						continue
					}
					if resizeDistributed(ctx, &p, id) {
						results = append(results, cli.OpResultRow{ResourceID: id, Action: "resize", Status: "Resized"})
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

	flags.StringSliceVar(&idNames, "umem-id", nil, "Required. Resource ID of redis instances to resize")
	flags.IntVar(&p.size, "size-gb", 0, "Required. Target memory size in GB")
	flags.StringVar(&p.blockID, "block-id", "", "Required for distributed redis. Block ID to resize")
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
	cmd.MarkFlagRequired("size-gb")

	return cmd
}

func resizeMasterReplica(ctx *cli.Context, p *resizeParams, id string) bool {
	client := cli.NewServiceClient(ctx, umem.NewClient)
	req := client.NewResizeURedisGroupRequest()
	req.Region = &p.region
	req.ProjectId = &p.projectID
	req.GroupId = &id
	req.Size = &p.size
	_, err := client.ResizeURedisGroup(req)
	if err != nil {
		ctx.HandleError(err)
		return false
	}
	fmt.Fprintf(ctx.ProgressWriter(), "redis[%s] resized\n", id)
	return true
}

func resizeDistributed(ctx *cli.Context, p *resizeParams, id string) bool {
	client := cli.NewServiceClient(ctx, umem.NewClient)
	req := client.NewResizeUDRedisBlockSizeRequest()
	req.Region = &p.region
	req.Zone = &p.zone
	req.ProjectId = &p.projectID
	req.SpaceId = &id
	req.BlockId = &p.blockID
	req.BlockSize = &p.size
	_, err := client.ResizeUDRedisBlockSize(req)
	if err != nil {
		ctx.HandleError(err)
		return false
	}
	fmt.Fprintf(ctx.ProgressWriter(), "redis[%s] resized\n", id)
	return true
}
