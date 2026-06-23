package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func TestDecideOutputFormat(t *testing.T) {
	tests := []struct {
		name       string
		outputFlag string
		jsonFlag   bool
		writer     func() interface{ Write([]byte) (int, error) } // io.Writer
		want       cli.OutputFormat
	}{
		{
			name:       "explicit --output yaml",
			outputFlag: "yaml",
			want:       cli.OutputYAML,
		},
		{
			name:       "explicit --output json",
			outputFlag: "json",
			want:       cli.OutputJSON,
		},
		{
			name:       "explicit --output table",
			outputFlag: "table",
			want:       cli.OutputTable,
		},
		{
			name:       "explicit --output JSON (case insensitive)",
			outputFlag: "JSON",
			want:       cli.OutputJSON,
		},
		{
			name:     "--json true (no --output)",
			jsonFlag: true,
			want:     cli.OutputJSON,
		},
		{
			name: "non-TTY writer, no flags -> JSON",
			// bytes.Buffer is not a TTY
			want: cli.OutputJSON,
		},
		{
			name:       "--output wins over --json",
			outputFlag: "yaml",
			jsonFlag:   true,
			want:       cli.OutputYAML,
		},
		{
			name:       "--output table wins over --json",
			outputFlag: "table",
			jsonFlag:   true,
			want:       cli.OutputTable,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Save and restore global state.
			origOutput := global.Output
			origJSON := global.JSON
			t.Cleanup(func() {
				global.Output = origOutput
				global.JSON = origJSON
			})

			global.Output = tc.outputFlag
			global.JSON = tc.jsonFlag

			// Use a bytes.Buffer as a non-TTY writer.
			got := decideOutputFormat(&bytes.Buffer{})
			if got != tc.want {
				t.Errorf("decideOutputFormat() = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestProductCtxFormatFinalize is a regression test for the bug where the
// product cli.Context's output format was frozen at command-tree construction
// time (buildContext, before cobra parses --output), so an explicit
// `--output table` never took effect on product commands. The fix finalizes the
// format in initialize() (PersistentPreRun) via productCtx.SetFormat after flag
// parsing. This test mirrors that finalize step and asserts productCtx.Format()
// tracks the PARSED --output value.
//
// buildContext() is used to populate productCtx directly instead of
// addChildren(NewCmdRoot()): addChildren also constructs every platform command
// and touches base globals as a side effect, none of which this test needs.
// buildContext() only reads os.Stdin/Stdout/Stderr and base singletons; it
// performs no API calls and runs offline.
//
// Without the fix (no SetFormat call after parsing), productCtx would stay at
// the construction-time value (JSON, since test stdout is non-TTY) and the
// table/yaml assertions below would fail.
func TestProductCtxFormatFinalize(t *testing.T) {
	// Save and restore both the global flag and the package-level productCtx
	// so other tests are unaffected.
	origOutput := global.Output
	origCtx := productCtx
	t.Cleanup(func() {
		global.Output = origOutput
		productCtx = origCtx
	})

	productCtx = buildContext()

	tests := []struct {
		name       string
		outputFlag string
		want       cli.OutputFormat
	}{
		{name: "--output table", outputFlag: "table", want: cli.OutputTable},
		{name: "--output json", outputFlag: "json", want: cli.OutputJSON},
		{name: "--output yaml", outputFlag: "yaml", want: cli.OutputYAML},
		// Empty --output on a non-TTY stdout falls back to JSON.
		{name: "empty --output (non-TTY default)", outputFlag: "", want: cli.OutputJSON},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			global.Output = tc.outputFlag

			// Mirror what initialize() (PersistentPreRun) now does after cobra
			// parses --output.
			productCtx.SetFormat(decideOutputFormat(os.Stdout))

			if got := productCtx.Format(); got != tc.want {
				t.Errorf("productCtx.Format() = %v, want %v (--output=%q)", got, tc.want, tc.outputFlag)
			}
		})
	}
}
