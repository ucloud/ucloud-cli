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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/uaccount"

	"github.com/ucloud/ucloud-cli/cmd/internal/platform"
)

// NewCmdRegion ucloud region
func NewCmdRegion(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "region",
		Short:   "List all region and zone",
		Long:    "List all region and zone",
		Example: "ucloud region",
		Run: func(cmd *cobra.Command, args []string) {
			regionIns, err := fetchRegion()
			if err != nil {
				platform.HandleError(err)
				return
			}
			regionList := make([]RegionTable, 0)
			for region, zones := range regionIns.Labels {
				regionList = append(regionList, RegionTable{region, strings.Join(zones, ", ")})
			}
			platform.PrintList(regionList, out)
		},
	}
	return cmd
}

// RegionTable 为显示region表格创建的类型
type RegionTable struct {
	Region string
	Zones  string
}

func getDefaultRegion() (string, string, error) {
	client := newServiceClient(uaccount.NewClient)
	req := client.NewGetRegionRequest()
	resp, err := client.GetRegion(req)
	if err != nil {
		return "", "", err
	}
	if resp.RetCode != 0 {
		return "", "", fmt.Errorf("Something wrong. RetCode:%d, Message:%s", resp.RetCode, resp.Message)
	}
	for _, region := range resp.Regions {
		if region.IsDefault == true {
			return region.Region, region.Zone, nil
		}
	}
	return "", "", fmt.Errorf("No default region")
}

// Region region, zone, isDefault
type Region struct {
	Labels        map[string][]string
	DefaultRegion string
	DefaultZone   string
}

func fetchRegion() (*Region, error) {
	client := newServiceClient(uaccount.NewClient)
	req := client.NewGetRegionRequest()
	resp, err := client.GetRegion(req)
	if err != nil {
		return nil, err
	}
	region := &Region{
		Labels: make(map[string][]string),
	}
	for _, r := range resp.Regions {
		region.Labels[r.Region] = append(region.Labels[r.Region], r.Zone)
		if r.IsDefault {
			region.DefaultRegion = r.Region
			region.DefaultZone = r.Zone
		}
	}
	return region, nil
}

func fetchRegionWithConfig(cfg *platform.AggConfig) (*Region, error) {
	client, err := newServiceClientForConfig(cfg, uaccount.NewClient)
	if err != nil {
		return nil, err
	}
	req := client.NewGetRegionRequest()
	resp, err := client.GetRegion(req)
	if err != nil {
		return nil, err
	}
	region := &Region{
		Labels: make(map[string][]string),
	}
	for _, r := range resp.Regions {
		region.Labels[r.Region] = append(region.Labels[r.Region], r.Zone)
		if r.IsDefault {
			region.DefaultRegion = r.Region
			region.DefaultZone = r.Zone
		}
	}
	return region, nil
}

func getAllRegions() ([]string, error) {
	regionIns, err := fetchRegion()
	if err != nil {
		return nil, err
	}
	list := []string{}
	for region := range regionIns.Labels {
		list = append(list, region)
	}
	return list, nil
}

// 仅在命令补全中使用，忽略错误
func getRegionList() []string {
	regionIns, err := fetchRegion()
	if err != nil {
		return nil
	}
	list := []string{}
	for region := range regionIns.Labels {
		list = append(list, region)
	}
	return list
}

func getZoneList(region string) []string {
	regionIns, err := fetchRegion()
	if err != nil {
		return nil
	}
	list := []string{}
	if region == "" {
		for _, zones := range regionIns.Labels {
			list = append(list, zones...)
		}
	} else {
		list = regionIns.Labels[region]
	}
	return list
}

var errNoDefaultProject = errors.New("No default project")

func getDefaultProject() (string, string, error) {
	client := newServiceClient(uaccount.NewClient)
	req := client.NewGetProjectListRequest()

	resp, err := client.GetProjectList(req)
	if err != nil {
		return "", "", err
	}
	for _, project := range resp.ProjectSet {
		if project.IsDefault == true {
			return project.ProjectId, project.ProjectName, nil
		}
	}
	return "", "", errNoDefaultProject
}

func getDefaultProjectWithConfig(cfg *platform.AggConfig) (string, string, error) {
	client, err := newServiceClientForConfig(cfg, uaccount.NewClient)
	if err != nil {
		return "", "", err
	}

	req := client.NewGetProjectListRequest()
	resp, err := client.GetProjectList(req)
	if err != nil {
		return "", "", err
	}
	for _, project := range resp.ProjectSet {
		if project.IsDefault == true {
			return project.ProjectId, project.ProjectName, nil
		}
	}
	return "", "", errNoDefaultProject
}

// fetchProjectListWithConfig 用指定 profile 的凭证拉取完整项目列表（含默认标记）
func fetchProjectListWithConfig(cfg *platform.AggConfig) ([]uaccount.ProjectListInfo, error) {
	client, err := newServiceClientForConfig(cfg, uaccount.NewClient)
	if err != nil {
		return nil, err
	}

	req := client.NewGetProjectListRequest()
	resp, err := client.GetProjectList(req)
	if err != nil {
		return nil, err
	}
	return resp.ProjectSet, nil
}

func fetchProjectWithConfig(cfg *platform.AggConfig) (map[string]bool, error) {
	client, err := newServiceClientForConfig(cfg, uaccount.NewClient)
	if err != nil {
		return nil, err
	}

	req := client.NewGetProjectListRequest()
	resp, err := client.GetProjectList(req)
	if err != nil {
		return nil, err
	}

	projects := map[string]bool{}
	for _, project := range resp.ProjectSet {
		projects[project.ProjectId] = true
	}
	return projects, nil
}

func getReasonableProject(cfg *platform.AggConfig) (string, error) {
	if cfg.ProjectID == "" {
		id, _, err := getDefaultProjectWithConfig(cfg)
		if err != nil {
			// %w 而非 %v：调用方需 errors.Is 识别 errNoDefaultProject（账号有项目但
			// 未设默认，属良性），放行而非拒绝落盘，口径与 ucloud init 一致
			return "", fmt.Errorf("fetch project failed: %w", err)
		}
		return id, nil
	}

	projects, err := fetchProjectWithConfig(cfg)
	if err != nil {
		return "", fmt.Errorf("fetch project failed: %v", err)
	}
	if _, ok := projects[cfg.ProjectID]; !ok {
		return "", fmt.Errorf("project[%s] does not exist", cfg.ProjectID)
	}

	return cfg.ProjectID, nil
}

func isUserCertified(userInfo *uaccount.UserInfo) bool {
	return userInfo.AuthState == "CERTIFIED"
}

func getUserInfo() (*uaccount.UserInfo, error) {
	client := newServiceClient(uaccount.NewClient)
	req := client.NewGetUserInfoRequest()
	var userInfo uaccount.UserInfo
	resp, err := client.GetUserInfo(req)

	if err != nil {
		return nil, err
	}

	if resp.RetCode != 0 {
		return nil, fmt.Errorf("Something wrong. RetCode:%d, Message:%s", resp.RetCode, resp.Message)
	}
	if len(resp.DataSet) == 1 {
		userInfo = resp.DataSet[0]
		bytes, err := json.Marshal(userInfo)
		if err != nil {
			return nil, err
		}
		fileFullPath := platform.GetConfigDir() + "/user.json"
		err = ioutil.WriteFile(fileFullPath, bytes, 0600)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("GetUserInfo DataSet length: %d", len(resp.DataSet))
	}
	return &userInfo, nil
}
