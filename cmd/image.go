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
	"strings"

	"github.com/spf13/cobra"

	. "github.com/ucloud/ucloud-cli/base"
)

//NewCmdUImage ucloud uimage
func NewCmdUImage() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "image",
		Short: "List images",
		Long:  `List images`,
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(NewCmdUImageList())

	return cmd
}

//ImageRow 表格行
type ImageRow struct {
	ImageName         string
	ImageID           string
	BasicImage        string
	ExtensibleFeature string
	CreationTime      string
	State             string
}

//NewCmdUImageList ucloud uimage list
func NewCmdUImageList() *cobra.Command {
	req := BizClient.NewDescribeImageRequest()
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List image",
		Long:    "List image",
		Example: "ucloud image list --image-type Base",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := BizClient.DescribeImage(req)
			if err != nil {
				HandleError(err)
				return
			}
			if global.json {
				PrintJSON(resp.ImageSet)
			} else {
				list := make([]ImageRow, 0)
				for _, image := range resp.ImageSet {
					row := ImageRow{}
					row.ImageName = image.ImageName
					row.ImageID = image.ImageId
					row.BasicImage = image.OsName
					row.ExtensibleFeature = strings.Join(image.Features, ",")
					row.CreationTime = FormatDate(image.CreateTime)
					row.State = image.State
					if row.State == "Available" {
						list = append(list, row)
					}
				}
				PrintTable(list, []string{"ImageName", "ImageID", "BasicImage", "ExtensibleFeature", "CreationTime", "State"})
			}
		},
	}
	req.ProjectId = cmd.Flags().String("project-id", ConfigInstance.ProjectID, "Assign project-id")
	req.Region = cmd.Flags().String("region", ConfigInstance.Region, "Assign region")
	req.Zone = cmd.Flags().String("zone", ConfigInstance.Zone, "Assign availability zone")
	req.ImageType = cmd.Flags().String("image-type", "", "'Base',Standard image; 'Business',image market; 'Custom',custom image; Return all types by default")
	req.OsType = cmd.Flags().String("os-type", "", "Linux or Windows. Return all types by default")
	req.ImageId = cmd.Flags().String("image-id", "", "iamge id such as 'uimage-xxx'")
	req.Offset = cmd.Flags().Int("offset", 0, "offset default 0")
	req.Limit = cmd.Flags().Int("limit", 500, "max count")
	return cmd
}
