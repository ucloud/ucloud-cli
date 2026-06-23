package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestUsageTemplateRenders is a regression test for the usageTmpl helper funcs.
//
// usageTmpl (cmd/root.go) uses two template functions that the forked
// cobra/pflag provided but upstream does NOT:
//   - add: the command-list separator, `{{... (add $index 1)}}`
//   - flagNames: the flag-name list, which replaced the fork-only
//     `.Flags.FlagNames`
//
// Both are registered via cobra.AddTemplateFunc in init(). If either
// registration is removed, rendering usageTmpl breaks: an unregistered `add`
// makes the template fail to parse and panics; an unregistered `flagNames`
// leaves a template-error string in the rendered output. This test renders the
// usage template and fails on either symptom.
func TestUsageTemplateRenders(t *testing.T) {
	root := NewCmdRoot()

	// Two dummy subcommands with Run funcs so HasAvailableSubCommands is true:
	// the `add`/command-list block then renders, and the flags block calls
	// flagNames.
	root.AddCommand(&cobra.Command{
		Use: "dummyalpha",
		Run: func(c *cobra.Command, args []string) {},
	})
	root.AddCommand(&cobra.Command{
		Use: "dummybeta",
		Run: func(c *cobra.Command, args []string) {},
	})

	// Renders usageTmpl. An unregistered `add` panics here (failing the test
	// naturally); an unregistered `flagNames` leaves a template-error string in
	// the output, caught by the assertions below.
	usage := root.UsageString()

	// The command list rendered (proves the `add`/command-list block ran).
	for _, name := range []string{"dummyalpha", "dummybeta"} {
		if !strings.Contains(usage, name) {
			t.Errorf("usage output missing dummy command %q; command list did not render.\nusage:\n%s", name, usage)
		}
	}

	// No template error leaked into the output (proves flagNames/add resolved).
	for _, bad := range []string{"template:", "not defined", "can't evaluate"} {
		if strings.Contains(usage, bad) {
			t.Errorf("usage output contains template error %q (a helper func is likely unregistered).\nusage:\n%s", bad, usage)
		}
	}
}
