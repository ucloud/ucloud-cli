package cli

import (
	"io"

	"github.com/ucloud/ucloud-cli/base"
)

// OutputFormat controls how command results are rendered.
type OutputFormat int

const (
	// OutputTable is the default tabular output (iota-zero value).
	OutputTable OutputFormat = iota
	// OutputJSON renders output as JSON.
	OutputJSON
	// OutputYAML renders output as YAML.
	OutputYAML
)

// Context is a per-invocation handle that product commands receive.
// It provides access to the authed SDK client, I/O streams, and
// common configuration. Heavy methods (NewServiceClient, PrintList,
// Confirm, Poller, flag binding) are added in later tasks (B2, B5, D1).
type Context struct {
	in     io.Reader
	out    io.Writer
	err    io.Writer
	format OutputFormat
	config *base.AggConfig

	// Completion candidate providers injected by the host so that bind
	// helpers can register dynamic completion without pkg/command importing
	// cmd or base.
	regionList  func() []string
	zoneList    func(region string) []string
	projectList func() []string
	// allRegions is the runtime all-region lister (returns an error, unlike the
	// completion providers) for non-standard flags like uhost --all-region.
	allRegions func() ([]string, error)
}

// Deps carries constructor arguments for NewContext.
type Deps struct {
	In     io.Reader
	Out    io.Writer
	Err    io.Writer
	Format OutputFormat
	Config *base.AggConfig

	RegionList  func() []string
	ZoneList    func(region string) []string
	ProjectList func() []string
	AllRegions  func() ([]string, error)
}

// NewContext constructs a Context from the provided Deps.
func NewContext(d Deps) *Context {
	return &Context{
		in:          d.In,
		out:         d.Out,
		err:         d.Err,
		format:      d.Format,
		config:      d.Config,
		regionList:  d.RegionList,
		zoneList:    d.ZoneList,
		projectList: d.ProjectList,
		allRegions:  d.AllRegions,
	}
}

// Out returns the output writer (stdout). Machine-readable results go here.
func (c *Context) Out() io.Writer { return c.out }

// Err returns the error/diagnostics writer (stderr). Human-facing narration
// and progress belong here so machine output on Out stays clean.
func (c *Context) Err() io.Writer { return c.err }

// In returns the input reader.
func (c *Context) In() io.Reader { return c.in }

// Format returns the output format requested for this invocation.
func (c *Context) Format() OutputFormat { return c.format }

// SetFormat overrides the output format. The host (cmd) calls this from its
// PersistentPreRun once --output has been parsed, because the Context is built
// at command-registration time, before cobra parses flags.
func (c *Context) SetFormat(f OutputFormat) { c.format = f }
