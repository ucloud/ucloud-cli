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

	"github.com/spf13/cobra"
)

//NewCmdUHost ucloud uhost
func NewCmdUHost() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "uhost",
		Short: "UHost managment",
		Long:  `UHost managment. Only list`,
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(NewCmdUHostList())
	return cmd
}

//NewCmdUHostList ucloud uhost list
func NewCmdUHostList() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "ls",
		Short: "List all UHost Instances",
		Long:  `List all UHost Instances`,
		Run: func(cmd *cobra.Command, args []string) {
			req := client.NewDescribeUHostInstanceRequest()
			bindGlobalParam(req)
			resp, err := client.DescribeUHostInstance(req)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			for _, uhost := range resp.UHostSet {
				fmt.Printf("UHostID:%s\n", uhost.UHostId)
			}
		},
	}
	return cmd
}
