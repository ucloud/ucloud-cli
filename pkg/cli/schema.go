package cli

import (
	"encoding/json"
	"sort"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// SchemaFlag describes one flag in the schema.
type SchemaFlag struct {
	Name      string `json:"name"`
	Shorthand string `json:"shorthand,omitempty"`
	Default   string `json:"default,omitempty"`
	Usage     string `json:"usage,omitempty"`
	Required  bool   `json:"required"`
}

// SchemaCommand describes one command in the schema.
type SchemaCommand struct {
	Path  string       `json:"path"`
	Use   string       `json:"use"`
	Short string       `json:"short,omitempty"`
	Flags []SchemaFlag `json:"flags,omitempty"`
}

// RenderSchemaJSON walks the command tree (sorted, deterministic) and returns
// a JSON array of SchemaCommand. Hidden/internal commands ARE included.
func RenderSchemaJSON(root *cobra.Command) (string, error) {
	var cmds []SchemaCommand
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		sc := SchemaCommand{Path: c.CommandPath(), Use: c.Use, Short: c.Short}
		var fs []*pflag.Flag
		c.Flags().VisitAll(func(f *pflag.Flag) { fs = append(fs, f) })
		sort.Slice(fs, func(i, j int) bool { return fs[i].Name < fs[j].Name })
		for _, f := range fs {
			required := false
			if rs, ok := f.Annotations[cobra.BashCompOneRequiredFlag]; ok && len(rs) > 0 && rs[0] == "true" {
				required = true
			}
			sc.Flags = append(sc.Flags, SchemaFlag{Name: f.Name, Shorthand: f.Shorthand, Default: f.DefValue, Usage: f.Usage, Required: required})
		}
		cmds = append(cmds, sc)
		children := c.Commands()
		sort.Slice(children, func(i, j int) bool { return children[i].Use < children[j].Use })
		for _, x := range children {
			walk(x)
		}
	}
	walk(root)
	b, err := json.MarshalIndent(cmds, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
