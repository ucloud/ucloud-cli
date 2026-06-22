package cli

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/base"
	"github.com/ucloud/ucloud-cli/pkg/command"
	"github.com/ucloud/ucloud-cli/pkg/ui"
)

// defaults reads the per-invocation region/zone/project defaults from the agg config.
func (c *Context) defaults() command.Defaults {
	if c.config == nil {
		return command.Defaults{}
	}
	return command.Defaults{Region: c.config.Region, Zone: c.config.Zone, ProjectID: c.config.ProjectID}
}

// SetCompletion forwards to command.SetCompletion.
func (c *Context) SetCompletion(cmd *cobra.Command, name string, fn func() []string) {
	command.SetCompletion(cmd, name, fn)
}

// SetFlagValues forwards to command.SetFlagValues.
func (c *Context) SetFlagValues(cmd *cobra.Command, name string, values ...string) {
	command.SetFlagValues(cmd, name, values...)
}

// BindRegion binds --region using ctx defaults + injected region completion provider.
func (c *Context) BindRegion(cmd *cobra.Command, req request.Common) {
	command.BindRegion(cmd, req, c.defaults(), c.regionList)
}

// BindZone binds --zone using ctx defaults + injected zone completion provider.
func (c *Context) BindZone(cmd *cobra.Command, req request.Common) {
	command.BindZone(cmd, req, c.defaults(), c.zoneList)
}

// BindZoneEmpty binds --zone with empty default + injected zone completion provider.
func (c *Context) BindZoneEmpty(cmd *cobra.Command, req request.Common) {
	command.BindZoneEmpty(cmd, req, c.zoneList)
}

// BindProjectID binds --project-id using ctx defaults + injected project completion provider.
func (c *Context) BindProjectID(cmd *cobra.Command, req request.Common) {
	command.BindProjectID(cmd, req, c.defaults(), c.projectList)
}

// BindLimit binds --limit into req via reflection.
func (c *Context) BindLimit(cmd *cobra.Command, req interface{}) { command.BindLimit(cmd, req) }

// BindOffset binds --offset into req via reflection.
func (c *Context) BindOffset(cmd *cobra.Command, req interface{}) { command.BindOffset(cmd, req) }

// BindChargeType binds --charge-type into req via reflection.
func (c *Context) BindChargeType(cmd *cobra.Command, req interface{}) {
	command.BindChargeType(cmd, req)
}

// BindQuantity binds --quantity into req via reflection.
func (c *Context) BindQuantity(cmd *cobra.Command, req interface{}) { command.BindQuantity(cmd, req) }

// PrintList renders dataSet to the ctx writer in the ctx format.
func (c *Context) PrintList(dataSet interface{}) {
	ui.Printer{Out: c.out, Format: ui.Format(c.format)}.PrintList(dataSet)
}

// PrintJSON renders dataSet as JSON to the ctx writer.
func (c *Context) PrintJSON(dataSet interface{}) error { return base.PrintJSON(dataSet, c.out) }

// Confirm prompts the user for a yes/no confirmation on the ctx streams.
func (c *Context) Confirm(yes bool, text string) bool { return ui.Confirm(c.in, c.out, yes, text) }

// HandleError logs err in the standard CLI error format.
func (c *Context) HandleError(err error) { base.HandleError(err) }

// PickResourceID extracts the resource ID from a "resourceID/name" string.
func (c *Context) PickResourceID(s string) string { return PickResourceID(s) }

// Poller wraps base.NewSpoller bound to ctx's writer.
func (c *Context) Poller(describeFunc func(string, *request.CommonBase) (interface{}, error)) *base.Poller {
	return base.NewSpoller(describeFunc, c.out)
}
