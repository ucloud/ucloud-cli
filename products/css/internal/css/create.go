package css

import (
	"encoding/base64"
	"fmt"

	"github.com/spf13/cobra"

	uessdk "github.com/ucloud/ucloud-sdk-go/services/ues"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newCreate ucloud css create
func newCreate(ctx *cli.Context) *cobra.Command {
	var async *bool
	var servicePasswd *string
	client := cli.NewServiceClient(ctx, uessdk.NewClient)
	req := client.NewCreateUESInstanceRequest()
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create UES instance",
		Long:  "Create UES instance",
		Run: func(cmd *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			// 服务密码需 base64 编码后提交；未指定时使用默认密码 changeme
			passwd := *servicePasswd
			if passwd == "" {
				passwd = "changeme"
			}
			req.ServicePasswd = sdk.String(base64.StdEncoding.EncodeToString([]byte(passwd)))
			resp, err := client.CreateUESInstance(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			text := fmt.Sprintf("ues[%s] is creating", resp.InstanceId)
			if *async {
				fmt.Fprintln(w, text)
			} else {
				ctx.PollerTo(w, describeUESInstanceByID(ctx)).Spoll(resp.InstanceId, text, []string{STATE_RUNNING, STATE_ABNORMAL})
			}
			ctx.EmitResult(cli.OpResultRow{ResourceID: resp.InstanceId, Action: "create", Status: "Creating"})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.InstanceName = flags.String("name", "", "Required. Instance name")
	req.AppVersion = flags.String("app-version", "", "Required. Application version, e.g. 7.10.2")
	req.AppName = flags.String("app-name", "elasticsearch", "Optional. Application name, default elasticsearch")
	req.NodeConf = flags.String("node-conf", "", "Required. Node configuration identifier")
	req.NodeDiskConf = flags.String("node-disk-conf", "CLOUD_RSSD", "Required. Node disk type")
	req.NodeDiskSize = flags.Int("node-disk-size-gb", 100, "Optional. Node disk size in GB, default 100")
	req.NodeSize = flags.Int("node-count", 3, "Optional. Node count, default 3")
	req.KibanaNodeConf = flags.String("kibana-node-conf", "", "Required. Kibana node configuration")
	req.KibanaNodeDiskConf = flags.String("kibana-disk-conf", "CLOUD_RSSD", "Required. Kibana disk type")
	req.VPCId = flags.String("vpc-id", "", "Required. VPC ID")
	req.SubnetId = flags.String("subnet-id", "", "Required. Subnet ID")
	req.ServiceUserName = flags.String("service-username", "", "Optional. Service username. elasticsearch default 'elastic'; OpenSearch fixed 'admin'")
	servicePasswd = flags.String("service-passwd", "", "Optional. Service password, default 'changeme'")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")
	req.ChargeType = flags.String("charge-type", "Month", "Optional. 'Year', 'Month', or 'Dynamic', default Month")
	req.Quantity = flags.Int("quantity", 1, "Optional. Purchase duration, default 1")
	req.BusinessId = flags.String("business-id", "", "Optional. Business group ID")
	req.Remark = flags.String("remark", "", "Optional. Remark")
	async = flags.Bool("async", false, "Optional. Do not wait for creation to finish")

	command.SetFlagValues(cmd, "charge-type", "Month", "Year", "Dynamic")

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("app-version")
	cmd.MarkFlagRequired("node-conf")
	cmd.MarkFlagRequired("node-disk-conf")
	cmd.MarkFlagRequired("kibana-node-conf")
	cmd.MarkFlagRequired("kibana-disk-conf")
	cmd.MarkFlagRequired("vpc-id")
	cmd.MarkFlagRequired("subnet-id")

	return cmd
}
