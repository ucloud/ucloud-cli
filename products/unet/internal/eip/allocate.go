package eip

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/unet"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newAllocate ucloud eip allocate
func newAllocate(ctx *cli.Context) *cobra.Command {
	var count *int
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewAllocateEIPRequest()
	cmd := &cobra.Command{
		Use:     "allocate",
		Short:   "Allocate EIP",
		Long:    "Allocate EIP",
		Example: "ucloud eip allocate --line BGP --bandwidth-mb 2",
		Run: func(cmd *cobra.Command, args []string) {
			if *req.OperatorName == "" {
				*req.OperatorName = getEIPLine(*req.Region)
			}
			results := []cli.OpResultRow{}
			for i := 0; i < *count; i++ {
				resp, err := client.AllocateEIP(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				for _, eip := range resp.EIPSet {
					fmt.Fprintf(ctx.ProgressWriter(), "allocate EIP[%s] ", eip.EIPId)
					for _, ip := range eip.EIPAddr {
						fmt.Fprintf(ctx.ProgressWriter(), "IP:%s  Line:%s \n", ip.IP, ip.OperatorName)
					}
					results = append(results, cli.OpResultRow{ResourceID: eip.EIPId, Action: "allocate", Status: "Allocated"})
				}
			}
			ctx.EmitResult(results...)
		},
	}
	cmd.Flags().SortFlags = false
	req.Bandwidth = cmd.Flags().Int("bandwidth-mb", 0, "Required. Bandwidth(Unit:Mbps).The range of value related to network charge mode. By traffic [1, 200]; by bandwidth [1,800] (Unit: Mbps); it could be 0 if the eip belong to the shared bandwidth")
	req.OperatorName = cmd.Flags().String("line", "", "Optional. 'BGP' or 'International'. 'BGP' could be set in China mainland regions, such as cn-bj2 etc. 'International' could be set in the regions beyond mainland, such as hk, tw-kh, us-ws etc.")
	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	req.PayMode = cmd.Flags().String("traffic-mode", "Bandwidth", "Optional. traffic-mode is an enumeration value. 'Traffic','Bandwidth' or 'ShareBandwidth'")
	req.ShareBandwidthId = cmd.Flags().String("share-bandwidth-id", "", "Optional. ShareBandwidthId, required only when traffic-mode is 'ShareBandwidth'")
	req.Quantity = cmd.Flags().Int("quantity", 1, "Optional. The duration of the instance. N years/months.")
	req.ChargeType = cmd.Flags().String("charge-type", "Month", "Optional. Enumeration value.'Year',pay yearly;'Month',pay monthly;'Dynamic', pay hourly(requires permission),'Trial', free trial(need permission)")
	req.Tag = cmd.Flags().String("group", "Default", "Optional. Group of your EIP.")
	req.Name = cmd.Flags().String("name", "EIP", "Optional. Name of your EIP.")
	req.Remark = cmd.Flags().String("remark", "", "Optional. Remark of your EIP.")
	count = cmd.Flags().Int("count", 1, "Optional. Count of EIP to allocate")

	command.SetFlagValues(cmd, "line", "BGP", "International")
	command.SetFlagValues(cmd, "traffic-mode", "Bandwidth", "Traffic", "ShareBandwidth")
	command.SetFlagValues(cmd, "charge-type", "Month", "Year", "Dynamic", "Trial")
	cmd.MarkFlagRequired("bandwidth-mb")
	return cmd
}

// getEIPLine returns the default EIP line for a region. Product-local copy of
// cmd/util.go getEIPLine (domain logic, D-D: COPIED into the product, never
// promoted to platform). "cn" regions default to BGP, others to International.
func getEIPLine(region string) (line string) {
	if strings.HasPrefix(region, "cn") {
		line = "BGP"
	} else {
		line = "International"
	}
	return
}
