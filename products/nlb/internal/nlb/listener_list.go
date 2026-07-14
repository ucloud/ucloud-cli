package nlb

import (
	"fmt"

	"github.com/spf13/cobra"

	nlbsdk "github.com/ucloud/ucloud-sdk-go/services/nlb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newListenerList implements `nlb listener list`.
func newListenerList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, nlbsdk.NewClient)
	req := client.NewDescribeNLBListenersRequest()

	var nlbID, listenerID string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List listeners of an NLB instance",
		Long:  "List the listeners of the specified NLB instance, along with each listener's backend targets.",
		Run: func(c *cobra.Command, args []string) {
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			req.NLBId = sdk.String(ctx.PickResourceID(nlbID))
			if id := ctx.PickResourceID(listenerID); id != "" {
				req.ListenerId = sdk.String(id)
			}
			resp, err := client.DescribeNLBListeners(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			rows := make([]ListenerRow, 0, len(resp.Listeners))
			for _, l := range resp.Listeners {
				rows = append(rows, ListenerRow{
					ListenerID:         l.ListenerId,
					Name:               l.Name,
					Protocol:           l.Protocol,
					Scheduler:          l.Scheduler,
					PortRange:          fmt.Sprintf("%d-%d", l.StartPort, l.EndPort),
					ForwardSrcIPMethod: l.ForwardSrcIPMethod,
					State:              l.State,
					StickinessTimeout:  l.StickinessTimeout,
					HealthCheckType:    l.HealthCheckConfig.Type,
					HealthCheckPort:    l.HealthCheckConfig.Port,
						HealthCheckReqMsg:  l.HealthCheckConfig.ReqMsg,
						HealthCheckResMsg:  l.HealthCheckConfig.ResMsg,
				})
			}
			ctx.PrintList(rows)

			for _, l := range resp.Listeners {
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
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)

	flags.StringVar(&nlbID, resourceIDFlag, "", "Required. Resource ID of the NLB instance.")
	flags.StringVar(&listenerID, "listener-id", "", "Optional. List only the specified listener.")
	req.Offset = flags.Int("offset", 0, "Optional. Offset.")
	req.Limit = flags.Int("limit", 100, "Optional. Limit.")

	cmd.MarkFlagRequired(resourceIDFlag)
	command.SetCompletion(cmd, resourceIDFlag, func() []string {
		return getAllNLBIDNames(ctx, derefStr(req.ProjectId), derefStr(req.Region))
	})
	command.SetCompletion(cmd, "listener-id", func() []string {
		return getAllListenerIDNames(ctx, nlbID, derefStr(req.ProjectId), derefStr(req.Region))
	})

	return cmd
}
