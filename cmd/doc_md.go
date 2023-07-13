// Copyright © 2018 NAME HERE tony.li@ucloud.cn
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	"github.com/ucloud/ucloud-sdk-go/ucloud/log"

	"github.com/ucloud/ucloud-cli/base"
)

// NewCmdDoc ucloud doc
func NewCmdDoc(out io.Writer) *cobra.Command {
	var dir, format string
	cmd := &cobra.Command{
		Use:   "gendoc",
		Short: "Generate documents for all commands",
		Long:  "Generate documents for all commands. Support markdown, rst and douku",
		Run: func(c *cobra.Command, args []string) {
			base.ConfigIns.Region = ""
			base.ConfigIns.ProjectID = ""
			base.ConfigIns.Zone = ""
			rootCmd := NewCmdRoot()
			addChildren(rootCmd)
			switch format {
			case "rst":
				emptyStr := func(s string) string { return "" }
				linkHandler := func(name, ref string) string {
					return fmt.Sprintf(":ref:`%s <%s>`", name, ref)
				}
				err := doc.GenReSTTreeCustom(rootCmd, dir, emptyStr, linkHandler)
				if err != nil {
					log.Fatal(err)
				}

			case "markdown":
				err := doc.GenMarkdownTree(rootCmd, dir)
				if err != nil {
					log.Fatal(err)
				}
			case "douku":
				prefix := "cli/cmd/"
				err := doc.GenDoukuTree(rootCmd, dir, prefix)
				printCmdIndex(rootCmd, 0, "/cli/cmd")
				if err != nil {
					log.Fatal(err)
				}
			default:
				fmt.Fprintf(out, "format %s is not supported\n", format)
			}
		},
	}

	cmd.Flags().StringVar(&dir, "dir", "", "Required. The directory where documents of commands are stored")
	cmd.Flags().StringVar(&format, "format", "douku", "Required. Format of the doucments. Accept values: markdown, rst and douku")

	cmd.Flags().SetFlagValues("format", "douku", "markdown", "rst")
	cmd.Flags().SetFlagValuesFunc("dir", func() []string {
		return base.GetFileList("")
	})

	cmd.MarkFlagRequired("dir")

	return cmd
}

func printCmdIndex(curr *cobra.Command, indent int, prefix string) {
	if curr.Name() == "help" {
		return
	}
	fmt.Printf("%s* [%s](%s%s)\n", strings.Repeat("    ", indent), curr.Name(), prefix, "/"+curr.Name())
	for _, cmd := range curr.Commands() {
		printCmdIndex(cmd, indent+1, prefix+"/"+curr.Name())
	}
}
