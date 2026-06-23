package command

import (
	"github.com/spf13/cobra"
)

// SetCompletion registers a dynamic completion candidate provider for a flag,
// via upstream cobra's RegisterFlagCompletionFunc.
func SetCompletion(cmd *cobra.Command, name string, fn func() []string) {
	_ = cmd.RegisterFlagCompletionFunc(name, func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		return fn(), cobra.ShellCompDirectiveNoFileComp
	})
}

// SetPersistentCompletion registers a dynamic completion provider for a
// persistent flag. Upstream RegisterFlagCompletionFunc resolves persistent
// flags itself, so this is identical to SetCompletion; kept as a distinct name
// for call-site clarity (the profile flag in cmd/root.go is persistent).
func SetPersistentCompletion(cmd *cobra.Command, name string, fn func() []string) {
	_ = cmd.RegisterFlagCompletionFunc(name, func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		return fn(), cobra.ShellCompDirectiveNoFileComp
	})
}

// SetFlagValues registers a static completion candidate set for a flag, via
// upstream cobra's RegisterFlagCompletionFunc.
func SetFlagValues(cmd *cobra.Command, name string, values ...string) {
	vals := append([]string(nil), values...)
	_ = cmd.RegisterFlagCompletionFunc(name, func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		return vals, cobra.ShellCompDirectiveNoFileComp
	})
}
