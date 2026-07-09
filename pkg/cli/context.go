package cli

import (
	"fmt"
	"io"
	"sync/atomic"

	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/command"
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
	in               io.Reader
	out              io.Writer
	err              io.Writer
	format           OutputFormat
	defaultsProvider func() command.Defaults

	// Completion candidate providers injected by the host so that bind
	// helpers can register dynamic completion without pkg/command importing
	// cmd or base.
	regionList  func() []string
	zoneList    func(region string) []string
	projectList func() []string
	// allRegions is the runtime all-region lister (returns an error, unlike the
	// completion providers) for non-standard flags like uhost --all-region.
	allRegions func() ([]string, error)

	clientConfig    func() *ucloud.Config
	buildCredential func() *auth.Credential
	attachHandlers  func(ucloud.ServiceClient)

	handleError func(io.Writer, error)
	logInfo     func(...string)
	logPrint    func(io.Writer, ...string)
	logWarn     func(io.Writer, ...string)
	logError    func(io.Writer, ...string)
	logFilePath func() string
	newPoller   func(func(string, *request.CommonBase) (interface{}, error), io.Writer) Poller

	// errCount tallies HandleError calls this invocation so the host (cmd) can
	// set a non-zero exit code when any product error occurred (aws/gcloud
	// convention). Atomic because product commands can call HandleError from
	// concurrent goroutines (e.g. uhost create's per-instance EIP binding in the
	// count>5 fan-out).
	errCount int32
}

// Deps carries constructor arguments for NewContext.
type Deps struct {
	In     io.Reader
	Out    io.Writer
	Err    io.Writer
	Format OutputFormat

	DefaultsProvider func() command.Defaults
	RegionList       func() []string
	ZoneList         func(region string) []string
	ProjectList      func() []string
	AllRegions       func() ([]string, error)

	ClientConfig    func() *ucloud.Config
	BuildCredential func() *auth.Credential
	AttachHandlers  func(ucloud.ServiceClient)

	HandleError func(io.Writer, error)
	LogInfo     func(...string)
	LogPrint    func(io.Writer, ...string)
	LogWarn     func(io.Writer, ...string)
	LogError    func(io.Writer, ...string)
	LogFilePath func() string
	NewPoller   func(func(string, *request.CommonBase) (interface{}, error), io.Writer) Poller
}

// NewContext constructs a Context from the provided Deps.
func NewContext(d Deps) *Context {
	if d.Out == nil {
		d.Out = io.Discard
	}
	if d.Err == nil {
		d.Err = io.Discard
	}
	if d.DefaultsProvider == nil {
		d.DefaultsProvider = func() command.Defaults { return command.Defaults{} }
	}
	if d.HandleError == nil {
		d.HandleError = func(w io.Writer, err error) {
			if err != nil {
				fmt.Fprintln(w, err)
			}
		}
	}
	if d.LogInfo == nil {
		d.LogInfo = func(...string) {}
	}
	if d.LogPrint == nil {
		d.LogPrint = func(w io.Writer, logs ...string) {
			for _, line := range logs {
				fmt.Fprintln(w, line)
			}
		}
	}
	if d.LogWarn == nil {
		d.LogWarn = d.LogPrint
	}
	if d.LogError == nil {
		d.LogError = d.LogPrint
	}
	if d.LogFilePath == nil {
		d.LogFilePath = func() string { return "" }
	}
	if d.NewPoller == nil {
		d.NewPoller = NewPoller
	}
	return &Context{
		in:               d.In,
		out:              d.Out,
		err:              d.Err,
		format:           d.Format,
		defaultsProvider: d.DefaultsProvider,
		regionList:       d.RegionList,
		zoneList:         d.ZoneList,
		projectList:      d.ProjectList,
		allRegions:       d.AllRegions,
		clientConfig:     d.ClientConfig,
		buildCredential:  d.BuildCredential,
		attachHandlers:   d.AttachHandlers,
		handleError:      d.HandleError,
		logInfo:          d.LogInfo,
		logPrint:         d.LogPrint,
		logWarn:          d.LogWarn,
		logError:         d.LogError,
		logFilePath:      d.LogFilePath,
		newPoller:        d.NewPoller,
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

// Failed reports whether any error was recorded via HandleError this invocation.
func (c *Context) Failed() bool { return atomic.LoadInt32(&c.errCount) > 0 }
