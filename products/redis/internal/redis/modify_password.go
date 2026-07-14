package redis

import (
	"encoding/base64"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/umem"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

type modifyPasswordParams struct {
	password  string
	region    string
	zone      string
	projectID string
}

// newModifyPassword returns ucloud redis modify-password.
func newModifyPassword(ctx *cli.Context) *cobra.Command {
	var idNames []string
	var p modifyPasswordParams
	cmd := &cobra.Command{
		Use:     "modify-password",
		Short:   "Modify redis instance password",
		Long:    "Modify redis instance password",
		Example: "ucloud redis modify-password --umem-id uredis-rl5xuxx/testcli1 --password newpassword",
		Run: func(c *cobra.Command, args []string) {
			results := []cli.OpResultRow{}
			encodedPassword := base64.StdEncoding.EncodeToString([]byte(p.password))
			for _, idname := range idNames {
				id := ctx.PickResourceID(idname)
				mode, err := describeRedisMode(ctx, id)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				switch mode {
				case redisModeMasterReplica:
					if modifyMasterReplicaPassword(ctx, &p, id, encodedPassword) {
						results = append(results, cli.OpResultRow{ResourceID: id, Action: "modify-password", Status: "Modified"})
					}
				case redisModeDistributed:
					if modifyDistributedPassword(ctx, &p, id, encodedPassword) {
						results = append(results, cli.OpResultRow{ResourceID: id, Action: "modify-password", Status: "Modified"})
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

	flags.StringSliceVar(&idNames, "umem-id", nil, "Required. Resource ID of redis instances to modify password")
	flags.StringVar(&p.password, "password", "", "Required. New password of the redis instance")
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
	cmd.MarkFlagRequired("password")

	return cmd
}

func modifyMasterReplicaPassword(ctx *cli.Context, p *modifyPasswordParams, id, encodedPassword string) bool {
	client := cli.NewServiceClient(ctx, umem.NewClient)
	req := client.NewModifyURedisGroupPasswordRequest()
	req.Region = &p.region
	req.ProjectId = &p.projectID
	req.Zone = &p.zone
	req.GroupId = &id
	req.Password = &encodedPassword
	_, err := client.ModifyURedisGroupPassword(req)
	if err != nil {
		ctx.HandleError(err)
		return false
	}
	fmt.Fprintf(ctx.ProgressWriter(), "redis[%s] password modified\n", id)
	return true
}

func modifyDistributedPassword(ctx *cli.Context, p *modifyPasswordParams, id, encodedPassword string) bool {
	client := cli.NewServiceClient(ctx, umem.NewClient)
	req := client.NewModifyUMemPasswordRequest()
	req.Region = &p.region
	req.Zone = &p.zone
	req.ProjectId = &p.projectID
	req.SpaceId = &id
	req.Password = &encodedPassword
	_, err := client.ModifyUMemPassword(req)
	if err != nil {
		ctx.HandleError(err)
		return false
	}
	fmt.Fprintf(ctx.ProgressWriter(), "redis[%s] password modified\n", id)
	return true
}
