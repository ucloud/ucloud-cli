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
	biz    *base.Client
	config *base.AggConfig
}

// Deps carries constructor arguments for NewContext.
type Deps struct {
	In     io.Reader
	Out    io.Writer
	Err    io.Writer
	Format OutputFormat
	Biz    *base.Client
	Config *base.AggConfig
}

// NewContext constructs a Context from the provided Deps.
func NewContext(d Deps) *Context {
	return &Context{
		in:     d.In,
		out:    d.Out,
		err:    d.Err,
		format: d.Format,
		biz:    d.Biz,
		config: d.Config,
	}
}

// Out returns the output writer.
func (c *Context) Out() io.Writer { return c.out }

// In returns the input reader.
func (c *Context) In() io.Reader { return c.in }

// Format returns the output format requested for this invocation.
func (c *Context) Format() OutputFormat { return c.format }
