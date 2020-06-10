//Copyright 2015 Red Hat Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package doc

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// GenDoukuTreeCustom is the the same as GenMarkdownTree, but
// with custom filePrepender and linkHandler.
func GenDoukuTreeCustom(index int, cmd *cobra.Command, dir string, filePrepender, linkHandler func(string) string) error {
	for i, c := range cmd.Commands() {
		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
			continue
		}
		if err := GenDoukuTreeCustom(i+1, c, dir, filePrepender, linkHandler); err != nil {
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

	if err := GenMarkdownCustom(cmd, f, linkHandler); err != nil {
		return err
	}
	return nil
}

// GenDoukuTree will generate a douku wiki page for this command and all
// descendants in the directory given. The header may be nil.
func GenDoukuTree(cmd *cobra.Command, dir string, linkPrefix string) error {
	doukuLink := func(s string) string {
		s = strings.TrimSuffix(s, ".md")
		return linkPrefix + strings.Replace(s, "_", "/", -1)
	}
	emptyStr := func(s string) string { return "" }
	return GenDoukuTreeCustom(0, cmd, dir, emptyStr, doukuLink)
}
