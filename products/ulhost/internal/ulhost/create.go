package ulhost

import (
	"encoding/base64"
	"fmt"

	"github.com/spf13/cobra"

	ucompsharesdk "github.com/ucloud/ucloud-sdk-go/services/ucompshare"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newCreate ucloud ulhost create
func newCreate(ctx *cli.Context) *cobra.Command {
	var async bool
	client := cli.NewServiceClient(ctx, ucompsharesdk.NewClient)
	req := client.NewCreateULHostInstanceRequest()
	cmd := &cobra.Command{
		Use:          "create",
		Short:        "Create ULHost instance",
		Long:         "Create ULHost instance",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			req.ImageId = sdk.String(ctx.PickResourceID(*req.ImageId))
			req.VPCId = sdk.String(ctx.PickResourceID(*req.VPCId))
			req.SubnetId = sdk.String(ctx.PickResourceID(*req.SubnetId))
			req.SecurityGroupId = sdk.String(ctx.PickResourceID(*req.SecurityGroupId))
			// Encode password to base64
			if *req.Password != "" {
				req.Password = sdk.String(base64.StdEncoding.EncodeToString([]byte(*req.Password)))
			}

			resp, err := client.CreateULHostInstance(req)
			if err != nil {
				return err
			}
			w := ctx.ProgressWriter()
			text := fmt.Sprintf("ulhost[%s] is creating", resp.ULHostId)
			if async {
				fmt.Fprintln(w, text)
			} else {
				prog := ctx.NewProgress()
				block := prog.NewBlock()
				prog.Sspoll(sdescribeULHostByID(ctx), resp.ULHostId, text, []string{HOST_RUNNING, HOST_FAIL}, block, &req.CommonBase)
			}
			return nil
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	ctx.BindRegion(cmd, req)
	req.BundleId = flags.String("bundle-id", "", "Required. Bundle ID of the ULHost instance, e.g. ulh.c1m1s40b30t800")
	req.ImageId = flags.String("image-id", "", "Required. Image ID. See 'ucloud ulhost image list'")
	req.Password = flags.String("password", "", "Required. Password of the ulhost instance")
	req.Name = flags.String("name", "", "Optional. ULHost instance name")
	req.ChargeType = flags.String("charge-type", "Month", "Optional. 'Year', pay yearly; 'Month', pay monthly; default: Month")
	req.Quantity = flags.Int("quantity", 1, "Optional. The duration of the instance. N years/months.")
	req.SecurityGroupId = flags.String("security-group-id", "", "Optional. Firewall ID, default: Web recommended firewall")
	req.VPCId = flags.String("vpc-id", "", "Optional. VPC ID. Default VPC will be used if not specified")
	req.SubnetId = flags.String("subnet-id", "", "Optional. Subnet ID. Default subnet will be used if not specified")
	req.CouponId = flags.String("coupon-id", "", "Optional. Coupon ID")
	flags.BoolVar(&async, "async", false, "Optional. Do not wait for the long-running operation to finish.")

	command.SetFlagValues(cmd, "charge-type", "Month", "Year")
	command.SetCompletion(cmd, "bundle-id", func() []string {
		return getULHostBundleIDList(ctx, *req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("bundle-id")
	cmd.MarkFlagRequired("image-id")
	cmd.MarkFlagRequired("password")

	return cmd
}

// getULHostBundleIDList returns bundle ID completion candidates.
func getULHostBundleIDList(ctx *cli.Context, project, region string) []string {
	client := cli.NewServiceClient(ctx, ucompsharesdk.NewClient)
	req := client.NewDescribeULHostBundlesRequest()
	req.ProjectId = sdk.String(project)
	req.Region = sdk.String(region)
	resp, err := client.DescribeULHostBundles(req)
	if err != nil {
		return nil
	}
	list := []string{}
	for _, bundle := range resp.Bundles {
		desc := formatBundleInfo(bundle.CPU, bundle.Memory, bundle.SysDiskSpace, bundle.Bandwidth, bundle.TrafficPacket)
		list = append(list, bundle.BundleId+"/"+desc)
	}
	return list
}
