package trace

// Stacktrace is record of golang stacktrack
type Stacktrace struct {
	Frames []*StacktraceFrame `json:"frames"`
}

// StacktraceFrame is the frame of stacktrace record by "github.com/pkg/errors"
// This stack is also same as "raven-go"
type StacktraceFrame struct {
	// At least one required
	Filename string `json:"filename,omitempty"`
	Function string `json:"function,omitempty"`
	Module   string `json:"module,omitempty"`

	// Optional
	Lineno       int      `json:"lineno,omitempty"`
	Colno        int      `json:"colno,omitempty"`
	AbsolutePath string   `json:"abs_path,omitempty"`
	ContextLine  string   `json:"context_line,omitempty"`
	PreContext   []string `json:"pre_context,omitempty"`
	PostContext  []string `json:"post_context,omitempty"`
	InApp        bool     `json:"in_app"`
}
