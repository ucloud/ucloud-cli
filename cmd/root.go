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

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/model"
	. "github.com/ucloud/ucloud-cli/util"
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

const version = "0.1.2"

var global GlobalFlag

//NewCmdRoot 创建rootCmd rootCmd represents the base command when called without any subcommands
func NewCmdRoot() *cobra.Command {
	var cmd = &cobra.Command{
		Use:                    "ucloud",
		Short:                  "UCloud CLI v" + version,
		Long:                   `UCloud CLI - manage UCloud resources and developer workflow`,
		BashCompletionFunction: "__ucloud_init_completion",
		Run: func(cmd *cobra.Command, args []string) {
			if global.version {
				Tracer.Printf("ucloud cli %s\n", version)
			} else if global.completion {
				NewCmdCompletion().Run(cmd, args)
			} else if global.config {
				config.ListConfig()
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
	cmd.Flags().BoolVar(&global.completion, "completion", false, "Create or update completion scripts")
	cmd.Flags().BoolVar(&global.config, "config", false, "Display configuration")
	cmd.Flags().BoolVar(&global.signup, "signup", false, "Launch UCloud sign up page in browser")

	cmd.AddCommand(NewCmdConfig())
	cmd.AddCommand(NewCmdRegion())
	cmd.AddCommand(NewCmdUHost())
	cmd.AddCommand(NewCmdEIP())
	cmd.AddCommand(NewCmdGssh())

	return cmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	command := NewCmdRoot()
	if err := command.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var context *model.Context

func init() {
	cobra.EnableCommandSorting = false
	// bizClient = service.NewClient(model.ClientConfig, model.Credential)
	context = model.GetContext(os.Stdout, SdkClient)
	cobra.OnInitialize(initialize)
	Tracer.AppendInfo("command", fmt.Sprintf("%v", os.Args))
}

func initialize(cmd *cobra.Command) {
	if global.debug {
		ClientConfig.Logger.SetLevel(5)
	}

	userInfo, err := LoadUserInfo()
	if err == nil {
		Tracer.AppendInfo("userName", userInfo.UserEmail)
		Tracer.AppendInfo("companyName", userInfo.CompanyName)
	} else {
		Tracer.AppendError(err)
	}

	if (cmd.Name() != "config" && cmd.Name() != "completion" && cmd.Name() != "version") && (cmd.Parent() != nil && cmd.Parent().Name() != "config") {
		if config.PrivateKey == "" {
			Tracer.Println("private-key is empty. Execute command 'ucloud config' to configure your private-key")
			os.Exit(0)
		}
		if config.PublicKey == "" {
			Tracer.Println("public-key is empty. Execute command 'ucloud config' to configure your public-key")
			os.Exit(0)
		}
	}
}
