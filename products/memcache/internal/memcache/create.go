package memcache

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/umem"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newCreate returns ucloud memcache create.
func newCreate(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, umem.NewClient)
	req := client.NewCreateUMemcacheGroupRequest()
	var region, zone, projectID string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create memcache instance",
		Long:  "Create memcache instance",
		Run: func(c *cobra.Command, args []string) {
			if *req.Size > 32 || *req.Size < 1 {
				fmt.Fprintln(ctx.ProgressWriter(), "size-gb should be between 1 and 32")
				return
			}
			if err := fillDefaultVPCAndSubnet(ctx, req.VPCId, req.SubnetId, *req.ProjectId, *req.Region, *req.Zone); err != nil {
				fmt.Fprintln(ctx.ProgressWriter(), err)
				return
			}
			resp, err := client.CreateUMemcacheGroup(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "memcache[%s] created\n", resp.GroupId)
			ctx.EmitResult(cli.OpResultRow{ResourceID: resp.GroupId, Action: "create", Status: "Created"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.Name = flags.String("name", "", "Required. Name of memcache instance to create")
	req.Size = flags.Int("size-gb", 1, "Optional. Memory size of memcache instance. Unit GB. Accpet values:1,2,4,8,16,32")
	req.VPCId = flags.String("vpc-id", "", "Optional. VPC ID. See 'ucloud vpc list'")
	req.SubnetId = flags.String("subnet-id", "", "Optional. Subnet ID. See 'ucloud subnet list'")
	flags.StringVar(&region, "region", ctx.DefaultRegion(), "Optional. Override default region for this command invocation, see 'ucloud region'")
	flags.StringVar(&zone, "zone", ctx.DefaultZone(), "Optional. Override default availability zone for this command invocation, see 'ucloud region'")
	flags.StringVar(&projectID, "project-id", ctx.DefaultProjectID(), "Optional. Override default project-id for this command invocation, see 'ucloud project list'")
	req.ChargeType = flags.String("charge-type", "Month", "Optional. Enumeration value.'Year',pay yearly;'Month',pay monthly; 'Dynamic', pay hourly; 'Trial', free trial(need permission)")
	req.Quantity = flags.Int("quantity", 1, "Optional. The duration of the instance. N years/months.")
	req.Tag = flags.String("group", "", "Optional. Business group")

	req.Region = &region
	req.Zone = &zone
	req.ProjectId = &projectID

	command.SetCompletion(cmd, "region", ctx.RegionList)
	command.SetCompletion(cmd, "zone", func() []string { return ctx.ZoneList(region) })
	command.SetCompletion(cmd, "project-id", ctx.ProjectList)
	command.SetFlagValues(cmd, "size-gb", "1", "2", "4", "8", "16", "32")
	command.SetFlagValues(cmd, "charge-type", "Month", "Dynamic", "Year")
	command.SetCompletion(cmd, "vpc-id", func() []string {
		return getAllVPCIdNames(ctx, projectID, region)
	})
	command.SetCompletion(cmd, "subnet-id", func() []string {
		return getAllSubnetIDNames(ctx, *req.VPCId, projectID, region)
	})

	cmd.MarkFlagRequired("name")

	return cmd
}
