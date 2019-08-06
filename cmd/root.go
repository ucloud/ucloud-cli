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
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	"github.com/ucloud/ucloud-sdk-go/ucloud/log"

	"github.com/ucloud/ucloud-cli/base"
)

var global = &base.Global

//NewCmdRoot 创建rootCmd rootCmd represents the base command when called without any subcommands
func NewCmdRoot() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "ucloud",
		Short:             "UCloud CLI v" + base.Version,
		Long:              `UCloud CLI - manage UCloud resources and developer workflow`,
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			if global.Version {
				base.Cxt.Printf("ucloud cli %s\n", base.Version)
			} else if global.Completion {
				NewCmdCompletion().Run(cmd, args)
			} else if global.Config {
				base.ListAggConfig(global.JSON)
			} else if global.Signup {
				NewCmdSignup().Run(cmd, args)
			} else {
				cmd.HelpFunc()(cmd, args)
			}
		},
	}

	cmd.PersistentFlags().BoolVarP(&global.Debug, "debug", "d", false, "Running in debug mode")
	cmd.PersistentFlags().BoolVarP(&global.JSON, "json", "j", false, "Print result in JSON format whenever possible")
	cmd.PersistentFlags().StringVarP(&global.Profile, "profile", "p", global.Profile, "Specifies the configuration for the operation")
	cmd.Flags().BoolVarP(&global.Version, "version", "v", false, "Display version")
	cmd.Flags().BoolVar(&global.Completion, "completion", false, "Turn on auto completion according to the prompt")
	cmd.Flags().BoolVar(&global.Config, "config", false, "Display configuration")
	cmd.Flags().BoolVar(&global.Signup, "signup", false, "Launch UCloud sign up page in browser")

	cmd.PersistentFlags().SetFlagValuesFunc("profile", func() []string { return base.AggConfigListIns.GetProfileNameList() })
	cmd.SetHelpTemplate(helpTmpl)
	cmd.SetUsageTemplate(usageTmpl)
	resetHelpFunc(cmd)

	return cmd
}

//NewCmdDoc ucloud doc
func NewCmdDoc(out io.Writer) *cobra.Command {
	var dir, format string
	cmd := &cobra.Command{
		Use:   "gendoc",
		Short: "Generate documents for all commands",
		Long:  "Generate documents for all commands. Support markdown, rst and douku",
		Run: func(c *cobra.Command, args []string) {
			rootCmd := NewCmdRoot()
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
				err := doc.GenDoukuTree(rootCmd, dir, "software/cli/cmd/")
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

const helpTmpl = `Usage:{{if .Runnable}}

  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:

  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:

  {{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Commands:{{range .Commands}}{{if .IsAvailableCommand}}

  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:

{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:

{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

//概要帮助信息模板
const usageTmpl = `Usage:{{if .Runnable}}
 {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}} [command] {{if $size:=len .Commands}}
 {{"command may be" | printf "%-20s"}} {{range $index,$cmd:= .Commands}}{{if .IsAvailableCommand}}{{$cmd.Name}}{{if gt $size  (add $index 1)}} | {{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableFlags}}
 {{"flags may be" | printf "%-20s"}} {{.Flags.FlagNames}}

Use "{{.CommandPath}} --help" for details.{{end}}
`

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cmd := NewCmdRoot()
	out := base.Cxt.GetWriter()
	base.InitConfig()
	cmd.AddCommand(NewCmdInit())
	cmd.AddCommand(NewCmdDoc(out))
	cmd.AddCommand(NewCmdConfig())
	cmd.AddCommand(NewCmdRegion(out))
	cmd.AddCommand(NewCmdProject())
	cmd.AddCommand(NewCmdUHost())
	cmd.AddCommand(NewCmdUPHost())
	cmd.AddCommand(NewCmdUImage())
	cmd.AddCommand(NewCmdSubnet())
	cmd.AddCommand(NewCmdVpc())
	cmd.AddCommand(NewCmdFirewall())
	cmd.AddCommand(NewCmdDisk())
	cmd.AddCommand(NewCmdEIP())
	cmd.AddCommand(NewCmdBandwidth())
	cmd.AddCommand(NewCmdUDPN(out))
	cmd.AddCommand(NewCmdULB())
	cmd.AddCommand(NewCmdGssh())
	cmd.AddCommand(NewCmdPathx())
	cmd.AddCommand(NewCmdMysql())
	cmd.AddCommand(NewCmdRedis())
	cmd.AddCommand(NewCmdMemcache())
	cmd.AddCommand(NewCmdExt())
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	for idx, arg := range os.Args {
		if arg == "--profile" && len(os.Args) > idx+1 && os.Args[idx+1] != "" {
			global.Profile = os.Args[idx+1]
		}
	}
	cobra.EnableCommandSorting = false
	cobra.OnInitialize(initialize)
	base.Cxt.AppendInfo("command", fmt.Sprintf("%v", os.Args))
}

func resetHelpFunc(cmd *cobra.Command) {
	for _, a := range os.Args {
		if a == "-h" {
			cmd.SetHelpTemplate(usageTmpl)
		}
	}
}

func initialize(cmd *cobra.Command) {
	flags := cmd.Flags()
	project, err := flags.GetString("project-id")
	if err == nil {
		base.ClientConfig.ProjectId = project
	}

	region, err := flags.GetString("region")
	if err == nil {
		base.ClientConfig.Region = region
	}

	zone, err := flags.GetString("zone")
	if err == nil {
		base.ClientConfig.Zone = zone
	}

	mode := os.Getenv("UCLOUD_CLI_DEBUG")
	if mode == "on" || global.Debug {
		base.ClientConfig.LogLevel = log.DebugLevel
		base.BizClient = base.NewClient(base.ClientConfig, base.AuthCredential)
	}

	if (cmd.Name() != "config" && cmd.Name() != "init" && cmd.Name() != "version") && (cmd.Parent() != nil && cmd.Parent().Name() != "config") {
		if base.ConfigIns.PrivateKey == "" {
			base.Cxt.Println("private-key is empty. Execute command 'ucloud init|config' to configure it or run 'ucloud config list' to check your configurations")
			os.Exit(0)
		}
		if base.ConfigIns.PublicKey == "" {
			base.Cxt.Println("public-key is empty. Execute command 'ucloud init|config' to configure it or run 'ucloud config list' to check your configurations")
			os.Exit(0)
		}
	}
}
