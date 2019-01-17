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
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/base"
)

//GlobalFlag 几乎所有接口都需要的参数，例如 region zone projectID
type GlobalFlag struct {
	debug      bool
	json       bool
	version    bool
	completion bool
	config     bool
	signup     bool
}

var global GlobalFlag

//NewCmdRoot 创建rootCmd rootCmd represents the base command when called without any subcommands
func NewCmdRoot() *cobra.Command {
	out := base.Cxt.GetWriter()
	cmd := &cobra.Command{
		Use:                    "ucloud",
		Short:                  "UCloud CLI v" + base.Version,
		Long:                   `UCloud CLI - manage UCloud resources and developer workflow`,
		BashCompletionFunction: "__ucloud_init_completion",
		Run: func(cmd *cobra.Command, args []string) {
			if global.version {
				base.Cxt.Printf("ucloud cli %s\n", base.Version)
			} else if global.completion {
				NewCmdCompletion().Run(cmd, args)
			} else if global.config {
				base.ListAggConfig(global.json)
			} else if global.signup {
				NewCmdSignup().Run(cmd, args)
			} else {
				cmd.HelpFunc()(cmd, args)
			}
		},
	}

	cmd.PersistentFlags().BoolVarP(&global.debug, "debug", "d", false, "Running in debug mode")
	cmd.PersistentFlags().BoolVarP(&global.json, "json", "j", false, "Print result in JSON format whenever possible")
	cmd.Flags().BoolVar(&global.version, "version", false, "Display version")
	cmd.Flags().BoolVar(&global.completion, "completion", false, "Turn on auto completion according to the prompt")
	cmd.Flags().BoolVar(&global.config, "config", false, "Display configuration")
	cmd.Flags().BoolVar(&global.signup, "signup", false, "Launch UCloud sign up page in browser")

	cmd.AddCommand(NewCmdInit())
	cmd.AddCommand(NewCmdConfig())
	cmd.AddCommand(NewCmdRegion())
	cmd.AddCommand(NewCmdProject())
	cmd.AddCommand(NewCmdUHost())
	cmd.AddCommand(NewCmdEIP())
	cmd.AddCommand(NewCmdGssh())
	cmd.AddCommand(NewCmdUImage())
	cmd.AddCommand(NewCmdSubnet())
	cmd.AddCommand(NewCmdVpc())
	cmd.AddCommand(NewCmdFirewall())
	cmd.AddCommand(NewCmdDisk())
	cmd.AddCommand(NewCmdBandwidthPkg())
	cmd.AddCommand(NewCmdSharedBW())
	cmd.AddCommand(NewCmdUDPN(out))

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
	rootCmd := NewCmdRoot()
	rootCmd.SetHelpTemplate(helpTmpl)
	rootCmd.SetUsageTemplate(usageTmpl)
	resetHelpFunc(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
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

	if global.debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	userInfo, err := base.LoadUserInfo()
	if err == nil {
		base.Cxt.AppendInfo("userName", userInfo.UserEmail)
		base.Cxt.AppendInfo("companyName", userInfo.CompanyName)
	} else {
		base.Cxt.PrintErr(err)
	}

	if (cmd.Name() != "config" && cmd.Name() != "init" && cmd.Name() != "version") && (cmd.Parent() != nil && cmd.Parent().Name() != "config") {
		if base.ConfigIns.PrivateKey == "" {
			base.Cxt.Println("private-key is empty. Execute command 'ucloud init' or 'ucloud config' to configure your private-key")
			os.Exit(0)
		}
		if base.ConfigIns.PublicKey == "" {
			base.Cxt.Println("public-key is empty. Execute command 'ucloud init' or 'ucloud config' to configure your public-key")
			os.Exit(0)
		}
	}
}
