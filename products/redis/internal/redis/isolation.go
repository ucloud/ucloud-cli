package redis

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/umem"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

type isolationParams struct {
	opType    string
	region    string
	zone      string
	projectID string
}

// newIsolation returns ucloud redis isolation.
func newIsolation(ctx *cli.Context) *cobra.Command {
	var idNames []string
	var p isolationParams
	cmd := &cobra.Command{
		Use:     "isolation",
		Short:   "Open or close redis instances of master-replica type",
		Long:    "Open or close redis instances of master-replica type. Only master-replica instances are supported. --type open opens redis, --type close closes redis",
		Example: "ucloud redis isolation --umem-id uredis-rl5xuxx/testcli1 --type open",
		Run: func(c *cobra.Command, args []string) {
			results := []cli.OpResultRow{}
			transformType := "UNBind"
			action := "close"
			if p.opType == "open" {
				transformType = "Bind"
				action = "open"
			}
			for _, idname := range idNames {
				id := ctx.PickResourceID(idname)
				mode, err := describeRedisMode(ctx, id)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				if mode != redisModeMasterReplica {
					fmt.Fprintf(ctx.ProgressWriter(), "redis[%s] is not master-replica type, skip\n", idname)
					continue
				}
				if isolationMasterReplica(ctx, &p, id, transformType, action) {
					results = append(results, cli.OpResultRow{ResourceID: id, Action: action, Status: "Done"})
				}
			}
			ctx.EmitResult(results...)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&idNames, "umem-id", nil, "Required. Resource ID of redis instances to open or close")
	flags.StringVar(&p.opType, "type", "close", "Required. Operation type of redis isolation. Accept values: 'open' or 'close'")
	flags.StringVar(&p.region, "region", ctx.DefaultRegion(), "Optional. Override default region for this command invocation, see 'ucloud region'")
	flags.StringVar(&p.zone, "zone", ctx.DefaultZone(), "Optional. Override default availability zone for this command invocation, see 'ucloud region'")
	flags.StringVar(&p.projectID, "project-id", ctx.DefaultProjectID(), "Optional. Override default project-id for this command invocation, see 'ucloud project list'")

	command.SetFlagValues(cmd, "type", "open", "close")
	command.SetCompletion(cmd, "region", ctx.RegionList)
	command.SetCompletion(cmd, "zone", func() []string { return ctx.ZoneList(p.region) })
	command.SetCompletion(cmd, "project-id", ctx.ProjectList)
	command.SetCompletion(cmd, "umem-id", func() []string {
		return getIDList(ctx, p.projectID, p.region)
	})

	cmd.MarkFlagRequired("umem-id")
	cmd.MarkFlagRequired("type")

	return cmd
}

func isolationMasterReplica(ctx *cli.Context, p *isolationParams, id, transformType, action string) bool {
	client := cli.NewServiceClient(ctx, umem.NewClient)
	req := client.NewISolationURedisGroupRequest()
	req.Region = &p.region
	req.Zone = &p.zone
	req.ProjectId = &p.projectID
	req.GroupId = &id
	req.TransformType = &transformType
	_, err := client.ISolationURedisGroup(req)
	if err != nil {
		ctx.HandleError(err)
		return false
	}
	fmt.Fprintf(ctx.ProgressWriter(), "redis[%s] %sed\n", id, action)
	return true
}
