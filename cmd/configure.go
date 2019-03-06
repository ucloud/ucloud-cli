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
	"reflect"

	"github.com/spf13/cobra"

	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"

	"github.com/ucloud/ucloud-cli/base"
)

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
			if base.ConfigIns.PrivateKey != "" && base.ConfigIns.PublicKey != "" {
				printHello()
				return
			}

			base.Cxt.Println(configDesc)
			base.ConfigIns.ConfigPublicKey()
			base.ConfigIns.ConfigPrivateKey()

			region, zone, err := getDefaultRegion()
			if err != nil {
				if uErr, ok := err.(uerr.Error); ok {
					if uErr.Code() == 172 {
						fmt.Println("public key or private key is invalid.")
						return
					}
				}
				base.Cxt.Println(err)
				return
			}
			base.ConfigIns.Region = region
			base.ConfigIns.Zone = zone
			base.Cxt.Printf("Configured default region:%s zone:%s\n", region, zone)

			projectID, projectName, err := getDefaultProject()
			if err != nil {
				base.Cxt.Println(err)
				return
			}
			base.ConfigIns.ProjectID = projectID
			base.Cxt.Printf("Configured default project:%s %s\n", projectID, projectName)
			base.ConfigIns.Timeout = base.DefaultTimeoutSec
			base.ConfigIns.BaseURL = base.DefaultBaseURL
			base.ConfigIns.Active = true
			base.Cxt.Printf("Configured default base url:%s\n", base.ConfigIns.BaseURL)
			base.Cxt.Printf("Configured default timeout_sec:%ds\n", base.ConfigIns.Timeout)
			base.Cxt.Printf("default name of CLI profile:%s\n", base.ConfigIns.Profile)
			base.Cxt.Println("You can change the default settings by running 'ucloud config'")
			base.ConfigIns.Save()
			printHello()
		},
	}
	return cmd
}

func printHello() {
	userInfo, err := getUserInfo()
	if err != nil {
		base.Cxt.PrintErr(err)
		return
	}
	base.Cxt.Printf("You are logged in as: [%s]\n", userInfo.UserEmail)
	certified := isUserCertified(userInfo)
	if !certified {
		base.Cxt.Println("\nWarning: Please authenticate the account with your valid documentation at 'https://accountv2.ucloud.cn/authentication'.")
	}
	base.Cxt.Println(helloUcloud)
}

//NewCmdConfig ucloud config
func NewCmdConfig() *cobra.Command {
	cfg := base.AggConfig{}
	cmd := &cobra.Command{
		Use:     "config",
		Short:   "Configure UCloud CLI options",
		Long:    `Configure UCloud CLI options such as private-key,public-key,default region and default project-id.`,
		Example: "ucloud config --profile=test --region=cn-bj2 --active",
		Run: func(c *cobra.Command, args []string) {
			//cacheConfig AggConfig read from $HOME/.ucloud/config.json+credential.json or empty shell
			cacheConfig, err := base.GetAggConfigByProfile(cfg.Profile)
			if err != nil {
				base.HandleError(err)
				return
			}
			//如有设置Region和Zone，确保设置的Region和Zone真实存在
			if cfg.Region != "" || cfg.Zone != "" {
				regionMap, err := fetchRegion()
				if err != nil {
					base.HandleError(err)
					return
				}

				region := cfg.Region
				if region == "" {
					region = cacheConfig.Region
				}

				zones, ok := regionMap[region]
				if !ok {
					base.Cxt.Printf("Error, region[%s] is not exist! See 'ucloud region'\n", region)
					return
				}

				zone := cfg.Zone
				if zone == "" {
					zone = cacheConfig.Zone
				}

				if zone != "" {
					zoneExist := false
					for _, zone := range zones {
						if zone == cfg.Zone {
							zoneExist = true
						}
					}
					if !zoneExist {
						base.Cxt.Printf("Error, zone[%s] not exist in region[%s]! See 'ucloud config list' and 'ucloud region'\n", zone, region)
						return
					}
				}
			}
			if cfg.Timeout <= 0 {
				base.HandleError(fmt.Errorf("timeout_sec must be greater than 0, accept %d", cfg.Timeout))
				return
			}

			changed := false
			cfg.ProjectID = base.PickResourceID(cfg.ProjectID)
			tmpCfgVal := reflect.ValueOf(cfg)
			configVal := reflect.ValueOf(cacheConfig).Elem()
			for i := 0; i < tmpCfgVal.NumField(); i++ {
				if fieldVal := tmpCfgVal.Field(i); fieldVal.Interface() != reflect.Zero(fieldVal.Type()).Interface() {
					configVal.Field(i).Set(fieldVal)
					changed = true
				}
			}
			if changed {
				err := cacheConfig.Save()
				if err != nil {
					base.HandleError(err)
				}
			} else {
				c.HelpFunc()(c, args)
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringVar(&cfg.Profile, "profile", "", "Required. Set name of CLI profile")
	flags.StringVar(&cfg.PublicKey, "public-key", "", "Optional. Set public key")
	flags.StringVar(&cfg.PrivateKey, "private-key", "", "Optional. Set private key")
	flags.StringVar(&cfg.Region, "region", "", "Optional. Set default region. For instance 'cn-bj2' See 'ucloud region'")
	flags.StringVar(&cfg.Zone, "zone", "", "Optional. Set default zone. For instance 'cn-bj2-02'. See 'ucloud region'")
	flags.StringVar(&cfg.ProjectID, "project-id", "", "Optional. Set default project. For instance 'org-xxxxxx'. See 'ucloud project list")
	flags.StringVar(&cfg.BaseURL, "base-url", base.DefaultBaseURL, "Optional. Set default base url. For instance 'https://api.ucloud.cn/'")
	flags.IntVar(&cfg.Timeout, "timeout-sec", base.DefaultTimeoutSec, "Optional. Set default timeout for requesting API. Unit: seconds")
	flags.BoolVar(&cfg.Active, "active", false, "Optional. Mark the profile to be effective")

	flags.SetFlagValuesFunc("profile", base.GetProfileNameList)
	flags.SetFlagValuesFunc("region", getRegionList)
	flags.SetFlagValuesFunc("project-id", getProjectList)
	flags.SetFlagValuesFunc("zone", func() []string {
		return getZoneList(cfg.Region)
	})

	cmd.MarkFlagRequired("profile")

	cmd.AddCommand(NewCmdConfigList())
	cmd.AddCommand(NewCmdConfigDelete())
	return cmd
}

//NewCmdConfigList ucloud config list
func NewCmdConfigList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list all settings",
		Long:  `list all settings`,
		Run: func(c *cobra.Command, args []string) {
			base.ListAggConfig(global.JSON)
		},
	}
	return cmd
}

//NewCmdConfigDelete ucloud config Delete
func NewCmdConfigDelete() *cobra.Command {
	var profile string
	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "delete settings by profile name",
		Long:    "delete settings by profile name",
		Example: "ucloud config delete --profile test",
		Run: func(c *cobra.Command, args []string) {
			profiles := base.GetProfileNameList()
			if profiles != nil {
				exist := false
				for _, p := range profiles {
					if p == profile {
						exist = true
						break
					}
				}
				if !exist {
					base.HandleError(fmt.Errorf("profile:%s is not exists", profile))
					return
				}
			}
			aggc, err := base.GetAggConfigByProfile(profile)
			if err != nil {
				base.HandleError(err)
			}
			if aggc.Active {
				base.HandleError(fmt.Errorf("the active config can not be deleted,please switch it to another one by 'ucloud config --profile xxx --active' and try again"))
				return
			}
			err = base.DeleteAggConfigByProfile(profile)
			if err != nil {
				base.HandleError(err)
			}
		},
	}
	cmd.Flags().StringVar(&profile, "profile", "", "Required. Name of settings item")
	cmd.MarkFlagRequired("profile")
	cmd.Flags().SetFlagValuesFunc("profile", base.GetProfileNameList)
	return cmd
}
