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
		Use:   "ucloud",
		Short: "The UCloud Command Line Interface v" + version,
		Long:  `The UCloud Command Line Interface is a tool to manage your UCloud services`,
		BashCompletionFunction: "__ucloud_init_completion",
	}

	cmd.PersistentFlags().StringVarP(&global.region, "region", "r", "", "Assign region(override default region of your config)")
	cmd.PersistentFlags().StringVarP(&global.projectID, "project-id", "p", "", "Assign projectId(override default projecId of your config)")
	cmd.PersistentFlags().BoolVarP(&global.debug, "debug", "d", false, "Running in debug mode")

	cmd.AddCommand(NewCmdVersion())
	cmd.AddCommand(NewCmdCompletion())
	cmd.AddCommand(NewCmdList())
	cmd.AddCommand(NewCmdConfig())
	cmd.AddCommand(NewCmdSignup())
	cmd.AddCommand(NewCmdUHost())
	cmd.AddCommand(NewCmdGssh())
	cmd.AddCommand(NewCmdEIP())

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

func init() {
	cobra.OnInitialize(initialize)
	model.ClientConfig.UserAgent = fmt.Sprintf("UCloud CLI v%s", version)
	model.ClientConfig.TracerData["command"] = fmt.Sprintf("%v", os.Args)
	model.GetContext()
}

func initialize(cmd *cobra.Command) {
	if global.debug {
		model.ClientConfig.LogLevel = 5
		model.ClientConfig.Logger = nil
	}
	client = service.NewClient(model.ClientConfig, model.Credential)

	userInfo, err := model.LoadUserInfo()
	if err == nil {
		model.ClientConfig.TracerData["userName"] = userInfo.UserEmail
		model.ClientConfig.TracerData["userID"] = userInfo.UserEmail
		model.ClientConfig.TracerData["companyName"] = userInfo.CompanyName
	} else {
		errorStr, ok := model.ClientConfig.TracerData["error"].(string)
		if ok {
			model.ClientConfig.TracerData["error"] = errorStr + "->" + err.Error()
		} else {
			model.ClientConfig.TracerData["error"] = err.Error()
		}
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
