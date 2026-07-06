package snapshot

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Render returns a deterministic text dump of the entire cobra command tree.
// It captures each command's path, use, short, and each flag's name/shorthand/default/required.
// Completion candidate values are intentionally NOT captured — they are verified separately.
func Render(root *cobra.Command) string { return RenderPlatform(root, nil) }

// RenderPlatform is Render minus the top-level subtrees named in skip.
// Product-claimed top-level commands are guarded by their own goldens under
// products/<name>/testdata/, so the platform golden must not duplicate them;
// only direct children of root are ever pruned.
func RenderPlatform(root *cobra.Command, skip map[string]bool) string {
	var b strings.Builder
	var walk func(c *cobra.Command, depth int)
	walk = func(c *cobra.Command, depth int) {
		fmt.Fprintf(&b, "%s\tuse=%s\tshort=%s\n", c.CommandPath(), c.Use, c.Short)
		var fs []*pflag.Flag
		c.Flags().VisitAll(func(f *pflag.Flag) { fs = append(fs, f) })
		sort.Slice(fs, func(i, j int) bool { return fs[i].Name < fs[j].Name })
		for _, f := range fs {
			req := ""
			if rs, ok := f.Annotations[cobra.BashCompOneRequiredFlag]; ok && len(rs) > 0 && rs[0] == "true" {
				req = "true"
			}
			fmt.Fprintf(&b, "  flag=%s\tshort=%s\tdefault=%s\trequired=%s\n", f.Name, f.Shorthand, f.DefValue, req)
		}
		ch := c.Commands()
		sort.Slice(ch, func(i, j int) bool { return ch[i].Use < ch[j].Use })
		for _, x := range ch {
			if depth == 0 && skip[x.Name()] {
				continue
			}
			walk(x, depth+1)
		}
	}
	walk(root, 0)
	return b.String()
}
