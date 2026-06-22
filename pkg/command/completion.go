package command

import "github.com/spf13/cobra"

// SetCompletion registers a dynamic completion candidate provider for a flag.
// INTERNALS NOTE: while on the lixiaojun629 cobra fork this delegates to the
// fork's *pflag.FlagSet.SetFlagValuesFunc. Task C2 (drop fork → upstream cobra)
// swaps the body to cobra's RegisterFlagCompletionFunc; the signature is stable
// so the 305 call sites migrated in C1 are unaffected.
func SetCompletion(cmd *cobra.Command, name string, fn func() []string) {
	cmd.Flags().SetFlagValuesFunc(name, fn)
}

// SetFlagValues registers a static completion candidate set for a flag.
// Same fork-internal-now, upstream-at-C2 strategy as SetCompletion.
func SetFlagValues(cmd *cobra.Command, name string, values ...string) {
	_ = cmd.Flags().SetFlagValues(name, values...)
}
