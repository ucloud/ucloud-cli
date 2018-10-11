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
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/uaccount"

	. "github.com/ucloud/ucloud-cli/base"
)

//NewCmdRegion ucloud region
func NewCmdRegion() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "region",
		Short:   "List all region and zone",
		Long:    "List all region and zone",
		Example: "ucloud region",
		Run: func(cmd *cobra.Command, args []string) {
			if err := listRegion(); err != nil {
				Cxt.PrintErr(err)
			}
		},
	}
	return cmd
}

func getDefaultRegion() (string, string, error) {
	req := &uaccount.GetRegionRequest{}
	resp, err := BizClient.GetRegion(req)
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

//RegionTable 为显示region表格创建的类型
type RegionTable struct {
	Region string
	Zones  string
}

func listRegion() error {
	req := &uaccount.GetRegionRequest{}
	resp, err := BizClient.GetRegion(req)
	if err != nil {
		return err
	}
	if resp.RetCode != 0 {
		return fmt.Errorf("Something wrong. RetCode:%d, Message:%s", resp.RetCode, resp.Message)
	}
	regionList := make([]RegionTable, 0)
	regionMap := make(map[string][]string)
	for _, region := range resp.Regions {
		regionMap[region.Region] = append(regionMap[region.Region], region.Zone)
	}
	for region, zones := range regionMap {
		regionList = append(regionList, RegionTable{region, strings.Join(zones, ", ")})
	}
	if global.json {
		PrintJSON(regionList)
	} else {
		PrintTable(regionList, []string{"Region", "Zones"})
	}
	return err
}

func getDefaultProject() (string, string, error) {
	req := BizClient.NewGetProjectListRequest()
	resp, err := BizClient.GetProjectList(req)
	if err != nil {
		return "", "", err
	}
	if resp.RetCode != 0 {
		return "", "", fmt.Errorf("Something wrong. RetCode:%d, Message:%s", resp.RetCode, resp.Message)
	}
	for _, project := range resp.ProjectSet {
		if project.IsDefault == true {
			return project.ProjectId, project.ProjectName, nil
		}
	}
	return "", "", fmt.Errorf("No default project")
}

func isUserCertified(userInfo *uaccount.UserInfo) bool {
	return userInfo.AuthState == "CERTIFIED"
}

func getUserInfo() (*uaccount.UserInfo, error) {
	req := BizClient.NewGetUserInfoRequest()
	var userInfo uaccount.UserInfo
	resp, err := BizClient.GetUserInfo(req)

	if err != nil {
		return nil, err
	}

	if resp.RetCode != 0 {
		return nil, fmt.Errorf("Something wrong. RetCode:%d, Message:%s", resp.RetCode, resp.Message)
	}
	if len(resp.DataSet) == 1 {
		userInfo = resp.DataSet[0]
		Cxt.AppendInfo("userName", userInfo.UserEmail)
		Cxt.AppendInfo("companyName", userInfo.CompanyName)
		bytes, err := json.Marshal(userInfo)
		if err != nil {
			return nil, err
		}
		fileFullPath := GetConfigPath() + "/user.json"
		err = ioutil.WriteFile(fileFullPath, bytes, 0600)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("GetUserInfo DataSet length: %d", len(resp.DataSet))
	}
	return &userInfo, nil
}
