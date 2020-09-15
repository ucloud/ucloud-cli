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
	"strconv"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/base"
	"github.com/ucloud/ucloud-sdk-go/ucloud/log"
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

func addChildren(root *cobra.Command) {
	out := base.Cxt.GetWriter()
	root.AddCommand(NewCmdInit())
	root.AddCommand(NewCmdDoc(out))
	root.AddCommand(NewCmdConfig())
	root.AddCommand(NewCmdRegion(out))
	root.AddCommand(NewCmdProject())
	root.AddCommand(NewCmdUHost())
	root.AddCommand(NewCmdUPHost())
	root.AddCommand(NewCmdUImage())
	root.AddCommand(NewCmdSubnet())
	root.AddCommand(NewCmdVpc())
	root.AddCommand(NewCmdFirewall())
	root.AddCommand(NewCmdDisk())
	root.AddCommand(NewCmdEIP())
	root.AddCommand(NewCmdBandwidth())
	root.AddCommand(NewCmdUDPN(out))
	root.AddCommand(NewCmdULB())
	root.AddCommand(NewCmdGssh())
	root.AddCommand(NewCmdPathx())
	root.AddCommand(NewCmdMysql())
	root.AddCommand(NewCmdRedis())
	root.AddCommand(NewCmdMemcache())
	root.AddCommand(NewCmdExt())
	root.AddCommand(NewCmdAPI(out))
	for _, c := range root.Commands() {
		if c.Name() != "init" && c.Name() != "gendoc" && c.Name() != "config" {
			c.PersistentFlags().StringVar(&global.PublicKey, "public-key", global.PublicKey, "Set public-key to override the public-key in local config file")
			c.PersistentFlags().StringVar(&global.PrivateKey, "private-key", global.PrivateKey, "Set private-key to override the private-key in local config file")
			c.PersistentFlags().StringVar(&global.BaseURL, "base-url", "", "Set base-url to override the base-url in local config file")
			c.PersistentFlags().IntVar(&global.Timeout, "timeout-sec", 0, "Set timeout-sec to override the timeout-sec in local config file")
			c.PersistentFlags().IntVar(&global.MaxRetryTimes, "max-retry-times", -1, "Set max-retry-times to override the max-retry-times in local config file")
		}
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cmd := NewCmdRoot()
	if base.InCloudShell {
		err := base.InitConfigInCloudShell()
		if err != nil {
			base.HandleError(err)
			return
		}
	}
	base.InitConfig()
	mode := os.Getenv("UCLOUD_CLI_DEBUG")
	if mode == "on" || global.Debug {
		base.ClientConfig.LogLevel = log.DebugLevel
		base.BizClient = base.NewClient(base.ClientConfig, base.AuthCredential)
	}

	addChildren(cmd)

	targetCmd, flags, err := cmd.Find(os.Args[1:])
	if err == nil {
		if targetCmd.Use == "api" {
			targetCmd.Run(targetCmd, flags)
			return
		}
	}

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	//-1表示不覆盖配置文件中的MaxRetryTimes参数
	global.MaxRetryTimes = -1
	for idx, arg := range os.Args {
		if arg == "--profile" && len(os.Args) > idx+1 && os.Args[idx+1] != "" {
			global.Profile = os.Args[idx+1]
		}
		if arg == "--public-key" && len(os.Args) > idx+1 && os.Args[idx+1] != "" {
			global.PublicKey = os.Args[idx+1]
		}
		if arg == "--private-key" && len(os.Args) > idx+1 && os.Args[idx+1] != "" {
			global.PrivateKey = os.Args[idx+1]
		}
		if arg == "--base-url" && len(os.Args) > idx+1 && os.Args[idx+1] != "" {
			global.BaseURL = os.Args[idx+1]
		}
		if arg == "--timeout-sec" && len(os.Args) > idx+1 && os.Args[idx+1] != "" {
			sec, err := strconv.Atoi(os.Args[idx+1])
			if err != nil {
				fmt.Printf("parse timeout-sec failed: %v\n", err)
			} else {
				global.Timeout = sec
			}
		}
		if arg == "--max-retry-times" && len(os.Args) > idx+1 && os.Args[idx+1] != "" {
			times, err := strconv.Atoi(os.Args[idx+1])
			if err != nil {
				fmt.Printf("parse max-retry-times failed: %v\n", err)
			} else {
				global.MaxRetryTimes = times
			}
		}
	}
	cobra.EnableCommandSorting = false
	cobra.OnInitialize(initialize)
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

	if (cmd.Name() != "config" && cmd.Name() != "init" && cmd.Name() != "version") && (cmd.Parent() != nil && cmd.Parent().Name() != "config") {
		if base.InCloudShell {
			return
		}
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
