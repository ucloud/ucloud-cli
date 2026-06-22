package cmd

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// genDoukuTreeCustom is the same as GenMarkdownTree, but with custom
// filePrepender and linkHandler. Relocated in-tree from the cobra fork's
// doc.GenDoukuTreeCustom so the fork can be dropped (Task C2). Uses only
// stable upstream cobra/doc API (doc.GenMarkdownCustom).
func genDoukuTreeCustom(index int, cmd *cobra.Command, dir string, filePrepender, linkHandler func(string) string) error {
	for i, c := range cmd.Commands() {
		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
			continue
		}
		if err := genDoukuTreeCustom(i+1, c, dir, filePrepender, linkHandler); err != nil {
			return err
		}
	}

	basename := strings.Replace(cmd.CommandPath(), " ", "/", -1) + ".md"
	filename := filepath.Join(dir, basename)

	fp, _ := filepath.Split(filename)
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		os.MkdirAll(fp, 0755)
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.WriteString(f, filePrepender(filename)); err != nil {
		return err
	}

	if err := doc.GenMarkdownCustom(cmd, f, linkHandler); err != nil {
		return err
	}
	return nil
}

// genDoukuTree generates a douku wiki page for this command and all
// descendants in the directory given. Relocated in-tree from the cobra fork.
func genDoukuTree(cmd *cobra.Command, dir string, linkPrefix string) error {
	doukuLink := func(s string) string {
		s = strings.TrimSuffix(s, ".md")
		return linkPrefix + strings.Replace(s, "_", "/", -1)
	}
	emptyStr := func(s string) string { return "" }
	return genDoukuTreeCustom(0, cmd, dir, emptyStr, doukuLink)
}
