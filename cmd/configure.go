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

	"github.com/ucloud/ucloud-cli/cmd/internal/platform"
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
			fromOAuth := platform.ConfigIns.AuthMode == platform.AuthModeOAuth
			if fromOAuth {
				ok := platform.Confirm(false, fmt.Sprintf("Profile '%s' currently uses OAuth login (auth_mode=oauth). Continue with AK/SK setup and switch this profile to key-based auth? (y/n):", platform.ConfigIns.Profile))
				if !ok {
					return
				}
				clearOAuthState(platform.ConfigIns)
			}

			if platform.ConfigIns.PrivateKey != "" && platform.ConfigIns.PublicKey != "" {
				if fromOAuth {
					if err := switchProfileToAKSK(platform.ConfigIns); err != nil {
						platform.HandleError(err)
						return
					}
				}
				printHello()
				return
			}

			fmt.Println(configDesc)
			platform.ConfigIns.ConfigPublicKey()
			platform.ConfigIns.ConfigPrivateKey()
			platform.ConfigIns.ConfigBaseURL()

			region, err := fetchRegionWithConfig(platform.ConfigIns)
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
			platform.ConfigIns.Region = region.DefaultRegion
			platform.ConfigIns.Zone = region.DefaultZone
			fmt.Printf("Configured default region:%s zone:%s\n", region.DefaultRegion, region.DefaultZone)

			projectID, projectName, err := getDefaultProjectWithConfig(platform.ConfigIns)
			if err != nil && !errors.Is(err, errNoDefaultProject) {
				platform.HandleError(err)
				return
			}
			if projectID != "" && projectName != "" {
				platform.ConfigIns.ProjectID = projectID
				fmt.Printf("Configured default project:%s %s\n", projectID, projectName)
			} else {
				fmt.Println("No default project, skip.")
			}
			platform.ConfigIns.Timeout = platform.DefaultTimeoutSec
			platform.ConfigIns.BaseURL = platform.DefaultBaseURL
			platform.ConfigIns.MaxRetryTimes = sdk.Int(platform.DefaultMaxRetryTimes)
			platform.ConfigIns.Active = true
			fmt.Printf("Configured default base url:%s\n", platform.ConfigIns.BaseURL)
			fmt.Printf("Configured default timeout_sec:%ds\n", platform.ConfigIns.Timeout)
			fmt.Printf("Active profile name:%s\n", platform.ConfigIns.Profile)
			fmt.Println("You can change the default settings by running 'ucloud config update'")
			platform.ConfigIns.ConfigUploadLog()
			err = saveInitProfile(platform.ConfigIns)
			if err != nil {
				platform.HandleError(fmt.Errorf("Error: %v", err))
			} else {
				platform.InitConfig()
				printHello()
			}
		},
	}
	return cmd
}

// saveInitProfile 持久化 init 完整配置流程的结果；profile 已存在时（OAuth-only profile
// 切回 AK/SK 的场景）覆盖保存——依赖 ConfigIns 即 manager map 内的同一指针（InitConfig 保证）
func saveInitProfile(cfg *platform.AggConfig) error {
	return platform.AggConfigListIns.UpdateAggConfig(cfg)
}

// clearOAuthState 清除 profile 的 oauth 状态（口径与 'ucloud auth logout' 一致），不落盘
func clearOAuthState(cfg *platform.AggConfig) {
	cfg.AuthMode = ""
	cfg.AccessToken = ""
	cfg.RefreshToken = ""
	cfg.ExpiresAt = 0
}

// switchProfileToAKSK 把 OAuth profile 切回 AK/SK：清除 oauth 状态并落盘
func switchProfileToAKSK(cfg *platform.AggConfig) error {
	clearOAuthState(cfg)
	return platform.AggConfigListIns.UpdateAggConfig(cfg)
}

func printHello() {
	userInfo, err := getUserInfo()
	if err != nil {
		platform.Cxt.PrintErr(err)
		return
	}
	platform.Cxt.Printf("You are logged in as: [%s]\n", userInfo.UserEmail)
	certified := isUserCertified(userInfo)
	if !certified {
		platform.Cxt.Println("\nWarning: Please authenticate the account with your valid documentation at 'https://accountv2.ucloud.cn/authentication'.")
	}
	platform.Cxt.Println(helloUcloud)
}

// 根据用户设置的region和zone,检查其合法性，补上缺失的部分，给出一个合理的符合用户本意设置的region和zone
func getReasonableRegionZone(cfg *platform.AggConfig) (string, string, error) {
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
	cfg := platform.AggConfig{}
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
				platform.HandleError(fmt.Errorf("timeout_sec must be greater than 0, accept %d", cfg.Timeout))
				return
			}

			//cacheConfig AggConfig read from $HOME/.ucloud/config.json+credential.json or empty shell
			cacheConfig, ok := platform.AggConfigListIns.GetAggConfigByProfile(cfg.Profile)
			//如果配置文件中找不到该profile 则添加配置
			if !ok {
				cacheConfig = &platform.AggConfig{
					PrivateKey:    cfg.PrivateKey,
					PublicKey:     cfg.PublicKey,
					Profile:       cfg.Profile,
					BaseURL:       cfg.BaseURL,
					ChannelKey:    cfg.ChannelKey,
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
					cacheConfig.BaseURL = platform.DefaultBaseURL
				}
			} else {
				cacheConfig.BaseURL = cfg.BaseURL
			}

			//channel-key 属连接类参数，与 base-url 同批应用：必须早于下方 region/project
			//远程校验，否则校验请求不带 key，专属云 profile 在配置时即报错而配不上。
			//用 Changed() 而非空值判断：空是合法值（专属云切回主站需清除它），
			//--channel-key "" 应能清空，这与 base-url「空=不改」的既有局限不同。
			if c.Flags().Changed("channel-key") {
				cacheConfig.ChannelKey = cfg.ChannelKey
			}

			if cfg.Timeout == 0 {
				if cacheConfig.Timeout == 0 {
					cacheConfig.Timeout = platform.DefaultTimeoutSec
				}
			} else {
				cacheConfig.Timeout = cfg.Timeout
			}

			if *cfg.MaxRetryTimes == 0 {
				if *cacheConfig.MaxRetryTimes == 0 {
					cacheConfig.MaxRetryTimes = sdk.Int(platform.DefaultMaxRetryTimes)
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

			//确保设置的Region和Zone真实存在。校验失败即整体放弃、不落盘：
			//此前用 else 保留原值后仍照常写盘，与 add/update 的 fail-closed 口径不一致
			region, zone, err := getReasonableRegionZone(cacheConfig)
			if err != nil {
				platform.HandleError(fmt.Errorf("verify region failed: %v", err))
				return
			}
			cacheConfig.Region = region
			cacheConfig.Zone = zone

			//如果用户填写的project和配置文件中该配置的project均为空，则调接口拉取默认project
			//如果用户填写的project不为空，则校验其是否真实存在;
			if cfg.ProjectID == "" {
				if cacheConfig.ProjectID == "" {
					//此处直接调用、未经 %v 包装，sentinel 链完整：errNoDefaultProject
					//属良性缺失，放行并留空 ProjectID，口径与 ucloud init 一致
					id, _, err := getDefaultProjectWithConfig(cacheConfig)
					if err != nil && !errors.Is(err, errNoDefaultProject) {
						platform.HandleError(fmt.Errorf("fetch default project failed: %v", err))
						return
					}
					if err == nil {
						cacheConfig.ProjectID = id
					}
				}
			} else {
				cfg.ProjectID = platform.PickResourceID(cfg.ProjectID)
				projects, err := fetchProjectWithConfig(cacheConfig)
				if err != nil {
					//远程不可达时此前直接采信用户输入并落盘，等于写入未经校验的 project；
					//现与其余路径一致：拒绝
					platform.HandleError(fmt.Errorf("fetch project failed: %v", err))
					return
				}
				if ok := projects[cfg.ProjectID]; !ok {
					platform.HandleError(fmt.Errorf("project %s you assigned not exists", cfg.ProjectID))
					if ok := projects[cacheConfig.ProjectID]; !ok {
						platform.HandleError(fmt.Errorf("project %s not exists, assign another one please", cacheConfig.ProjectID))
					}
					return
				}
				cacheConfig.ProjectID = cfg.ProjectID
			}

			if active != "" {
				if active == "true" {
					cacheConfig.Active = true
				} else if active == "false" {
					cacheConfig.Active = false
				} else {
					platform.HandleError(fmt.Errorf("flag active should be true or false. received %s", active))
				}
			}

			if upload != "" {
				if upload == "true" {
					cacheConfig.AgreeUploadLog = true
				} else if upload == "false" {
					cacheConfig.AgreeUploadLog = false
				} else {
					platform.HandleError(fmt.Errorf("flag agree-upload-log should be true or false. received %s", active))
				}
			}

			err = platform.AggConfigListIns.UpdateAggConfig(cacheConfig)
			if err != nil {
				platform.HandleError(err)
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
	flags.StringVar(&cfg.ChannelKey, "channel-key", "", "Optional. Set channel-key for a dedicated cloud channel that reuses the main-site domain. For instance 'ch_xxx'. Leave empty for the main site or a channel with its own domain")
	flags.IntVar(&cfg.Timeout, "timeout-sec", 0, "Optional. Set default timeout for requesting API. Unit: seconds")
	cfg.MaxRetryTimes = flags.Int("max-retry-times", 0, "Optional. Set default max-retry-times for idempotent APIs which can be called many times without side effect, for example 'ReleaseEIP'")
	flags.StringVar(&active, "active", "", "Optional. Mark the profile to be effective or not. Accept valeus: true or false")
	flags.StringVar(&upload, "agree-upload-log", "false", "Optional. Agree to upload log in local file ~/.ucloud/cli.log or not. Accept valeus: true or false")

	command.SetFlagValues(cmd, "active", "true", "false")
	command.SetFlagValues(cmd, "agree-upload-log", "true", "false")
	command.SetCompletion(cmd, "profile", func() []string { return platform.AggConfigListIns.GetProfileNameList() })
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
	cfg := &platform.AggConfig{}
	cmd := &cobra.Command{
		Use:   "add",
		Short: "add configuration",
		Long:  "add configuration",
		Run: func(c *cobra.Command, args []string) {
			//远程校验失败即整体放弃，不得落盘：否则 profile 会带着被抹空的 region/zone
			//建成，且 --active true 时还会把原有 active profile 顶掉
			region, zone, err := getReasonableRegionZone(cfg)
			if err != nil {
				platform.HandleError(err)
				return
			}
			cfg.Region = region
			cfg.Zone = zone

			//归一化 project-id：补全吐出的是 org-xxx/Name 形式（见 getProjectList），
			//与主命令/update 一致剥成裸 id。否则合法的补全值会被 getReasonableProject
			//判为「project does not exist」——叠加 fail-closed 会硬拦一个真实存在的 project
			if cfg.ProjectID != "" {
				cfg.ProjectID = platform.PickResourceID(cfg.ProjectID)
			}

			//errNoDefaultProject 是良性缺失（账号有项目但未设默认），放行并留空 ProjectID，
			//口径与 ucloud init 一致；其余错误一律拒绝落盘
			project, err := getReasonableProject(cfg)
			if err != nil && !errors.Is(err, errNoDefaultProject) {
				platform.HandleError(err)
				return
			}
			cfg.ProjectID = project

			if cfg.Timeout <= 0 {
				platform.HandleError(fmt.Errorf("timeout_sec must be greater than 0, accept %d", cfg.Timeout))
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

			err = platform.AggConfigListIns.Append(cfg)
			if err != nil {
				platform.HandleError(err)
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
	flags.StringVar(&cfg.BaseURL, "base-url", platform.DefaultBaseURL, "Optional. Set default base url. For instance 'https://api.ucloud.cn/'")
	flags.StringVar(&cfg.ChannelKey, "channel-key", "", "Optional. Set channel-key for a dedicated cloud channel that reuses the main-site domain. For instance 'ch_xxx'. Leave empty for the main site or a channel with its own domain")
	flags.IntVar(&cfg.Timeout, "timeout-sec", platform.DefaultTimeoutSec, "Optional. Set default timeout for requesting API. Unit: seconds")
	cfg.MaxRetryTimes = flags.Int("max-retry-times", platform.DefaultMaxRetryTimes, "Optional. Set default max-retry-times for idempotent APIs which can be called many times without side effect, for example 'ReleaseEIP'")
	flags.StringVar(&active, "active", "false", "Optional. Mark the profile to be effective or not. Accept valeus: true or false")
	flags.StringVar(&upload, "agree-upload-log", "false", "Optional. Agree to upload log in local file ~/.ucloud/cli.log or not. Accept valeus: true or false")

	command.SetFlagValues(cmd, "active", "true", "false")
	command.SetFlagValues(cmd, "agree-upload-log", "true", "false")
	command.SetCompletion(cmd, "profile", func() []string { return platform.AggConfigListIns.GetProfileNameList() })
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
	cfg := &platform.AggConfig{}
	cmd := &cobra.Command{
		Use:   "update",
		Short: "update configurations",
		Long:  "update configurations",
		Run: func(c *cobra.Command, args []string) {
			//cacheConfig AggConfig read from $HOME/.ucloud/config.json+credential.json or empty shell
			cacheConfig, ok := platform.AggConfigListIns.GetAggConfigByProfile(cfg.Profile)
			if !ok {
				platform.HandleError(fmt.Errorf("profile %s not exist", cfg.Profile))
				return
			}

			//GetAggConfigByProfile 返回的是 manager map 内条目本身。改动一律先落在副本上，
			//校验全部通过后才交给 UpdateAggConfig 写回，任一步失败都不污染内存态、更不落盘。
			//这也堵死了 OAuth token 刷新 handler 持 manager 写回时把半成品一并 Save 的路径。
			//注意 MaxRetryTimes 是 *int，浅拷贝与原对象共享该指针：下方只做整体替换，
			//不得改成 *draft.MaxRetryTimes = x 的原地写。
			draft := *cacheConfig

			//AK/SK 只写内存即可供下方远程校验使用：BuildClientRuntime 完全从传入的
			//AggConfig 构造 sdk.Config 与 credential，不读磁盘，故无需先行落盘。
			if cfg.PrivateKey != "" {
				draft.PrivateKey = cfg.PrivateKey
			}
			if cfg.PublicKey != "" {
				draft.PublicKey = cfg.PublicKey
			}

			//先应用连接类参数(base-url/channel-key/timeout-sec/max-retry-times)，确保接下来的远程校验
			//打到新网关而不是旧的(可能已不可用的)网关，避免旧base-url坏掉后无法改回的死锁
			if cfg.BaseURL != "" {
				draft.BaseURL = cfg.BaseURL
			}

			//channel-key 同属连接类参数：专属云 profile 的远程校验请求必须带上它，
			//否则网关报 174 而校验失败，profile 永远配不上。
			//Changed() 使 --channel-key "" 可清除（专属云切回主站的场景）。
			if c.Flags().Changed("channel-key") {
				draft.ChannelKey = cfg.ChannelKey
			}

			if timeout != "" {
				seconds, err := strconv.Atoi(timeout)
				if err != nil {
					platform.HandleError(fmt.Errorf("parse timeout-sec failed: %v", err))
					return
				}
				draft.Timeout = seconds
			}

			//报告被检查的那个值：cfg.Timeout 是本命令从不赋值的字段（flag 绑定的是
			//局部变量 timeout），报它恒为 0，与实际被拒的值无关
			if draft.Timeout <= 0 {
				platform.HandleError(fmt.Errorf("timeout-sec must be greater than 0, accept %d", draft.Timeout))
				return
			}

			if maxRetries != "" {
				times, err := strconv.Atoi(maxRetries)
				if err != nil {
					platform.HandleError(fmt.Errorf("parse max-retry-times failed: %v", err))
					return
				}
				draft.MaxRetryTimes = &times
			}

			//同上，且更隐蔽：cfg.MaxRetryTimes 是本命令从不赋值的 *int，
			//%d 对指针打印的是地址（nil 即 0），并非用户传入的值。必须解引用。
			if *draft.MaxRetryTimes < 0 {
				platform.HandleError(fmt.Errorf("max-retry-times must be greater than or equal to 0, accept %d", *draft.MaxRetryTimes))
				return
			}

			//如有设置Region和Zone，确保设置的Region和Zone真实存在
			if cfg.Region != "" {
				draft.Region = cfg.Region
			}
			if cfg.Zone != "" {
				draft.Zone = cfg.Zone
			}

			region, zone, err := getReasonableRegionZone(&draft)
			if err != nil {
				platform.HandleError(err)
				return
			}

			draft.Region = region
			draft.Zone = zone

			if cfg.ProjectID != "" {
				draft.ProjectID = platform.PickResourceID(cfg.ProjectID)
			}

			//errNoDefaultProject 是良性缺失（账号有项目但未设默认），放行并留空 ProjectID，
			//口径与 ucloud init 一致；其余错误一律拒绝落盘——此处曾漏 return 而抹空 ProjectID
			project, err := getReasonableProject(&draft)
			if err != nil && !errors.Is(err, errNoDefaultProject) {
				platform.HandleError(err)
				return
			}
			draft.ProjectID = project

			if active == "true" {
				draft.Active = true
			} else if active == "false" {
				draft.Active = false
			}

			if upload == "true" {
				draft.AgreeUploadLog = true
			} else if upload == "false" {
				draft.AgreeUploadLog = false
			}

			err = platform.AggConfigListIns.UpdateAggConfig(&draft)
			if err != nil {
				platform.HandleError(err)
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringVar(&cfg.Profile, "profile", "", "Required. Set name of CLI profile")
	//公私钥在 update 中为可选：仅在非空时更新（见 Run）。只有 profile 是 MarkFlagRequired
	flags.StringVar(&cfg.PublicKey, "public-key", "", "Optional. Set public key")
	flags.StringVar(&cfg.PrivateKey, "private-key", "", "Optional. Set private key")
	flags.StringVar(&cfg.Region, "region", "", "Optional. Set default region. For instance 'cn-bj2' See 'ucloud region'")
	flags.StringVar(&cfg.Zone, "zone", "", "Optional. Set default zone. For instance 'cn-bj2-02'. See 'ucloud region'")
	flags.StringVar(&cfg.ProjectID, "project-id", "", "Optional. Set default project. For instance 'org-xxxxxx'. See 'ucloud project list")
	flags.StringVar(&cfg.BaseURL, "base-url", "", "Optional. Set default base url. For instance 'https://api.ucloud.cn/'")
	flags.StringVar(&cfg.ChannelKey, "channel-key", "", "Optional. Set channel-key for a dedicated cloud channel that reuses the main-site domain. For instance 'ch_xxx'. Pass an empty value to clear it")
	flags.StringVar(&timeout, "timeout-sec", "", "Optional. Set default timeout for requesting API. Unit: seconds")
	flags.StringVar(&maxRetries, "max-retry-times", "", "Optional. Set default max retry times for idempotent APIs which can be called many times without side effect, for example 'ReleaseEIP'")
	flags.StringVar(&active, "active", "", "Optional. Mark the profile to be effective")
	flags.StringVar(&upload, "agree-upload-log", "", "Optional. Agree to upload log in local file ~/.ucloud/cli.log or not. Accept valeus: true or false")

	command.SetCompletion(cmd, "profile", func() []string { return platform.AggConfigListIns.GetProfileNameList() })
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
			platform.ListAggConfig(global.JSON)
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
			profiles := platform.AggConfigListIns.GetProfileNameList()
			allProfileMap := make(map[string]bool)
			for _, p := range profiles {
				allProfileMap[p] = true
			}

			for _, p := range profileList {
				if allProfileMap[p] {
					err := platform.AggConfigListIns.DeleteByProfile(p)
					if err != nil {
						platform.HandleError(err)
					}
				} else {
					platform.HandleError(fmt.Errorf("profile %s does not exist", p))
				}
			}
		},
	}
	cmd.Flags().StringSliceVar(&profileList, "profile", nil, "Required. Name of settings item")
	cmd.MarkFlagRequired("profile")
	command.SetCompletion(cmd, "profile", func() []string { return platform.AggConfigListIns.GetProfileNameList() })
	return cmd
}
