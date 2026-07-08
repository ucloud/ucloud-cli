package redis

import (
	"fmt"
	"unicode/utf8"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/umem"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

type createParams struct {
	name       string
	password   string
	size       int
	region     string
	zone       string
	projectID  string
	chargeType string
	quantity   int
	group      string
	vpcID      string
	subnetID   string
	version    string
	blockCnt   int
	proxySize  int
}

// newCreate returns ucloud redis create.
func newCreate(ctx *cli.Context) *cobra.Command {
	var redisType string
	var p createParams
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create redis instance",
		Long:  "Create redis instance",
		Run: func(c *cobra.Command, args []string) {
			if l := utf8.RuneCountInString(p.name); l < 6 || l > 63 {
				fmt.Fprintln(ctx.ProgressWriter(), "length of name should be between 6 and 63")
				return
			}
			if p.password != "" {
				if l := len(p.password); l < 6 || l > 36 {
					fmt.Fprintln(ctx.ProgressWriter(), "length of password should be between 6 and 36")
					return
				}
			}
			if err := fillDefaultVPCAndSubnet(ctx, &p.vpcID, &p.subnetID, p.projectID, p.region, p.zone); err != nil {
				fmt.Fprintln(ctx.ProgressWriter(), err)
				return
			}
			switch redisType {
			case "master-replica":
				createMasterReplica(ctx, &p)
			case "distributed":
				createDistributed(ctx, &p)
			default:
				fmt.Fprintf(ctx.ProgressWriter(), "unknow redis type[%s], it's should be 'master-replica' or 'distributed'\n", redisType)
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&p.name, "name", "", "Required. Name of the redis to create. Range of the name length is [6,63]")
	flags.StringVar(&redisType, "type", "", "Required. Type of the redis. Accept values:'master-replica','distributed'")
	flags.IntVar(&p.size, "size-gb", 2, "Optional. Memory size. Default value 2GB. Unit GB")
	flags.StringVar(&p.version, "version", "6.0", "Optional. Version of redis. Accept values: '4.0', '5.0', '6.0', '7.0'")
	flags.StringVar(&p.vpcID, "vpc-id", "", "Optional. VPC ID. This field is required under VPC2.0. See 'ucloud vpc list'")
	flags.StringVar(&p.subnetID, "subnet-id", "", "Optional. Subnet ID. This field is required under VPC2.0. See 'ucloud subnet list'")
	flags.StringVar(&p.password, "password", "", "Optional. Password of redis to create. Range of the password length is [6,36] and the password can only contain letters and numbers")

	flags.IntVar(&p.blockCnt, "block-cnt", 2, "Optional. Block count. Default value 2(for distributed redis type).")
	flags.IntVar(&p.proxySize, "proxy-size", 2, "Optional. Proxy size. Default value 2(for distributed redis type) Unit Core")

	flags.StringVar(&p.region, "region", ctx.DefaultRegion(), "Optional. Override default region for this command invocation, see 'ucloud region'")
	flags.StringVar(&p.zone, "zone", ctx.DefaultZone(), "Optional. Override default availability zone for this command invocation, see 'ucloud region'")
	flags.StringVar(&p.projectID, "project-id", ctx.DefaultProjectID(), "Optional. Override default project-id for this command invocation, see 'ucloud project list'")
	flags.StringVar(&p.chargeType, "charge-type", "Month", "Optional. Enumeration value.'Year',pay yearly;'Month',pay monthly; 'Dynamic', pay hourly; 'Trial', free trial(need permission)")
	flags.IntVar(&p.quantity, "quantity", 1, "Optional. The duration of the instance. N years/months.")
	flags.StringVar(&p.group, "group", "", "Optional. Business group")

	command.SetCompletion(cmd, "region", ctx.RegionList)
	command.SetCompletion(cmd, "zone", func() []string { return ctx.ZoneList(p.region) })
	command.SetCompletion(cmd, "project-id", ctx.ProjectList)
	command.SetFlagValues(cmd, "version", "4.0", "5.0", "6.0", "7.0")
	command.SetFlagValues(cmd, "type", "master-replica", "distributed")
	command.SetFlagValues(cmd, "charge-type", "Month", "Dynamic", "Year")
	command.SetCompletion(cmd, "vpc-id", func() []string {
		return getAllVPCIdNames(ctx, p.projectID, p.region)
	})
	command.SetCompletion(cmd, "subnet-id", func() []string {
		return getAllSubnetIDNames(ctx, p.vpcID, p.projectID, p.region)
	})

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("type")

	return cmd
}

func createMasterReplica(ctx *cli.Context, p *createParams) {
	client := cli.NewServiceClient(ctx, umem.NewClient)
	req := client.NewCreateURedisGroupRequest()
	req.Region = &p.region
	req.Zone = &p.zone
	req.ProjectId = &p.projectID
	req.Name = &p.name
	req.HighAvailability = sdk.String("enable")
	req.Size = &p.size
	req.Version = &p.version
	req.VPCId = &p.vpcID
	req.SubnetId = &p.subnetID
	req.ChargeType = &p.chargeType
	req.Quantity = &p.quantity
	req.Tag = &p.group
	if p.password != "" {
		req.Password = &p.password
	}

	resp, err := client.CreateURedisGroup(req)
	if err != nil {
		ctx.HandleError(err)
		return
	}
	fmt.Fprintf(ctx.ProgressWriter(), "redis[%s] created\n", resp.GroupId)
	ctx.EmitResult(cli.OpResultRow{ResourceID: resp.GroupId, Action: "create", Status: "Created"})
}

func createDistributed(ctx *cli.Context, p *createParams) {
	client := cli.NewServiceClient(ctx, umem.NewClient)
	req := client.NewCreateUMemSpaceRequest()
	req.Region = &p.region
	req.Zone = &p.zone
	req.ProjectId = &p.projectID
	req.Name = &p.name
	req.Protocol = sdk.String("redis")

	if p.blockCnt <= 0 {
		fmt.Fprintln(ctx.ProgressWriter(), "block-cnt should be greater than 0")
		return
	}
	if p.size%p.blockCnt != 0 {
		fmt.Fprintf(ctx.ProgressWriter(), "size-gb(%d) should be divisible by block-cnt(%d)\n", p.size, p.blockCnt)
		return
	}
	if p.proxySize%2 != 0 {
		fmt.Fprintf(ctx.ProgressWriter(), "proxy-size(%d) should be a multiple of 2\n", p.proxySize)
		return
	}

	req.BlockCnt = &p.blockCnt
	req.ProxySize = &p.proxySize
	req.Size = &p.size
	req.Version = &p.version
	req.VPCId = &p.vpcID
	req.SubnetId = &p.subnetID
	req.ChargeType = &p.chargeType
	req.Quantity = &p.quantity
	req.Tag = &p.group
	if p.password != "" {
		req.Password = &p.password
	}

	resp, err := client.CreateUMemSpace(req)
	if err != nil {
		ctx.HandleError(err)
		return
	}
	fmt.Fprintf(ctx.ProgressWriter(), "redis[%s] created\n", resp.SpaceId)
	ctx.EmitResult(cli.OpResultRow{ResourceID: resp.SpaceId, Action: "create", Status: "Created"})
}
