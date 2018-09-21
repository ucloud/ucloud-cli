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
	. "github.com/ucloud/ucloud-cli/util"
)

//NewCmdUHost ucloud uhost
func NewCmdUHost() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "uhost",
		Short: "List UHost instance",
		Long:  `List UHost instance`,
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(NewCmdUHostList())
	return cmd
}

//NewCmdUHostList [ucloud uhost list]
func NewCmdUHostList() *cobra.Command {
	req := BizClient.NewDescribeUHostInstanceRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all UHost Instances",
		Long:  `List all UHost Instances`,
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := BizClient.DescribeUHostInstance(req)
			if err != nil {
				Tracer.Println(err)
				return
			}
			if resp.RetCode != 0 {
				HandleBizError(resp)
			} else {
				PrintTable(resp.UHostSet, []string{"UHostId", "Name", "UHostType", "Zone", "Tag", "State"})
			}
		},
	}
	cmd.Flags().SortFlags = false
	cmd.Flags().StringVar(&req.Region, "region", ConfigInstance.Region, "Assign region(override default region of your config)")
	cmd.Flags().StringVar(&req.Zone, "zone", "", "Zone")
	cmd.Flags().StringVar(&req.ProjectId, "project-id", ConfigInstance.ProjectID, "Assign project-id(override default projec-id of your config)")
	cmd.Flags().StringSliceVar(&req.UHostIds, "uhost-id", make([]string, 0), "UHost Instance ID, multiple values separated by comma(without space)")
	cmd.Flags().StringVar(&req.Tag, "tag", "", "UGroup")
	cmd.Flags().IntVar(&req.Offset, "offset", 0, "offset default 0")
	cmd.Flags().IntVar(&req.Limit, "limit", 20, "limit default 20, max value 100")

	return cmd
}

//NewCmdUHostCreate [ucloud uhost create]
func NewCmdUHostCreate() *cobra.Command {
	req := BizClient.NewCreateUHostInstanceRequest()
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create UHost Instance",
		Long:  "Create UHost Instance",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := BizClient.CreateUHostInstance(req)
			if err != nil {
				Tracer.Println(err)
				return
			}
			if resp.RetCode != 0 {
				HandleBizError(resp)
			} else {
				Tracer.Println(resp)
			}
		},
	}

	cmd.Flags().SortFlags = false

	cmd.Flags().StringVar(&req.Region, "region", ConfigInstance.Region, "Assign region(override default region of your config)")
	cmd.Flags().StringVar(&req.Zone, "zone", "", "Zone")
	cmd.Flags().StringVar(&req.ProjectId, "project-id", ConfigInstance.ProjectID, "Assign project-id(override default projec-id of your config)")
	cmd.Flags().StringVar(&req.ImageId, "image-id", "", "The ID of image. Obtain by 'ucloud image list'")
	cmd.Flags().StringVar(&req.Password, "password", "", "Password of the uhost user")

	return cmd
}
