package uhost

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	uhostsdk "github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/services/unet"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newClone ucloud uhost clone
func newClone(ctx *cli.Context) *cobra.Command {
	var uhostID *string
	var async *bool

	var password string
	var keyPairId string

	client := cli.NewServiceClient(ctx, uhostsdk.NewClient)
	unetClient := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewCreateUHostInstanceRequest()
	cmd := &cobra.Command{
		Use:   "clone",
		Short: "Create an uhost with the same configuration as another uhost, excluding bound eip and udisk",
		Long:  "Create an uhost with the same configuration as another uhost, excluding bound eip and udisk",
		Run: func(com *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			if len(password) > 0 {
				req.LoginMode = sdk.String("Password")
				req.KeyPairId = nil
				req.Password = sdk.String(password)
			} else if len(keyPairId) > 0 {
				req.LoginMode = sdk.String("KeyPair")
				req.KeyPairId = sdk.String(keyPairId)
				req.Password = nil
			} else {
				fmt.Fprintln(ctx.Err(), errors.New("password or key-pair-id is required"))
				return
			}
			*uhostID = ctx.PickResourceID(*uhostID)
			queryReq := client.NewDescribeUHostInstanceRequest()
			queryReq.ProjectId = req.ProjectId
			queryReq.Region = req.Region
			queryReq.Zone = req.Zone
			queryReq.UHostIds = []string{*uhostID}
			queryResp, err := client.DescribeUHostInstance(queryReq)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if len(queryResp.UHostSet) < 1 {
				fmt.Fprintln(ctx.Err(), fmt.Errorf("uhost[%s] not exist", *uhostID))
				return
			}
			if queryResp.UHostSet[0].SecGroupInstance == true {
				fmt.Fprintln(ctx.Err(), fmt.Errorf("uhost[%s] is in security groups, it is not allowed to clone", *uhostID))
				return
			}
			queryFirewallReq := unetClient.NewDescribeFirewallRequest()
			queryFirewallReq.ProjectId = req.ProjectId
			queryFirewallReq.Region = req.Region
			queryFirewallReq.ResourceId = uhostID
			queryFirewallReq.ResourceType = sdk.String("uhost")

			firewallResp, err := unetClient.DescribeFirewall(queryFirewallReq)
			if err != nil {
				ctx.HandleError(err)
				return
			}

			if len(firewallResp.DataSet) == 1 {
				req.SecurityGroupId = &firewallResp.DataSet[0].FWId
			}

			uhostIns := queryResp.UHostSet[0]

			req.ImageId = &uhostIns.BasicImageId
			req.CPU = &uhostIns.CPU
			req.Memory = &uhostIns.Memory
			for _, ip := range uhostIns.IPSet {
				if ip.Type == "Private" {
					req.VPCId = &ip.VPCId
					req.SubnetId = &ip.SubnetId
				}
			}
			req.ChargeType = &uhostIns.ChargeType
			req.UHostType = &uhostIns.UHostType
			req.NetCapability = &uhostIns.NetCapability

			for _, disk := range uhostIns.DiskSet {
				item := uhostsdk.UHostDisk{
					Size:   sdk.Int(disk.Size),
					Type:   sdk.String(disk.DiskType),
					IsBoot: sdk.String(disk.IsBoot),
				}
				if disk.BackupType != "" {
					item.BackupType = sdk.String(disk.BackupType)
				}
				req.Disks = append(req.Disks, item)
			}
			req.Tag = &uhostIns.Tag
			resp, err := client.CreateUHostInstance(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if len(resp.UHostIds) == 1 {
				text := fmt.Sprintf("cloned uhost:[%s] is initializing", resp.UHostIds[0])
				if *async {
					fmt.Fprintln(w, text)
				} else {
					ctx.PollerTo(w, describeUHostByID(ctx, *req.ProjectId, *req.Region, *req.Zone)).Spoll(resp.UHostIds[0], text, []string{HOST_RUNNING, HOST_FAIL})
				}
			} else {
				ctx.HandleError(fmt.Errorf("expect uhost count 1, accept %d", len(resp.UHostIds)))
				return
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	uhostID = flags.String("uhost-id", "", "Required. Resource ID of the uhost to clone from")
	flags.StringVar(&password, "password", "", "Optional. Password of the uhost user(root/ubuntu)")
	flags.StringVar(&keyPairId, "key-pair-id", "", "Optional. Resource ID of ssh key pair. See 'ucloud api --Action DescribeUHostKeyPairs' Where both password and key-pair-id are set, the key-pair-id is ignored")

	req.Name = flags.String("name", "", "Optional. Name of the uhost to clone")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")
	async = flags.Bool("async", false, "Optional. Do not wait for the long-running operation to finish.")
	command.SetCompletion(cmd, "uhost-id", func() []string {
		return getUhostList(ctx, []string{HOST_RUNNING, HOST_STOPPED}, *req.ProjectId, *req.Region, *req.Zone)
	})
	cmd.MarkFlagRequired("uhost-id")
	return cmd
}
