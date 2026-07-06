package eip

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/unet"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newList ucloud eip list
func newList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewDescribeEIPRequest()
	fetchAll := false
	pageOff := false
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List all EIP instances",
		Long:    `List all EIP instances`,
		Example: "ucloud eip list",
		Run: func(cmd *cobra.Command, args []string) {
			var eipList []unet.UnetEIPSet
			if fetchAll || pageOff {
				list, err := fetchAllEip(ctx, *req.ProjectId, *req.Region)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				eipList = list
			} else {
				resp, err := client.DescribeEIP(req)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				eipList = resp.EIPSet
			}

			list := make([]EIPRow, 0)
			for _, eip := range eipList {
				row := EIPRow{}
				row.Name = eip.Name
				for _, ip := range eip.EIPAddr {
					row.IP += ip.IP + " " + ip.OperatorName + "   "
				}
				row.ResourceID = eip.EIPId
				row.Group = eip.Tag
				row.ChargeMode = eip.PayMode
				row.Bandwidth = strconv.Itoa(eip.Bandwidth) + "Mb"
				if eip.Resource.ResourceID != "" {
					row.BindResource = fmt.Sprintf("%s|%s(%s)", eip.Resource.ResourceName, eip.Resource.ResourceID, eip.Resource.ResourceType)
				}
				row.Status = eip.Status
				row.ExpirationTime = time.Unix(int64(eip.ExpireTime), 0).Format("2006-01-02")
				list = append(list, row)
			}
			ctx.PrintList(list)
		},
	}

	flags := cmd.Flags()
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	req.Offset = flags.Int("offset", 0, "Optional. Offset default 0")
	req.Limit = flags.Int("limit", 50, "Optional. Limit default 50, max value 100")
	flags.BoolVar(&fetchAll, "list-all", false, "List all eip")
	flags.BoolVar(&pageOff, "page-off", false, "Optional. Paging or not. Accept values: true or false")
	command.SetFlagValues(cmd, "list-all", "true", "false")
	flags.MarkDeprecated("list-all", "please use '--page-off' instead")

	return cmd
}
