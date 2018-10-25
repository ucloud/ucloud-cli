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
	"reflect"

	"github.com/ucloud/ucloud-cli/ux"

	"github.com/spf13/cobra"

	. "github.com/ucloud/ucloud-cli/base"
)

var config = ConfigInstance

const configDesc = `Public-key and private-key could be acquired from https://console.ucloud.cn/uapi/apikey.`

const helloUcloud = `
  _   _      _ _         _   _ _____ _                 _ 
  | | | |    | | |       | | | /  __ \ |               | |
  | |_| | ___| | | ___   | | | | /  \/ | ___  _   _  __| |
  |  _  |/ _ \ | |/ _ \  | | | | |   | |/ _ \| | | |/ _\ |
  | | | |  __/ | | (_) | | |_| | \__/\ | (_) | |_| | (_| |
  \_| |_/\___|_|_|\___/   \___/ \____/_|\___/ \__,_|\__,_|
  `

//NewCmdInit ucloud init
func NewCmdInit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize UCloud CLI options",
		Long:  `Initialize UCloud CLI options such as private-key,public-key,default region,zone and project.`,
		Run: func(cmd *cobra.Command, args []string) {
			Cxt.Println(configDesc)
			if len(config.PrivateKey) != 0 && len(config.PublicKey) != 0 {
				confirm, err := ux.Prompt("Your have already configured public-key and private-key. Do you want to overwrite it? (y/n):")
				if err != nil {
					Cxt.Println(err)
					return
				}
				if confirm {
					printHello()
					return
				}
			}
			config.ClearConfig()
			ClientConfig.Region = ""
			ClientConfig.ProjectId = ""
			config.ConfigPublicKey()
			config.ConfigPrivateKey()

			region, zone, err := getDefaultRegion()
			if err != nil {
				Cxt.Println(err)
				return
			}
			config.Region = region
			config.Zone = zone
			Cxt.Printf("Configured default region:%s zone:%s\n", region, zone)

			projectId, projectName, err := getDefaultProject()
			if err != nil {
				Cxt.Println(err)
				return
			}
			config.ProjectID = projectId
			Cxt.Printf("Configured default project:%s %s\n", projectId, projectName)
			config.SaveConfig()
			printHello()
		},
	}
	return cmd
}

func printHello() {
	userInfo, err := getUserInfo()
	Cxt.Printf("You are logged in as: [%s]\n", userInfo.UserEmail)
	certified := isUserCertified(userInfo)
	if err != nil {
		Cxt.PrintErr(err)
	} else if certified == false {
		Cxt.Println("\nWarning: Please authenticate the account with your valid documentation at 'https://accountv2.ucloud.cn/authentication'.")
	}
	Cxt.Println(helloUcloud)
}

//NewCmdConfig ucloud config
func NewCmdConfig() *cobra.Command {
	cfg := Config{}
	cmd := &cobra.Command{
		Use:     "config",
		Short:   "Configure UCloud CLI options",
		Long:    `Configure UCloud CLI options such as private-key,public-key,default region and default project-id.`,
		Example: "ucloud config list; ucloud config --region cn-bj2",
		Run: func(cmd *cobra.Command, args []string) {
			if cfg.Region != "" || cfg.Zone != "" {
				regionMap, err := fetchRegion()
				if err != nil {
					HandleError(err)
					return
				}

				region := cfg.Region
				if region == "" {
					region = config.Region
				}

				zones, ok := regionMap[region]
				if !ok {
					Cxt.Printf("Error, region[%s] not exist! See 'ucloud region'\n", region)
					return
				}

				zone := cfg.Zone
				if zone == "" {
					zone = config.Zone
				}

				if zone != "" {
					zoneExist := false
					for _, zone := range zones {
						if zone == cfg.Zone {
							zoneExist = true
						}
					}
					if !zoneExist {
						Cxt.Printf("Error, zone[%s] not exist in region[%s]! See 'ucloud config list' and 'ucloud region'\n", zone, region)
						return
					}
				}
			}

			tmpCfgVal := reflect.ValueOf(cfg)
			configVal := reflect.ValueOf(config).Elem()
			changed := false
			for i := 0; i < tmpCfgVal.NumField(); i++ {
				if fieldVal := tmpCfgVal.Field(i).String(); fieldVal != "" {
					configVal.Field(i).SetString(fieldVal)
					changed = true
				}
			}
			if changed {
				config.SaveConfig()
			} else {
				cmd.HelpFunc()(cmd, args)
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringVar(&cfg.PublicKey, "public-key", "", "Optional. Set public key")
	flags.StringVar(&cfg.PrivateKey, "private-key", "", "Optional. Set private key")
	flags.StringVar(&cfg.Region, "region", "", "Optional. Set default region. For instance 'cn-bj2' See 'ucloud region'")
	flags.StringVar(&cfg.Zone, "zone", "", "Optional. Set default zone. For instance 'cn-bj2-02'. See 'ucloud region'")
	flags.StringVar(&cfg.ProjectID, "project-id", "", "Optional. Set default project. For instance 'org-xxxxxx'. See 'ucloud project list")

	cmd.AddCommand(NewCmdConfigList())
	cmd.AddCommand(NewCmdConfigClear())

	return cmd
}

//NewCmdConfigList ucloud config list
func NewCmdConfigList() *cobra.Command {
	configListCmd := &cobra.Command{
		Use:   "list",
		Short: "list all settings",
		Long:  `list all settings`,
		Run: func(cmd *cobra.Command, args []string) {
			config.ListConfig(global.json)
		},
	}
	return configListCmd
}

//NewCmdConfigClear ucloud config clear
func NewCmdConfigClear() *cobra.Command {
	configClearCmd := &cobra.Command{
		Use:   "clear",
		Short: "clear all settings",
		Long:  "clear all settings",
		Run: func(cmd *cobra.Command, args []string) {
			config.ClearConfig()
		},
	}
	return configClearCmd
}
