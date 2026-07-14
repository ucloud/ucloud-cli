package cli

import (
	"io"
	"sync/atomic"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/command"
	"github.com/ucloud/ucloud-cli/pkg/ui"
)

// defaults reads the per-invocation region/zone/project defaults from the host provider.
func (c *Context) defaults() command.Defaults {
	if c.defaultsProvider == nil {
		return command.Defaults{}
	}
	return c.defaultsProvider()
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

// BindGroup binds --group into req.Tag via reflection.
func (c *Context) BindGroup(cmd *cobra.Command, req interface{}) { command.BindGroup(cmd, req) }

// RegionList / ZoneList / ProjectList expose the injected completion providers
// for non-standard flags (e.g. --target-region) where the standard Bind*
// helpers don't apply. Nil-safe: return nil when no provider was injected.
func (c *Context) RegionList() []string {
	if c.regionList == nil {
		return nil
	}
	return c.regionList()
}

// ZoneList returns the availability zones for the given region.
func (c *Context) ZoneList(region string) []string {
	if c.zoneList == nil {
		return nil
	}
	return c.zoneList(region)
}

// ProjectList returns the project id/name completion candidates.
func (c *Context) ProjectList() []string {
	if c.projectList == nil {
		return nil
	}
	return c.projectList()
}

// DefaultRegion / DefaultProjectID expose the per-invocation config defaults
// (the same values Bind* helpers use) for hand-written flags where the standard
// Bind* helpers don't apply — e.g. a product command that needs the configured
// default region/project as a flag default but must NOT register region/project
// completion (mirrors the RegionList rationale). Nil-safe: empty when no config.
func (c *Context) DefaultRegion() string {
	return c.defaults().Region
}

// DefaultProjectID returns the per-invocation default project id from config.
func (c *Context) DefaultProjectID() string {
	return c.defaults().ProjectID
}

// DefaultZone returns the per-invocation default availability zone from config,
// for hand-written --zone flags that must NOT register zone completion (same
// rationale as DefaultRegion/DefaultProjectID). Nil-safe: empty when no config.
func (c *Context) DefaultZone() string {
	return c.defaults().Zone
}

// AllRegions returns every region the account can see, propagating the
// fetch error (unlike RegionList, which is for completion and drops it). Used
// by runtime fan-out flags such as uhost --all-region. Nil-safe.
func (c *Context) AllRegions() ([]string, error) {
	if c.allRegions == nil {
		return nil, nil
	}
	return c.allRegions()
}

// BindCommonParams binds all common flags in one call using ctx defaults +
// injected completion providers. It binds region/zone/project when req
// satisfies request.Common, plus --limit/--offset/--charge-type/--quantity for
// whichever of those fields exist on req (absent fields are skipped, no panic).
func (c *Context) BindCommonParams(cmd *cobra.Command, req interface{}) {
	command.BindCommonParams(cmd, req, c.defaults(), c.regionList, c.zoneList, c.projectList)
}

// PrintList renders dataSet to the ctx writer in the ctx format.
func (c *Context) PrintList(dataSet interface{}) {
	ui.Printer{Out: c.out, Format: ui.Format(c.format)}.PrintList(dataSet)
}

// PrintJSON renders dataSet as JSON to the ctx writer.
func (c *Context) PrintJSON(dataSet interface{}) error { return ui.PrintJSON(dataSet, c.out) }

// Confirm prompts for a yes/no confirmation, returning three outcomes:
// (true,nil) confirmed / (false,nil) declined / (false,err) non-interactive
// without --yes. Interactivity is judged on the input stream (c.in), so tests
// injecting a buffer are non-interactive. The prompt goes to the progress
// writer (stderr in json/yaml) to keep stdout machine-clean.
func (c *Context) Confirm(yes bool, text string) (bool, error) {
	return ui.Confirm(c.in, c.ProgressWriter(), yes, ui.IsReaderTTY(c.in), text)
}

// HandleError renders err (business RetCode / transport error) to stderr — never
// stdout, so machine output on stdout stays clean — and records it to the
// cli.log file / telemetry.
func (c *Context) HandleError(err error) {
	atomic.AddInt32(&c.errCount, 1)
	c.handleError(c.err, err)
}

// LogInfo / LogPrint / LogWarn / LogError forward to the platform logger
// (cli.log + optional telemetry, with redaction) for non-request product
// diagnostics (warnings, errors, status). API request logging is handled
// automatically by the platform SDK handler — products do NOT log requests
// themselves (see batch-1 plan Part 0 Task 0.2 / D-C).
// LogInfo writes to the log file only (no console). LogPrint/LogWarn/LogError
// send their console copy to stderr (ctx.Err), never stdout, so machine output
// on stdout stays clean; all four still record to cli.log / telemetry.
func (c *Context) LogInfo(logs ...string)  { c.logInfo(logs...) }
func (c *Context) LogPrint(logs ...string) { c.logPrint(c.err, logs...) }
func (c *Context) LogWarn(logs ...string)  { c.logWarn(c.err, logs...) }
func (c *Context) LogError(logs ...string) { c.logError(c.err, logs...) }

// LogFilePath returns the path of the CLI log file (e.g. for "check logs in …").
func (c *Context) LogFilePath() string { return c.logFilePath() }

// PickResourceID extracts the resource ID from a "resourceID/name" string.
func (c *Context) PickResourceID(s string) string { return PickResourceID(s) }

// Poller returns a platform poller bound to ctx's writer.
func (c *Context) Poller(describeFunc func(string, *request.CommonBase) (interface{}, error)) Poller {
	return c.newPoller(describeFunc, c.out)
}

// PollerTo wraps base.NewSpoller bound to an explicit writer, so callers can
// route progress narration to stderr (e.g. in json/yaml mode) while keeping
// machine output on stdout. Products cannot import base directly, so this
// exposes the writer-parameterized poller through the Context.
func (c *Context) PollerTo(w io.Writer, describeFunc func(string, *request.CommonBase) (interface{}, error)) Poller {
	return c.newPoller(describeFunc, w)
}
