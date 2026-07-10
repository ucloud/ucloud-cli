package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func subCmd(t *testing.T, root *cobra.Command, name string) *cobra.Command {
	t.Helper()
	for _, c := range root.Commands() {
		if c.Use == name {
			return c
		}
	}
	t.Fatalf("uhost subcommand %q not found", name)
	return nil
}

func topLevelCmd(t *testing.T, commands []*cobra.Command, name string) *cobra.Command {
	t.Helper()
	for _, c := range commands {
		if c.Use == name {
			return c
		}
	}
	t.Fatalf("product top-level command %q not found", name)
	return nil
}
