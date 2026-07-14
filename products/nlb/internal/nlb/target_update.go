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

// parseTargetUpdate parses a "--target" value for the update command.
//
// Format: "targetId[:weight|enabled[:enabled]]"
//
//	"nrs-abc:5"        → weight=5
//	"nrs-abc:false"    → enabled=false
//	"nrs-abc:5:false"  → weight=5, enabled=false
//
// The second field is auto-detected: if it looks like a bool (true/false/0/1)
// it is treated as enabled; otherwise it is treated as weight.
func parseTargetUpdate(s string) (id string, t nlbsdk.UpdateNLBTargetsAttributeParamTargets, err error) {
	parts := strings.Split(s, ":")
	if len(parts) < 1 || len(parts) > 3 {
		return "", t, fmt.Errorf("invalid --target %q, expected \"targetId[:weight|enabled[:enabled]]\"", s)
	}
	if len(parts) < 2 {
		return "", t, fmt.Errorf("invalid --target %q: at least weight or enabled must be provided", s)
	}
	id = parts[0]
	t.Id = sdk.String(id)

	isBool := func(s string) bool { return s == "true" || s == "false" || s == "0" || s == "1" }

	p1 := parts[1]
	hasSecondField := len(parts) >= 3

	if isBool(p1) {
		// "nrs-abc:false" or "nrs-abc:false:..." — p1 is enabled
		switch p1 {
		case "true", "1":
			t.Enabled = sdk.Bool(true)
		case "false", "0":
			t.Enabled = sdk.Bool(false)
		}
		if hasSecondField {
			return "", t, fmt.Errorf("invalid --target %q: too many fields for \"targetId:enabled\"", s)
		}
	} else {
		// "nrs-abc:5" or "nrs-abc:5:false" — p1 is weight
		w, err := strconv.Atoi(p1)
		if err != nil {
			return "", t, fmt.Errorf("invalid weight/enabled %q in --target %q", p1, s)
		}
		t.Weight = sdk.Int(w)
		if hasSecondField {
			switch parts[2] {
			case "true", "1":
				t.Enabled = sdk.Bool(true)
			case "false", "0":
				t.Enabled = sdk.Bool(false)
			default:
				return "", t, fmt.Errorf("invalid enabled value %q in --target %q, expected true/false", parts[2], s)
			}
		}
	}
	return id, t, nil
}

// newTargetUpdate implements `nlb target update`.
func newTargetUpdate(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, nlbsdk.NewClient)
	req := client.NewUpdateNLBTargetsAttributeRequest()

	var nlbID, listenerID string
	var targets []string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update backend targets of an NLB listener",
		Long:  "Update the weight and/or enabled state of one or more NLB targets. Each target can have different values.",
		Run: func(c *cobra.Command, args []string) {
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			req.NLBId = sdk.String(ctx.PickResourceID(nlbID))
			req.ListenerId = sdk.String(ctx.PickResourceID(listenerID))

			updates := make([]nlbsdk.UpdateNLBTargetsAttributeParamTargets, 0, len(targets))
			ids := make([]string, 0, len(targets))
			for _, raw := range targets {
				id, u, err := parseTargetUpdate(raw)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				id = ctx.PickResourceID(id)
				u.Id = sdk.String(id)
				ids = append(ids, id)
				updates = append(updates, u)
			}
			req.Targets = updates

			if _, err := client.UpdateNLBTargetsAttribute(req); err != nil {
				ctx.HandleError(err)
				return
			}
			results := make([]cli.OpResultRow, 0, len(ids))
			for _, id := range ids {
				fmt.Fprintf(ctx.ProgressWriter(), "nlb-target[%s] updated\n", id)
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "update-target", Status: "Updated"})
			}
			ctx.EmitResult(results...)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)

	flags.StringVar(&nlbID, resourceIDFlag, "", "Required. Resource ID of the NLB instance.")
	flags.StringVar(&listenerID, "listener-id", "", "Required. Resource ID of the listener.")
	flags.StringSliceVar(&targets, "target", nil, "Required. Repeatable. Target as \"targetId:weight|enabled[:enabled]\", e.g. \"nrs-abc:5\" or \"nrs-abc:false\" or \"nrs-abc:5:false\".")

	cmd.MarkFlagRequired(resourceIDFlag)
	cmd.MarkFlagRequired("listener-id")
	cmd.MarkFlagRequired("target")
	command.SetCompletion(cmd, resourceIDFlag, func() []string {
		return getAllNLBIDNames(ctx, derefStr(req.ProjectId), derefStr(req.Region))
	})
	command.SetCompletion(cmd, "listener-id", func() []string {
		return getAllListenerIDNames(ctx, nlbID, derefStr(req.ProjectId), derefStr(req.Region))
	})
	command.SetCompletion(cmd, "target", func() []string {
		return getAllTargetIDNames(ctx, nlbID, listenerID, derefStr(req.ProjectId), derefStr(req.Region))
	})

	return cmd
}
