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
	"github.com/ucloud/ucloud-cli/util"
	"github.com/ucloud/ucloud-sdk-go/service/uaccount"
	"github.com/ucloud/ucloud-sdk-go/service/uaccount/types"
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
				listRegion()
			case "project":
				listProject()
			default:
				fmt.Println("object should be region or project")
			}
		},
	}
	cmd.Flags().StringVar(&listObject, "object", "", "Object to list,region or project. Required")
	cmd.MarkFlagRequired("object")
	return cmd
}

func listRegion() error {
	req := &uaccount.GetRegionRequest{}
	resp, err := client.GetRegion(req)
	if err != nil {
		return err
	}
	if resp.RetCode != 0 {
		return fmt.Errorf("Something wrong. RetCode:%d, Message:%s", resp.RetCode, resp.Message)
	}
	for _, region := range resp.Regions {
		fmt.Printf("Region: %s, Zone: %s\n", region.Region, region.Zone)
	}
	return nil
}

func listProject() error {
	req := &uaccount.GetProjectListRequest{}
	resp, err := client.GetProjectList(req)
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

func isUserCertified() (bool, error) {
	userInfo, err := getUserInfo()
	if err != nil {
		return false, err
	}
	return userInfo.AuthState == "CERTIFIED", nil
}

func getUserInfo() (*types.UserInfo, error) {
	req := client.NewGetUserInfoRequest()
	var userInfo types.UserInfo
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
		fileFullPath := util.GetConfigPath() + "/user.json"
		err = ioutil.WriteFile(fileFullPath, bytes, 0600)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("GetUserInfo DataSet length: %d", len(resp.DataSet))
	}
	return &userInfo, nil
}
