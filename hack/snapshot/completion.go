package snapshot

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// completionResult holds the outcome of classifying a flag's completion func.
type completionResult struct {
	registered bool     // false → no completion func registered; skip.
	isDynamic  bool     // true → completion requires network (BizClient); record as "dynamic".
	candidates []string // non-nil when isDynamic=false and registered=true.
}

// classifyFlag invokes the completion func for the named flag and classifies it
// as static (fixed candidate set) or dynamic (requires network).
//
// Dynamic detection: SetCompletion closures touch the network-backing globals
// which the test nils after tree construction, causing a nil-pointer panic.
// Platform (cmd) closures dereference base.BizClient; product (products/udb)
// closures go through cli.NewServiceClient, which builds an SDK client from
// base.ClientConfig — so the test nils both (see TestWriteCompletionBaseline).
// We recover from the panic and mark the flag dynamic. A closure may also
// signal dynamic explicitly by returning cobra.ShellCompDirectiveError.
// SetFlagValues closures return a fixed slice and never touch those globals, so
// they succeed without panicking and are recorded as static.
func classifyFlag(c *cobra.Command, flagName string) completionResult {
	fn, ok := c.GetFlagCompletionFunc(flagName)
	if !ok {
		return completionResult{}
	}

	var isDynamic bool
	var candidates []string

	func() {
		defer func() {
			if r := recover(); r != nil {
				isDynamic = true
			}
		}()
		results, directive := fn(c, []string{}, "")
		if directive == cobra.ShellCompDirectiveError {
			isDynamic = true
			return
		}
		candidates = results
	}()

	return completionResult{registered: true, isDynamic: isDynamic, candidates: candidates}
}

// RenderCompletion returns a deterministic text dump of completion registrations
// for the entire cobra command tree rooted at root.
//
// Format (one line per flag that has a registered completion func):
//
//	<CommandPath>\t<flagName>\tstatic\t<comma-joined sorted candidates>
//	<CommandPath>\t<flagName>\tdynamic
//
// Flags with no registered completion are omitted.
// Subcommands are visited in sorted order; flags are visited in sorted order.
func RenderCompletion(root *cobra.Command) string {
	var b strings.Builder
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		// Collect all flags on this command (non-persistent only; persistent flags
		// are registered on the defining command and appear there too).
		var fs []*pflag.Flag
		c.Flags().VisitAll(func(f *pflag.Flag) { fs = append(fs, f) })
		sort.Slice(fs, func(i, j int) bool { return fs[i].Name < fs[j].Name })

		for _, f := range fs {
			r := classifyFlag(c, f.Name)
			if !r.registered {
				continue // no completion func → skip.
			}
			if r.isDynamic {
				// Registered but requires network/BizClient.
				fmt.Fprintf(&b, "%s\t%s\tdynamic\n", c.CommandPath(), f.Name)
			} else {
				// Static enum — sort candidates for determinism.
				sorted := append([]string(nil), r.candidates...)
				sort.Strings(sorted)
				fmt.Fprintf(&b, "%s\t%s\tstatic\t%s\n", c.CommandPath(), f.Name, strings.Join(sorted, ","))
			}
		}

		ch := c.Commands()
		sort.Slice(ch, func(i, j int) bool { return ch[i].Use < ch[j].Use })
		for _, x := range ch {
			walk(x)
		}
	}
	walk(root)
	return b.String()
}
