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
	"errors"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"

	"github.com/ucloud/ucloud-cli/base"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

const configDesc = `Public-key and private-key could be acquired from https://console.ucloud.cn/uaccount/api_manage`

const helloUcloud = `
  _   _      _ _         _   _ _____ _                 _ 
  | | | |    | | |       | | | /  __ \ |               | |
  | |_| | ___| | | ___   | | | | /  \/ | ___  _   _  __| |
  |  _  |/ _ \ | |/ _ \  | | | | |   | |/ _ \| | | |/ _\ |
  | | | |  __/ | | (_) | | |_| | \__/\ | (_) | |_| | (_| |
  \_| |_/\___|_|_|\___/   \___/ \____/_|\___/ \__,_|\__,_|

If you want add or modify your configurations, run 'ucloud config add/update'`

// NewCmdInit ucloud init
func NewCmdInit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize UCloud CLI options",
		Long:  `Initialize UCloud CLI options such as private-key,public-key,default region,zone and project.`,
		Run: func(cmd *cobra.Command, args []string) {
			fromOAuth := base.ConfigIns.AuthMode == base.AuthModeOAuth
			if fromOAuth {
				ok := base.Confirm(false, fmt.Sprintf("Profile '%s' currently uses OAuth login (auth_mode=oauth). Continue with AK/SK setup and switch this profile to key-based auth? (y/n):", base.ConfigIns.Profile))
				if !ok {
					return
				}
				clearOAuthState(base.ConfigIns)
			}

			if base.ConfigIns.PrivateKey != "" && base.ConfigIns.PublicKey != "" {
				if fromOAuth {
					if err := switchProfileToAKSK(base.ConfigIns); err != nil {
						base.HandleError(err)
						return
					}
				}
				printHello()
				return
			}

			fmt.Println(configDesc)
			base.ConfigIns.ConfigPublicKey()
			base.ConfigIns.ConfigPrivateKey()
			base.ConfigIns.ConfigBaseURL()

			region, err := fetchRegionWithConfig(base.ConfigIns)
			if err != nil {
				if uErr, ok := err.(uerr.Error); ok {
					if uErr.Code() == 172 {
						fmt.Println("public key or private key is invalid.")
						return
					}
				}
				fmt.Println(err)
				return
			}
			base.ConfigIns.Region = region.DefaultRegion
			base.ConfigIns.Zone = region.DefaultZone
			fmt.Printf("Configured default region:%s zone:%s\n", region.DefaultRegion, region.DefaultZone)

			projectID, projectName, err := getDefaultProjectWithConfig(base.ConfigIns)
			if err != nil && !errors.Is(err, errNoDefaultProject) {
				base.HandleError(err)
				return
			}
			if projectID != "" && projectName != "" {
				base.ConfigIns.ProjectID = projectID
				fmt.Printf("Configured default project:%s %s\n", projectID, projectName)
			} else {
				fmt.Println("No default project, skip.")
			}
			base.ConfigIns.Timeout = base.DefaultTimeoutSec
			base.ConfigIns.BaseURL = base.DefaultBaseURL
			base.ConfigIns.MaxRetryTimes = sdk.Int(base.DefaultMaxRetryTimes)
			base.ConfigIns.Active = true
			fmt.Printf("Configured default base url:%s\n", base.ConfigIns.BaseURL)
			fmt.Printf("Configured default timeout_sec:%ds\n", base.ConfigIns.Timeout)
			fmt.Printf("Active profile name:%s\n", base.ConfigIns.Profile)
			fmt.Println("You can change the default settings by running 'ucloud config update'")
			base.ConfigIns.ConfigUploadLog()
			err = saveInitProfile(base.ConfigIns)
			if err != nil {
				base.HandleError(fmt.Errorf("Error: %v", err))
			} else {
				base.InitConfig()
				printHello()
			}
		},
	}
	return cmd
}

// saveInitProfile 持久化 init 完整配置流程的结果；profile 已存在时（OAuth-only profile
// 切回 AK/SK 的场景）覆盖保存——依赖 ConfigIns 即 manager map 内的同一指针（InitConfig 保证）
func saveInitProfile(cfg *base.AggConfig) error {
	return base.AggConfigListIns.UpdateAggConfig(cfg)
}

// clearOAuthState 清除 profile 的 oauth 状态（口径与 'ucloud auth logout' 一致），不落盘
func clearOAuthState(cfg *base.AggConfig) {
	cfg.AuthMode = ""
	cfg.AccessToken = ""
	cfg.RefreshToken = ""
	cfg.ExpiresAt = 0
}

// switchProfileToAKSK 把 OAuth profile 切回 AK/SK：清除 oauth 状态并落盘
func switchProfileToAKSK(cfg *base.AggConfig) error {
	clearOAuthState(cfg)
	return base.AggConfigListIns.UpdateAggConfig(cfg)
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

// 根据用户设置的region和zone,检查其合法性，补上缺失的部分，给出一个合理的符合用户本意设置的region和zone
func getReasonableRegionZone(cfg *base.AggConfig) (string, string, error) {
	userRegion := cfg.Region
	userZone := cfg.Zone
	//如果zone设置了，region不能为空，因为这种情况较难判断给出一个合理的region
	if userRegion == "" && userZone != "" {
		return "", "", fmt.Errorf("region is needed if zone is assigned")
	}

	regionIns, err := fetchRegionWithConfig(cfg)
	if err != nil {
		return "", "", err
	}

	if userRegion == "" && userZone == "" {
		userRegion = regionIns.DefaultRegion
		userZone = regionIns.DefaultZone
	}

	zones, ok := regionIns.Labels[userRegion]
	if !ok {
		return "", "", fmt.Errorf("region[%s] is not exist! See 'ucloud region'", userRegion)
	}

	if userZone != "" {
		zoneExist := false
		for _, zone := range zones {
			if zone == userZone {
				zoneExist = true
			}
		}
		if !zoneExist {
			return "", "", fmt.Errorf("zone[%s] not exist in region[%s]! See 'ucloud config list' and 'ucloud region'", userZone, userRegion)
		}
	} else if len(zones) > 0 {
		userZone = zones[0]
	}

	return userRegion, userZone, nil
}

// NewCmdConfig ucloud config
func NewCmdConfig() *cobra.Command {
	var active, upload string
	cfg := base.AggConfig{}
	cmd := &cobra.Command{
		Use:     "config",
		Short:   "add or update configurations",
		Long:    `add or update configurations, such as private-key, public-key, default region and zone, base-url, timeout-sec, and default project-id`,
		Example: "ucloud config --profile=test --region cn-bj2 --active true",
		Run: func(c *cobra.Command, args []string) {
			if cfg.Profile == "" {
				c.HelpFunc()(c, args)
				return
			}

			if cfg.Timeout < 0 {
				base.HandleError(fmt.Errorf("timeout_sec must be greater than 0, accept %d", cfg.Timeout))
				return
			}

			//cacheConfig AggConfig read from $HOME/.ucloud/config.json+credential.json or empty shell
			cacheConfig, ok := base.AggConfigListIns.GetAggConfigByProfile(cfg.Profile)
			//如果配置文件中找不到该profile 则添加配置
			if !ok {
				cacheConfig = &base.AggConfig{
					PrivateKey:    cfg.PrivateKey,
					PublicKey:     cfg.PublicKey,
					Profile:       cfg.Profile,
					BaseURL:       cfg.BaseURL,
					Timeout:       cfg.Timeout,
					Active:        cfg.Active,
					Region:        cfg.Region,
					Zone:          cfg.Zone,
					ProjectID:     cfg.ProjectID,
					MaxRetryTimes: cfg.MaxRetryTimes,
				}
			}

			if cfg.PrivateKey != "" {
				cacheConfig.PrivateKey = cfg.PrivateKey
			}
			if cfg.PublicKey != "" {
				cacheConfig.PublicKey = cfg.PublicKey
			}

			if cfg.BaseURL == "" {
				if cacheConfig.BaseURL == "" {
					cacheConfig.BaseURL = base.DefaultBaseURL
				}
			} else {
				cacheConfig.BaseURL = cfg.BaseURL
			}

			if cfg.Timeout == 0 {
				if cacheConfig.Timeout == 0 {
					cacheConfig.Timeout = base.DefaultTimeoutSec
				}
			} else {
				cacheConfig.Timeout = cfg.Timeout
			}

			if *cfg.MaxRetryTimes == 0 {
				if *cacheConfig.MaxRetryTimes == 0 {
					cacheConfig.MaxRetryTimes = sdk.Int(base.DefaultMaxRetryTimes)
				}
			} else {
				cacheConfig.MaxRetryTimes = cfg.MaxRetryTimes
			}

			if cfg.Region != "" {
				cacheConfig.Region = cfg.Region
			}
			if cfg.Zone != "" {
				cacheConfig.Zone = cfg.Zone
			}

			//确保设置的Region和Zone真实存在
			region, zone, err := getReasonableRegionZone(cacheConfig)
			if err != nil {
				base.HandleError(fmt.Errorf("verify region failed: %v", err))
			} else {
				cacheConfig.Region = region
				cacheConfig.Zone = zone
			}

			//如果用户填写的project和配置文件中该配置的project均为空，则调接口拉取默认project
			//如果用户填写的project不为空，则校验其是否真实存在;
			if cfg.ProjectID == "" {
				if cacheConfig.ProjectID == "" {
					id, _, err := getDefaultProjectWithConfig(cacheConfig)
					if err != nil {
						base.HandleError(fmt.Errorf("fetch default project failed: %v", err))
					} else {
						cacheConfig.ProjectID = id
					}
				}
			} else {
				cfg.ProjectID = base.PickResourceID(cfg.ProjectID)
				projects, err := fetchProjectWithConfig(cacheConfig)
				if err != nil {
					cacheConfig.ProjectID = cfg.ProjectID
				} else {
					if ok := projects[cfg.ProjectID]; ok {
						cacheConfig.ProjectID = cfg.ProjectID
					} else {
						base.HandleError(fmt.Errorf("project %s you assigned not exists", cfg.ProjectID))
						if ok := projects[cacheConfig.ProjectID]; !ok {
							base.HandleError(fmt.Errorf("project %s not exists, assign another one please", cacheConfig.ProjectID))
						}
					}
				}
			}

			if active != "" {
				if active == "true" {
					cacheConfig.Active = true
				} else if active == "false" {
					cacheConfig.Active = false
				} else {
					base.HandleError(fmt.Errorf("flag active should be true or false. received %s", active))
				}
			}

			if upload != "" {
				if upload == "true" {
					cacheConfig.AgreeUploadLog = true
				} else if upload == "false" {
					cacheConfig.AgreeUploadLog = false
				} else {
					base.HandleError(fmt.Errorf("flag agree-upload-log should be true or false. received %s", active))
				}
			}

			err = base.AggConfigListIns.UpdateAggConfig(cacheConfig)
			if err != nil {
				base.HandleError(err)
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
	flags.StringVar(&cfg.BaseURL, "base-url", "", "Optional. Set default base url. For instance 'https://api.ucloud.cn/'")
	flags.IntVar(&cfg.Timeout, "timeout-sec", 0, "Optional. Set default timeout for requesting API. Unit: seconds")
	cfg.MaxRetryTimes = flags.Int("max-retry-times", 0, "Optional. Set default max-retry-times for idempotent APIs which can be called many times without side effect, for example 'ReleaseEIP'")
	flags.StringVar(&active, "active", "", "Optional. Mark the profile to be effective or not. Accept valeus: true or false")
	flags.StringVar(&upload, "agree-upload-log", "false", "Optional. Agree to upload log in local file ~/.ucloud/cli.log or not. Accept valeus: true or false")

	command.SetFlagValues(cmd, "active", "true", "false")
	command.SetFlagValues(cmd, "agree-upload-log", "true", "false")
	command.SetCompletion(cmd, "profile", func() []string { return base.AggConfigListIns.GetProfileNameList() })
	command.SetCompletion(cmd, "region", getRegionList)
	command.SetCompletion(cmd, "project-id", getProjectList)
	command.SetCompletion(cmd, "zone", func() []string {
		return getZoneList(cfg.Region)
	})

	cmd.AddCommand(NewCmdConfigAdd())
	cmd.AddCommand(NewCmdConfigUpdate())
	cmd.AddCommand(NewCmdConfigList())
	cmd.AddCommand(NewCmdConfigDelete())
	return cmd
}

// NewCmdConfigAdd ucloud config add
func NewCmdConfigAdd() *cobra.Command {
	var active, upload string
	cfg := &base.AggConfig{}
	cmd := &cobra.Command{
		Use:   "add",
		Short: "add configuration",
		Long:  "add configuration",
		Run: func(c *cobra.Command, args []string) {
			region, zone, err := getReasonableRegionZone(cfg)
			if err != nil {
				base.HandleError(err)
			}
			cfg.Region = region
			cfg.Zone = zone

			project, err := getReasonableProject(cfg)
			if err != nil {
				base.HandleError(err)
			}
			cfg.ProjectID = project

			if cfg.Timeout <= 0 {
				base.HandleError(fmt.Errorf("timeout_sec must be greater than 0, accept %d", cfg.Timeout))
				return
			}

			if active == "true" {
				cfg.Active = true
			} else if active == "false" {
				cfg.Active = false
			} else {
				fmt.Printf("active should be true or false, received %s\n", active)
			}

			if upload == "true" {
				cfg.AgreeUploadLog = true
			} else if upload == "false" {
				cfg.AgreeUploadLog = false
			} else {
				fmt.Printf("agree-upload-log should be true or false, received %s\n", active)
			}

			err = base.AggConfigListIns.Append(cfg)
			if err != nil {
				base.HandleError(err)
				return
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringVar(&cfg.Profile, "profile", "", "Required. Set name of CLI profile")
	flags.StringVar(&cfg.PublicKey, "public-key", "", "Required. Set public key")
	flags.StringVar(&cfg.PrivateKey, "private-key", "", "Required. Set private key")
	flags.StringVar(&cfg.Region, "region", "", "Optional. Set default region. For instance 'cn-bj2' See 'ucloud region'")
	flags.StringVar(&cfg.Zone, "zone", "", "Optional. Set default zone. For instance 'cn-bj2-02'. See 'ucloud region'")
	flags.StringVar(&cfg.ProjectID, "project-id", "", "Optional. Set default project. For instance 'org-xxxxxx'. See 'ucloud project list")
	flags.StringVar(&cfg.BaseURL, "base-url", base.DefaultBaseURL, "Optional. Set default base url. For instance 'https://api.ucloud.cn/'")
	flags.IntVar(&cfg.Timeout, "timeout-sec", base.DefaultTimeoutSec, "Optional. Set default timeout for requesting API. Unit: seconds")
	cfg.MaxRetryTimes = flags.Int("max-retry-times", base.DefaultMaxRetryTimes, "Optional. Set default max-retry-times for idempotent APIs which can be called many times without side effect, for example 'ReleaseEIP'")
	flags.StringVar(&active, "active", "false", "Optional. Mark the profile to be effective or not. Accept valeus: true or false")
	flags.StringVar(&upload, "agree-upload-log", "false", "Optional. Agree to upload log in local file ~/.ucloud/cli.log or not. Accept valeus: true or false")

	command.SetFlagValues(cmd, "active", "true", "false")
	command.SetFlagValues(cmd, "agree-upload-log", "true", "false")
	command.SetCompletion(cmd, "profile", func() []string { return base.AggConfigListIns.GetProfileNameList() })
	command.SetCompletion(cmd, "region", getRegionList)
	command.SetCompletion(cmd, "project-id", getProjectList)
	command.SetCompletion(cmd, "zone", func() []string {
		return getZoneList(cfg.Region)
	})

	cmd.MarkFlagRequired("profile")
	cmd.MarkFlagRequired("public-key")
	cmd.MarkFlagRequired("private-key")

	return cmd
}

// NewCmdConfigUpdate ucloud config update
func NewCmdConfigUpdate() *cobra.Command {
	var timeout, active, maxRetries, upload string
	cfg := &base.AggConfig{}
	cmd := &cobra.Command{
		Use:   "update",
		Short: "update configurations",
		Long:  "update configurations",
		Run: func(c *cobra.Command, args []string) {
			//cacheConfig AggConfig read from $HOME/.ucloud/config.json+credential.json or empty shell
			cacheConfig, ok := base.AggConfigListIns.GetAggConfigByProfile(cfg.Profile)
			if !ok {
				base.HandleError(fmt.Errorf("profile %s not exist", cfg.Profile))
				return
			}

			if cfg.PrivateKey != "" {
				cacheConfig.PrivateKey = cfg.PrivateKey
			}
			if cfg.PublicKey != "" {
				cacheConfig.PublicKey = cfg.PublicKey
			}

			//如果配置了公私钥，则先更新让其生效, 为接下来拉取Region,Zone做准备
			if cfg.PrivateKey != "" || cfg.PublicKey != "" {
				base.AggConfigListIns.UpdateAggConfig(cacheConfig)
			}

			//先应用连接类参数(base-url/timeout-sec/max-retry-times)，确保接下来的远程校验
			//打到新网关而不是旧的(可能已不可用的)网关，避免旧base-url坏掉后无法改回的死锁
			if cfg.BaseURL != "" {
				cacheConfig.BaseURL = cfg.BaseURL
			}

			if timeout != "" {
				seconds, err := strconv.Atoi(timeout)
				if err != nil {
					base.HandleError(fmt.Errorf("parse timeout-sec failed: %v", err))
					return
				}
				cacheConfig.Timeout = seconds
			}

			if cacheConfig.Timeout <= 0 {
				base.HandleError(fmt.Errorf("timeout-sec must be greater than 0, accept %d", cfg.Timeout))
				return
			}

			if maxRetries != "" {
				times, err := strconv.Atoi(maxRetries)
				if err != nil {
					base.HandleError(fmt.Errorf("parse max-retry-times failed: %v", err))
					return
				}
				cacheConfig.MaxRetryTimes = &times
			}

			if *cacheConfig.MaxRetryTimes < 0 {
				base.HandleError(fmt.Errorf("max-retry-timesc must be greater than or equal to 0, accept %d", cfg.MaxRetryTimes))
				return
			}

			//如有设置Region和Zone，确保设置的Region和Zone真实存在
			if cfg.Region != "" {
				cacheConfig.Region = cfg.Region
			}
			if cfg.Zone != "" {
				cacheConfig.Zone = cfg.Zone
			}

			region, zone, err := getReasonableRegionZone(cacheConfig)
			if err != nil {
				base.HandleError(err)
				return
			}

			cacheConfig.Region = region
			cacheConfig.Zone = zone

			if cfg.ProjectID != "" {
				cacheConfig.ProjectID = base.PickResourceID(cfg.ProjectID)
			}

			project, err := getReasonableProject(cacheConfig)
			if err != nil {
				base.HandleError(err)
			}
			cacheConfig.ProjectID = project

			if active == "true" {
				cacheConfig.Active = true
			} else if active == "false" {
				cacheConfig.Active = false
			}

			if upload == "true" {
				cacheConfig.AgreeUploadLog = true
			} else if upload == "false" {
				cacheConfig.AgreeUploadLog = false
			}

			err = base.AggConfigListIns.UpdateAggConfig(cacheConfig)
			if err != nil {
				base.HandleError(err)
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringVar(&cfg.Profile, "profile", "", "Required. Set name of CLI profile")
	flags.StringVar(&cfg.PublicKey, "public-key", "", "Required. Set public key")
	flags.StringVar(&cfg.PrivateKey, "private-key", "", "Required. Set private key")
	flags.StringVar(&cfg.Region, "region", "", "Optional. Set default region. For instance 'cn-bj2' See 'ucloud region'")
	flags.StringVar(&cfg.Zone, "zone", "", "Optional. Set default zone. For instance 'cn-bj2-02'. See 'ucloud region'")
	flags.StringVar(&cfg.ProjectID, "project-id", "", "Optional. Set default project. For instance 'org-xxxxxx'. See 'ucloud project list")
	flags.StringVar(&cfg.BaseURL, "base-url", "", "Optional. Set default base url. For instance 'https://api.ucloud.cn/'")
	flags.StringVar(&timeout, "timeout-sec", "", "Optional. Set default timeout for requesting API. Unit: seconds")
	flags.StringVar(&maxRetries, "max-retry-times", "", "Optional. Set default max retry times for idempotent APIs which can be called many times without side effect, for example 'ReleaseEIP'")
	flags.StringVar(&active, "active", "", "Optional. Mark the profile to be effective")
	flags.StringVar(&upload, "agree-upload-log", "", "Optional. Agree to upload log in local file ~/.ucloud/cli.log or not. Accept valeus: true or false")

	command.SetCompletion(cmd, "profile", func() []string { return base.AggConfigListIns.GetProfileNameList() })
	command.SetCompletion(cmd, "region", getRegionList)
	command.SetCompletion(cmd, "project-id", getProjectList)
	command.SetCompletion(cmd, "zone", func() []string {
		return getZoneList(cfg.Region)
	})
	command.SetFlagValues(cmd, "active", "true", "false")
	command.SetFlagValues(cmd, "agree-upload-log", "true", "false")

	cmd.MarkFlagRequired("profile")

	return cmd
}

// NewCmdConfigList ucloud config list
func NewCmdConfigList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list all configurations",
		Long:  `list all configurations`,
		Run: func(c *cobra.Command, args []string) {
			base.ListAggConfig(global.JSON)
		},
	}
	return cmd
}

// NewCmdConfigDelete ucloud config Delete
func NewCmdConfigDelete() *cobra.Command {
	var profileList []string
	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "delete configurations by profile name",
		Long:    "delete configurations by profile name",
		Example: "ucloud config delete --profile test",
		Run: func(c *cobra.Command, args []string) {
			profiles := base.AggConfigListIns.GetProfileNameList()
			allProfileMap := make(map[string]bool)
			for _, p := range profiles {
				allProfileMap[p] = true
			}

			for _, p := range profileList {
				if allProfileMap[p] {
					err := base.AggConfigListIns.DeleteByProfile(p)
					if err != nil {
						base.HandleError(err)
					}
				} else {
					base.HandleError(fmt.Errorf("profile %s does not exist", p))
				}
			}
		},
	}
	cmd.Flags().StringSliceVar(&profileList, "profile", nil, "Required. Name of settings item")
	cmd.MarkFlagRequired("profile")
	command.SetCompletion(cmd, "profile", func() []string { return base.AggConfigListIns.GetProfileNameList() })
	return cmd
}
