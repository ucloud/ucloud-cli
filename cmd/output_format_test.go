package cmd

import (
	"bytes"
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
