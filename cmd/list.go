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

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/service/uaccount"
	"github.com/ucloud/ucloud-sdk-go/service/uaccount/types"

	. "github.com/ucloud/ucloud-cli/util"
)

//NewCmdList ucloud ls
func NewCmdList() *cobra.Command {
	var listObject string
	var cmd = &cobra.Command{
		Use:     "ls",
		Short:   "List all regions or projects",
		Long:    "List all regions or projects",
		Example: "ucloud ls --object region; ucloud ls --object project",
		Run: func(cmd *cobra.Command, args []string) {
			switch listObject {
			case "region":
				if err := listRegion(); err != nil {
					fmt.Println(err)
				}
			case "project":
				if err := listProject(); err != nil {
					fmt.Println(err)
				}
			default:
				fmt.Println("object should be region or project")
			}
		},
	}
	cmd.Flags().StringVar(&listObject, "object", "", "Object to list,region or project. Required")
	cmd.MarkFlagRequired("object")
	return cmd
}

func getDefaultRegion() (string, error) {
	req := &uaccount.GetRegionRequest{}
	resp, err := BizClient.GetRegion(req)
	if err != nil {
		return "", err
	}
	if resp.RetCode != 0 {
		return "", fmt.Errorf("Something wrong. RetCode:%d, Message:%s", resp.RetCode, resp.Message)
	}
	for _, region := range resp.Regions {
		if region.IsDefault == true {
			return region.Region, nil
		}
	}
	return "", fmt.Errorf("No default region")
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
	var regionMap = map[string]bool{}
	var regionList []string
	for _, region := range resp.Regions {
		if _, ok := regionMap[region.Region]; !ok {
			regionList = append(regionList, region.Region)
		}
		regionMap[region.Region] = true
	}
	for index, region := range regionList {
		fmt.Printf("[%2d] %s\n", index, region)
	}
	return nil
}

func getDefaultProject() (string, error) {
	req := BizClient.NewGetProjectListRequest()
	resp, err := BizClient.GetProjectList(req)
	if err != nil {
		return "", err
	}
	if resp.RetCode != 0 {
		return "", fmt.Errorf("Something wrong. RetCode:%d, Message:%s", resp.RetCode, resp.Message)
	}
	for _, project := range resp.ProjectSet {
		if project.IsDefault == true {
			return project.ProjectId, nil
		}
	}
	return "", fmt.Errorf("No default project")
}

func listProject() error {
	req := &uaccount.GetProjectListRequest{}
	resp, err := BizClient.GetProjectList(req)
	if err != nil {
		return err
	}
	if resp.RetCode != 0 {
		return fmt.Errorf("Something wrong. RetCode:%d, Message:%s", resp.RetCode, resp.Message)
	}
	for _, project := range resp.ProjectSet {
		fmt.Printf("ProjectId: %s, ProjectName:%s\n", project.ProjectId, project.ProjectName)
	}
	return nil
}

func isUserCertified(userInfo *types.UserInfo) bool {
	return userInfo.AuthState == "CERTIFIED"
}

func getUserInfo() (*types.UserInfo, error) {
	req := BizClient.NewGetUserInfoRequest()
	var userInfo types.UserInfo
	resp, err := BizClient.GetUserInfo(req)

	if err != nil {
		return nil, err
	}

	if resp.RetCode != 0 {
		return nil, fmt.Errorf("Something wrong. RetCode:%d, Message:%s", resp.RetCode, resp.Message)
	}
	if len(resp.DataSet) == 1 {
		userInfo = resp.DataSet[0]
		Tracer.AppendInfo("userName", userInfo.UserEmail)
		Tracer.AppendInfo("companyName", userInfo.CompanyName)
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
