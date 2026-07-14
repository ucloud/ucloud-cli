package gssh

import (
	"fmt"
	"net"
	"strings"

	"github.com/spf13/cobra"

	pathxsdk "github.com/ucloud/ucloud-sdk-go/services/pathx"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newCreate ucloud gssh create
func newCreate(ctx *cli.Context) *cobra.Command {
	var targetIP *net.IP
	client := cli.NewServiceClient(ctx, pathxsdk.NewClient)
	req := client.NewCreateGlobalSSHInstanceRequest()
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create GlobalSSH instance",
		Long:    "Create GlobalSSH instance",
		Example: "ucloud gssh create --location Washington --target-ip 8.8.8.8",
		Run: func(cmd *cobra.Command, args []string) {
			port := *req.Port
			for code, area := range areaCodeMap {
				if area == *req.AreaCode {
					*req.AreaCode = code
				}
			}
			if port < 1 || port > 65535 || port == 80 || port == 443 || port == 65123 {
				fmt.Fprintln(ctx.ProgressWriter(), "The port number should be between 1 and 65535, and cannot be 80, 443 or 65123")
				return
			}
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			req.TargetIP = sdk.String(targetIP.String())
			resp, err := client.CreateGlobalSSHInstance(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "gssh[%s] created\n", resp.InstanceId)
			ctx.EmitResult(cli.OpResultRow{ResourceID: resp.InstanceId, Action: "create", Status: "Created"})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	req.AreaCode = flags.String("location", "", "Required. Location of the source server. See 'ucloud gssh location'")
	targetIP = flags.IP("target-ip", nil, "Required. IP of the source server. Required")
	ctx.BindProjectID(cmd, req)
	req.Port = flags.Int("port", 22, "Optional. Port of The SSH service between 1 and 65535. Do not use ports such as 80, 443 or 65123.")
	req.Remark = flags.String("remark", "", "Optional. Remark of your GlobalSSH.")
	req.ChargeType = flags.String("charge-type", "Month", "Optional.'Year',pay yearly;'Month',pay monthly;'Dynamic', pay hourly(requires access)")
	req.Quantity = flags.Int("quantity", 1, "Optional. The duration of the instance. N years/months.")
	req.InstanceType = flags.String("instance-type", "", "Optional. Possible values: 'Ultimate','Enterprise', 'Basic', 'Free'(Default value)")
	req.ForwardRegion = flags.String("forward-region", "", "Optional. You can select one of 'cn-bj2','cn-sh2','cn-gd' When instance-type is 'Basic'")
	req.BandwidthPackage = flags.Int("bandwidth-package", 0, "Optional. You can set one of 0, 20, 40 When instance-type is 'Ultimate'")
	cmd.MarkFlagRequired("location")
	cmd.MarkFlagRequired("target-ip")
	command.SetFlagValues(cmd, "location", "LosAngeles", "Singapore", "Lagos", "HongKong", "Tokyo", "Washington", "Frankfurt")
	command.SetFlagValues(cmd, "charge-type", "Month", "Year", "Dynamic", "Trial")
	command.SetFlagValues(cmd, "bandwidth-package", "0", "20", "40")
	command.SetFlagValues(cmd, "forward-region", "cn-bj2", "cn-sh2", "cn-gd")
	command.SetFlagValues(cmd, "instance-type", "Free", "Basic", "Enterprise", "Ultimate")
	ctx.SetCompletion(cmd, "target-ip", func() []string {
		eips := getAllEip(ctx, *req.ProjectId, ctx.DefaultRegion(), nil, nil)
		for idx, eip := range eips {
			eips[idx] = strings.SplitN(eip, "/", 2)[1]
		}
		return eips
	})
	return cmd
}
