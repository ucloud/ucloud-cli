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
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/uaccount"

	"github.com/ucloud/ucloud-cli/base"
)

//NewCmdProject ucloud project
func NewCmdProject() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "project",
		Short:   "List,create,update and delete project",
		Long:    "List,create,update and delete project",
		Example: "ucloud project",
	}
	cmd.AddCommand(NewCmdProjectList())
	cmd.AddCommand(NewCmdProjectCreate())
	cmd.AddCommand(NewCmdProjectUpdate())
	cmd.AddCommand(NewCmdProjectDelete())
	return cmd
}

//NewCmdProjectList ucloud project list
func NewCmdProjectList() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List project",
		Long:    "List project",
		Example: "ucloud project list",
		Run: func(cmd *cobra.Command, args []string) {
			listProject()
		},
	}
	return cmd
}

//NewCmdProjectCreate ucloud project create
func NewCmdProjectCreate() *cobra.Command {
	req := base.BizClient.NewCreateProjectRequest()
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create project",
		Long:    "Create project",
		Example: "ucloud project create --name xxx",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := base.BizClient.CreateProject(req)
			if err != nil {
				base.Cxt.PrintErr(err)
			} else {
				if resp.RetCode != 0 {
					base.HandleBizError(resp)
				} else {
					base.Cxt.Printf("Project:%q created successfully.\n", resp.ProjectId)
				}
			}
		},
	}
	req.ProjectName = cmd.Flags().String("name", "", "The name of project")
	req.ParentId = cmd.Flags().String("parent-id", "", "The parent project id")
	cmd.MarkFlagRequired("name")
	return cmd
}

//NewCmdProjectUpdate ucloud project update
func NewCmdProjectUpdate() *cobra.Command {
	req := base.BizClient.NewModifyProjectRequest()
	cmd := &cobra.Command{
		Use:     "update",
		Short:   "Update project name",
		Long:    "Update project name",
		Example: "ucloud project update --id org-xxx --name new_name",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := base.BizClient.ModifyProject(req)
			if err != nil {
				base.Cxt.PrintErr(err)
			} else {
				if resp.RetCode != 0 {
					base.HandleBizError(resp)
				} else {
					base.Cxt.Printf("Project:%s updated successfully.\n", *req.ProjectId)
				}
			}
		},
	}
	req.ProjectId = cmd.Flags().String("id", "", "The project id")
	req.ProjectName = cmd.Flags().String("name", "", "The new name of project")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("id")
	return cmd
}

//NewCmdProjectDelete ucloud project delete
func NewCmdProjectDelete() *cobra.Command {
	req := base.BizClient.NewTerminateProjectRequest()
	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete project",
		Long:    "Delete project",
		Example: "ucloud project delete --id org-xxx",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := base.BizClient.TerminateProject(req)
			if err != nil {
				base.Cxt.PrintErr(err)
			} else {
				if resp.RetCode != 0 {
					base.HandleBizError(resp)
				} else {
					base.Cxt.Printf("Project:%s deleted successfully.\n", *req.ProjectId)
				}
			}
		},
	}
	req.ProjectId = cmd.Flags().String("id", "", "The project id")
	cmd.MarkFlagRequired("id")
	return cmd
}

func listProject() error {
	req := &uaccount.GetProjectListRequest{}
	resp, err := base.BizClient.GetProjectList(req)
	if err != nil {
		return err
	}
	if resp.RetCode != 0 {
		return base.HandleBizError(resp)
	}
	if global.json {
		base.PrintJSON(resp.ProjectSet)
	} else {
		base.PrintTable(resp.ProjectSet, []string{"ProjectId", "ProjectName"})
	}
	return nil
}

func getProjectList() []string {
	req := &uaccount.GetProjectListRequest{}
	resp, err := base.BizClient.GetProjectList(req)
	if err != nil {
		return nil
	}
	list := []string{}
	for _, p := range resp.ProjectSet {
		list = append(list, p.ProjectId+"/"+p.ProjectName)
	}
	return list
}
