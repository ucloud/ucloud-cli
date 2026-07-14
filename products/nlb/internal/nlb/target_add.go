package nlb

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	nlbsdk "github.com/ucloud/ucloud-sdk-go/services/nlb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// parseTarget parses a "--target" value.
//
// Non-IP format: "resourceType:resourceId:port[:weight[:enabled]]"
//
//	e.g. "UHost:uhost-abc:80" or "UHost:uhost-abc:80:5:false"
//
// IP format:     "IP:ip:port[:weight]:vpcId[:subnetId[:enabled]]"
//
//	e.g. "IP:10.0.0.1:80:vpc-xxx" or "IP:10.0.0.1:80:1:vpc-xxx:subnet-yyy:false"
func parseTarget(s string) (nlbsdk.AddNLBTargetsParamTargets, error) {
	parts := strings.Split(s, ":")
	if len(parts) < 3 || len(parts) > 7 {
		return nlbsdk.AddNLBTargetsParamTargets{}, fmt.Errorf(
			"invalid --target %q, expected \"resourceType:resourceId:port[:weight[:enabled]]\" or \"IP:ip:port[:weight]:vpcId[:subnetId[:enabled]]\"", s)
	}

	parseEnabled := func(idx int) (*bool, error) {
		if idx >= len(parts) {
			return sdk.Bool(true), nil
		}
		switch parts[idx] {
		case "true", "1":
			return sdk.Bool(true), nil
		case "false", "0":
			return sdk.Bool(false), nil
		default:
			return nil, fmt.Errorf("invalid enabled value %q in --target %q, expected true/false", parts[idx], s)
		}
	}

	resourceType := parts[0]
	resourceID := parts[1]
	port, err := strconv.Atoi(parts[2])
	if err != nil {
		return nlbsdk.AddNLBTargetsParamTargets{}, fmt.Errorf("invalid port in --target %q: %w", s, err)
	}

	t := nlbsdk.AddNLBTargetsParamTargets{
		Port:         sdk.Int(port),
		ResourceType: sdk.String(resourceType),
	}

	isBool := func(s string) bool { return s == "true" || s == "false" || s == "0" || s == "1" }

	if resourceType == "IP" {
		t.ResourceIP = sdk.String(resourceID)
		if len(parts) < 4 {
			t.Weight = sdk.Int(1)
			t.Enabled = sdk.Bool(true)
			return t, fmt.Errorf("invalid --target %q: IP type requires vpcId: \"IP:ip:port[:weight]:vpcId[:subnetId[:enabled]]\"", s)
		}
		// IP format: IP:ip:port [weight?] vpcId [subnetId?] [enabled?]
		// If parts[3] is a number → weight at [3], vpcId at [4]
		// If parts[3] is not a number → vpcId at [3] (weight=1)
		// When no explicit weight and the last part looks like a bool,
		// treat it as enabled, not subnetId.
		if _, err := strconv.Atoi(parts[3]); err == nil {
			// Has explicit weight at [3]
			w, _ := strconv.Atoi(parts[3])
			t.Weight = sdk.Int(w)
			t.VPCId = sdk.String(parts[4])
			if len(parts) >= 6 {
				t.SubnetId = sdk.String(parts[5])
			}
			enabled, err := parseEnabled(6)
			if err != nil {
				return nlbsdk.AddNLBTargetsParamTargets{}, err
			}
			t.Enabled = enabled
		} else {
			// No explicit weight, parts[3] is vpcId
			t.Weight = sdk.Int(1)
			t.VPCId = sdk.String(parts[3])
			remaining := parts[4:]
			if len(remaining) == 0 {
				t.Enabled = sdk.Bool(true)
			} else if len(remaining) == 1 {
				if isBool(remaining[0]) {
					enabled, _ := parseEnabled(4) // parts[4] is enabled
					t.Enabled = enabled
				} else {
					t.SubnetId = sdk.String(remaining[0])
					t.Enabled = sdk.Bool(true)
				}
			} else if len(remaining) == 2 {
				t.SubnetId = sdk.String(remaining[0])
				enabled, err := parseEnabled(5)
				if err != nil {
					return nlbsdk.AddNLBTargetsParamTargets{}, err
				}
				t.Enabled = enabled
			}
		}
	} else {
		t.ResourceId = sdk.String(resourceID)
		// Non-IP: resourceType:resourceId:port[:weight[:enabled]]
		weight := 1
		if len(parts) >= 4 {
			weight, err = strconv.Atoi(parts[3])
			if err != nil {
				return nlbsdk.AddNLBTargetsParamTargets{}, fmt.Errorf("invalid weight in --target %q: %w", s, err)
			}
		}
		t.Weight = sdk.Int(weight)
		enabled, err := parseEnabled(4)
		if err != nil {
			return nlbsdk.AddNLBTargetsParamTargets{}, err
		}
		t.Enabled = enabled
	}
	return t, nil
}

// newTargetAdd implements `nlb target add`.
func newTargetAdd(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, nlbsdk.NewClient)
	req := client.NewAddNLBTargetsRequest()

	var nlbID, listenerID string
	var targets []string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add backend targets to an NLB listener",
		Long:  "Add one or more backend service nodes to an NLB listener. Supports mixed resource types in one call.",
		Run: func(c *cobra.Command, args []string) {
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			req.NLBId = sdk.String(ctx.PickResourceID(nlbID))
			req.ListenerId = sdk.String(ctx.PickResourceID(listenerID))

			parsed := make([]nlbsdk.AddNLBTargetsParamTargets, 0, len(targets))
			for _, t := range targets {
				pt, err := parseTarget(t)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				parsed = append(parsed, pt)
			}
			req.Targets = parsed

			resp, err := client.AddNLBTargets(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			results := make([]cli.OpResultRow, 0, len(resp.Targets))
			for _, t := range resp.Targets {
				fmt.Fprintf(ctx.ProgressWriter(), "nlb-target[%s] added\n", t.Id)
				results = append(results, cli.OpResultRow{ResourceID: t.Id, Action: "add-target", Status: "Added"})
			}
			ctx.EmitResult(results...)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)

	flags.StringVar(&nlbID, resourceIDFlag, "", "Required. Resource ID of the NLB instance.")
	flags.StringVar(&listenerID, "listener-id", "", "Required. Resource ID of the listener to add targets to.")
	flags.StringSliceVar(&targets, "target", nil,
		"Required. Repeatable. Non-IP: \"resourceType:resourceId:port[:weight[:enabled]]\". IP: \"IP:ip:port[:weight]:vpcId[:subnetId[:enabled]]\".")

	command.SetCompletion(cmd, resourceIDFlag, func() []string {
		return getAllNLBIDNames(ctx, derefStr(req.ProjectId), derefStr(req.Region))
	})
	command.SetCompletion(cmd, "listener-id", func() []string {
		return getAllListenerIDNames(ctx, nlbID, derefStr(req.ProjectId), derefStr(req.Region))
	})

	cmd.MarkFlagRequired(resourceIDFlag)
	cmd.MarkFlagRequired("listener-id")
	cmd.MarkFlagRequired("target")

	return cmd
}
