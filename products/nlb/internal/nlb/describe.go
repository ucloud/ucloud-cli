package nlb

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	nlbsdk "github.com/ucloud/ucloud-sdk-go/services/nlb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newDescribe implements `nlb describe`.
func newDescribe(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, nlbsdk.NewClient)
	req := client.NewDescribeNetworkLoadBalancersRequest()

	var nlbID string

	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Show details of one NLB instance",
		Long:  "Show the full attribute/value detail of a single NLB instance, along with its listeners and each listener's backend targets.",
		Run: func(c *cobra.Command, args []string) {
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			id := ctx.PickResourceID(nlbID)
			req.NLBIds = []string{id}
			req.ShowDetail = sdk.Bool(true)

			resp, err := client.DescribeNetworkLoadBalancers(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if len(resp.NLBs) == 0 {
				ctx.HandleError(fmt.Errorf("NLB instance %q not found", id))
				return
			}
			n := resp.NLBs[0]

			ips := make([]string, 0, len(n.IPInfos))
			for _, ip := range n.IPInfos {
				direction := "Forward"
				if ip.Type == 2 {
					direction = "Backward"
				}
				ips = append(ips, fmt.Sprintf("%s(%s,%s)", ip.IP, ip.AddressType, direction))
			}

			rows := []cli.DescribeRow{
				{Attribute: "ResourceID", Content: n.NLBId},
				{Attribute: "Name", Content: n.Name},
				{Attribute: "Status", Content: n.Status},
				{Attribute: "VPC", Content: n.VPCId},
				{Attribute: "Subnet", Content: n.SubnetId},
				{Attribute: "IPVersion", Content: n.IPVersion},
				{Attribute: "IP", Content: strings.Join(ips, ",")},
				{Attribute: "ForwardingMode", Content: n.ForwardingMode},
				{Attribute: "ChargeType", Content: n.ChargeType},
				{Attribute: "AutoRenew", Content: strconv.FormatBool(n.AutoRenewEnabled)},
				{Attribute: "PurchaseValue", Content: common.FormatDate(n.PurchaseValue)},
				{Attribute: "Group", Content: n.Tag},
				{Attribute: "Remark", Content: n.Remark},
				{Attribute: "ListenerCount", Content: fmt.Sprintf("%d", len(n.Listeners))},
				{Attribute: "CreationTime", Content: common.FormatDate(n.CreateTime)},
			}
			printDescribe(ctx, rows)

			if len(n.Listeners) > 0 {
				fmt.Fprintln(ctx.ProgressWriter(), "\nListeners:")
				details := make([]ListenerDetailRow, 0, len(n.Listeners))
				for _, l := range n.Listeners {
					details = append(details, ListenerDetailRow{
						ListenerID:         l.ListenerId,
						Name:               l.Name,
						Protocol:           l.Protocol,
						PortRange:          fmt.Sprintf("%d-%d", l.StartPort, l.EndPort),
						Scheduler:          l.Scheduler,
						ForwardSrcIPMethod: l.ForwardSrcIPMethod,
						StickinessTimeout:  l.StickinessTimeout,
						State:              l.State,
						HealthCheckType:    l.HealthCheckConfig.Type,
						HealthCheckPort:    l.HealthCheckConfig.Port,
						HealthCheckReqMsg:  l.HealthCheckConfig.ReqMsg,
						HealthCheckResMsg:  l.HealthCheckConfig.ResMsg,
						TargetCount:        len(l.Targets),
					})
				}
				ctx.PrintList(details)

				for _, l := range n.Listeners {
					fmt.Fprintf(ctx.ProgressWriter(), "\nTargets of %s(%s):\n", l.ListenerId, l.Name)
					targets := make([]TargetRow, 0, len(l.Targets))
					for _, t := range l.Targets {
						resourceID := t.ResourceId
						if resourceID == "" {
							resourceID = t.ResourceIP // IP-type targets carry the address here
						}
						targets = append(targets, TargetRow{
							TargetID:     t.Id,
							Name:         t.ResourceName,
							ResourceType: t.ResourceType,
							ResourceID:   resourceID,
							Port:         t.Port,
							Weight:       t.Weight,
							Enabled:      t.Enabled,
							State:        t.State,
						})
					}
					ctx.PrintList(targets)
				}
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&nlbID, resourceIDFlag, "", "Required. Resource ID of the NLB instance to describe.")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)

	cmd.MarkFlagRequired(resourceIDFlag)
	command.SetCompletion(cmd, resourceIDFlag, func() []string {
		return getAllNLBIDNames(ctx, derefStr(req.ProjectId), derefStr(req.Region))
	})

	return cmd
}

// printDescribe renders describe rows without column headers in table mode,
// printing each attribute/content pair as an aligned key-value row.
func printDescribe(ctx *cli.Context, rows []cli.DescribeRow) {
	if ctx.Format() != cli.OutputTable {
		ctx.PrintList(rows)
		return
	}
	maxWidth := 0
	for _, r := range rows {
		if len(r.Attribute) > maxWidth {
			maxWidth = len(r.Attribute)
		}
	}
	for _, r := range rows {
		fmt.Fprintf(ctx.Out(), "%-*s  %s\n", maxWidth, r.Attribute, r.Content)
	}
}
