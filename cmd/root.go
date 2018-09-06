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
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/service"
)

//GlobalFlag 几乎所有接口都需要的参数，例如 region zone projectID
type GlobalFlag struct {
	region    string
	projectID string
	debug     bool
}

var global GlobalFlag
var client = service.NewClient(model.ClientConfig, model.Credential)

//NewCmdRoot 创建rootCmd rootCmd represents the base command when called without any subcommands
func NewCmdRoot() *cobra.Command {
	var cmd = &cobra.Command{
		Use:                    "ucloud",
		Short:                  "UCloud CLI v" + version,
		Long:                   `UCloud CLI - manage UCloud resources and developer workflow`,
		BashCompletionFunction: "__ucloud_init_completion",
	}

	cmd.PersistentFlags().StringVarP(&global.region, "region", "r", "", "Assign region(override default region of your config)")
	cmd.PersistentFlags().StringVarP(&global.projectID, "project-id", "p", "", "Assign project-id(override default projec-id of your config)")
	cmd.PersistentFlags().BoolVarP(&global.debug, "debug", "d", false, "Running in debug mode")

	cmd.AddCommand(NewCmdSignup())
	cmd.AddCommand(NewCmdConfig())
	cmd.AddCommand(NewCmdList())
	cmd.AddCommand(NewCmdUHost())
	cmd.AddCommand(NewCmdEIP())
	cmd.AddCommand(NewCmdGssh())
	cmd.AddCommand(NewCmdCompletion())
	cmd.AddCommand(NewCmdVersion())

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
	client = service.NewClient(model.ClientConfig, model.Credential)
	context = model.GetContext(os.Stdout, model.ClientConfig)
	cobra.OnInitialize(initialize)
	model.ClientConfig.UserAgent = fmt.Sprintf("UCloud CLI v%s", version)
	model.ClientConfig.TracerData["command"] = fmt.Sprintf("%v", os.Args)
}

func initialize(cmd *cobra.Command) {
	if global.debug {
		model.ClientConfig.LogLevel = 5
		model.ClientConfig.Logger = nil
	}

	userInfo, err := model.LoadUserInfo()
	if err == nil {
		model.ClientConfig.TracerData["userName"] = userInfo.UserEmail
		model.ClientConfig.TracerData["userID"] = userInfo.UserEmail
		model.ClientConfig.TracerData["companyName"] = userInfo.CompanyName
	} else {
		context.AppendError(err)
	}

	//上报服务对Origin请求头有限制，必须以'.ucloud.cn'结尾，因此这里伪造了一个sdk.ucloud.cn,跟其他上报区分
	model.ClientConfig.HTTPHeaders["Origin"] = "https://sdk.ucloud.cn"

	if (cmd.Name() != "config" && cmd.Name() != "completion" && cmd.Name() != "version") && cmd.Parent().Name() != "config" {
		if config.PrivateKey == "" {
			fmt.Println("private-key is empty. Execute command 'ucloud config' to configure your private-key")
			os.Exit(0)
		}
		if config.PublicKey == "" {
			fmt.Println("public-key is empty. Execute command 'ucloud config' to configure your public-key")
			os.Exit(0)
		}
		if config.Region == "" {
			fmt.Println("Default region is empty. Execute command 'ucloud config set region' to configure your default region")
			os.Exit(0)
		}
		if config.ProjectID == "" {
			fmt.Println("Default project-id is empty. Execute command 'ucloud config set project' to configure your default project-id")
			os.Exit(0)
		}
	}
}

func bindGlobalParam(req request.Common) {
	if global.region != "" {
		req.SetRegion(global.region)
	}
	if global.projectID != "" {
		req.SetProjectId(global.projectID)
	}
}
