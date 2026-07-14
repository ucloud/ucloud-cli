package redis

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/umem"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

type modifyNameParams struct {
	name      string
	region    string
	zone      string
	projectID string
}

// newModifyName returns ucloud redis modify-name.
func newModifyName(ctx *cli.Context) *cobra.Command {
	var idNames []string
	var p modifyNameParams
	cmd := &cobra.Command{
		Use:     "modify-name",
		Short:   "Modify redis instance name",
		Long:    "Modify redis instance name",
		Example: "ucloud redis modify-name --umem-id uredis-rl5xuxx/testcli1,uredis-xsdfa/testcli2 --name newname",
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
					if modifyMasterReplicaName(ctx, &p, id) {
						results = append(results, cli.OpResultRow{ResourceID: id, Action: "modify-name", Status: "Modified"})
					}
				case redisModeDistributed:
					if modifyDistributedName(ctx, &p, id) {
						results = append(results, cli.OpResultRow{ResourceID: id, Action: "modify-name", Status: "Modified"})
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

	flags.StringSliceVar(&idNames, "umem-id", nil, "Required. Resource ID of redis instances to modify name")
	flags.StringVar(&p.name, "name", "", "Required. New name of the redis instance")
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
	cmd.MarkFlagRequired("name")

	return cmd
}

func modifyMasterReplicaName(ctx *cli.Context, p *modifyNameParams, id string) bool {
	client := cli.NewServiceClient(ctx, umem.NewClient)
	req := client.NewModifyURedisGroupNameRequest()
	req.Region = &p.region
	req.ProjectId = &p.projectID
	req.GroupId = &id
	req.Name = &p.name
	_, err := client.ModifyURedisGroupName(req)
	if err != nil {
		ctx.HandleError(err)
		return false
	}
	fmt.Fprintf(ctx.ProgressWriter(), "redis[%s] name modified\n", id)
	return true
}

func modifyDistributedName(ctx *cli.Context, p *modifyNameParams, id string) bool {
	client := cli.NewServiceClient(ctx, umem.NewClient)
	req := client.NewModifyUMemSpaceNameRequest()
	req.Region = &p.region
	req.Zone = &p.zone
	req.ProjectId = &p.projectID
	req.SpaceId = &id
	req.Name = &p.name
	_, err := client.ModifyUMemSpaceName(req)
	if err != nil {
		ctx.HandleError(err)
		return false
	}
	fmt.Fprintf(ctx.ProgressWriter(), "redis[%s] name modified\n", id)
	return true
}
