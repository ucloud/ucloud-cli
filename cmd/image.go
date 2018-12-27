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
	"io"
	"strings"

	"github.com/spf13/cobra"

	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/base"
	"github.com/ucloud/ucloud-cli/model/cli"
	"github.com/ucloud/ucloud-cli/model/status"
)

//NewCmdUImage ucloud uimage
func NewCmdUImage() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "image",
		Short: "List and manipulate images",
		Long:  `List and manipulate images`,
		Args:  cobra.NoArgs,
	}
	writer := base.Cxt.GetWriter()
	cmd.AddCommand(NewCmdUImageList())
	cmd.AddCommand(NewCmdImageCopy(writer))
	cmd.AddCommand(NewCmdUImageDelete())
	createImageCmd := NewCmdUhostCreateImage(writer)
	createImageCmd.Use = "create"
	cmd.AddCommand(createImageCmd)

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
	req := base.BizClient.NewDescribeImageRequest()
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List image",
		Long:    "List image",
		Example: "ucloud image list --image-type Base",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := base.BizClient.DescribeImage(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			if global.json {
				base.PrintJSON(resp.ImageSet)
			} else {
				list := make([]ImageRow, 0)
				for _, image := range resp.ImageSet {
					row := ImageRow{}
					row.ImageName = image.ImageName
					row.ImageID = image.ImageId
					row.BasicImage = image.OsName
					row.ExtensibleFeature = strings.Join(image.Features, ",")
					row.CreationTime = base.FormatDate(image.CreateTime)
					row.State = image.State
					if row.State == "Available" {
						list = append(list, row)
					}
				}
				base.PrintTable(list, []string{"ImageName", "ImageID", "BasicImage", "ExtensibleFeature", "CreationTime"})
			}
		},
	}
	req.ProjectId = cmd.Flags().String("project-id", base.ConfigInstance.ProjectID, "Optional. Assign project-id")
	req.Region = cmd.Flags().String("region", base.ConfigInstance.Region, "Optional. Assign region")
	req.Zone = cmd.Flags().String("zone", "", "Optional. Assign availability zone")
	req.ImageType = cmd.Flags().String("image-type", "", "Optional. 'Base',Standard image; 'Business',image market; 'Custom',custom image; Return all types by default")
	req.OsType = cmd.Flags().String("os-type", "", "Optional. Linux or Windows. Return all types by default")
	req.ImageId = cmd.Flags().String("image-id", "", "Optional. Resource ID of image")
	req.Offset = cmd.Flags().Int("offset", 0, "Optional. Offset default 0")
	req.Limit = cmd.Flags().Int("limit", 500, "Optional. Max count")
	cmd.Flags().SetFlagValues("image-type", "Base", "Business", "Custom")
	return cmd
}

// func NewCmdImageImport() *cobra.Command {
// 	req := BizClient.NewImportCustomImageRequest()
// }

//NewCmdUImageDelete ucloud image delete
func NewCmdUImageDelete() *cobra.Command {
	var imageIDs *[]string
	req := base.BizClient.NewTerminateCustomImageRequest()
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete custom images",
		Long:  "Delete custom images",
		Run: func(cmd *cobra.Command, args []string) {
			for _, id := range *imageIDs {
				req.ImageId = sdk.String(base.PickResourceID(id))
				resp, err := base.BizClient.TerminateCustomImage(req)
				if err != nil {
					base.HandleError(err)
					return
				}
				base.Cxt.Printf("image[%s] deleted\n", resp.ImageId)
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	imageIDs = cmd.Flags().StringSlice("image-id", nil, "Required. Resource ID of images")
	req.ProjectId = cmd.Flags().String("project-id", base.ConfigInstance.ProjectID, "Optional. Assign project-id")
	req.Region = cmd.Flags().String("region", base.ConfigInstance.Region, "Optional. Assign region")
	req.Zone = cmd.Flags().String("zone", "", "Optional. Assign availability zone")
	cmd.MarkFlagRequired("image-id")
	flags.SetFlagValuesFunc("image-id", func() []string {
		return getImageList([]string{status.IMAGE_AVAILABLE, status.IMAGE_COPYING, status.IMAGE_MAKING}, cli.IAMGE_CUSTOM, *req.ProjectId, *req.Region, "")
	})
	return cmd
}

//NewCmdImageCopy ucloud image copy
func NewCmdImageCopy(out io.Writer) *cobra.Command {
	var imageIDs *[]string
	var async *bool
	req := base.BizClient.NewCopyCustomImageRequest()
	cmd := &cobra.Command{
		Use:   "copy",
		Short: "Copy custom images",
		Long:  "Copy custom images",
		Run: func(c *cobra.Command, args []string) {
			*req.ProjectId = base.PickResourceID(*req.ProjectId)
			*req.TargetProjectId = base.PickResourceID(*req.TargetProjectId)
			for _, id := range *imageIDs {
				id = base.PickResourceID(id)
				req.SourceImageId = &id
				resp, err := base.BizClient.CopyCustomImage(req)
				if err != nil {
					base.HandleError(err)
					return
				}
				text := fmt.Sprintf("image[%s] is coping", resp.TargetImageId)
				if *async {
					fmt.Fprintln(out, text)
				} else {
					poller := base.NewPoller(describeImageByID, out)
					poller.Poll(resp.TargetImageId, *req.TargetProjectId, *req.TargetRegion, "", text, []string{status.IMAGE_AVAILABLE, status.IMAGE_UNAVAILABLE})
				}
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	imageIDs = cmd.Flags().StringSlice("source-image-id", nil, "Required. Resource ID of source image")
	req.ProjectId = cmd.Flags().String("project-id", base.ConfigInstance.ProjectID, "Optional. Assign project-id")
	req.Region = cmd.Flags().String("region", base.ConfigInstance.Region, "Optional. Assign region")
	req.Zone = cmd.Flags().String("zone", base.ConfigInstance.Zone, "Optional. Assign availability zone")
	req.TargetRegion = flags.String("target-region", base.ConfigInstance.Region, "Optional. Target region. See 'ucloud region'")
	req.TargetProjectId = flags.String("target-project", base.ConfigInstance.ProjectID, "Optional. Target Project ID. See 'ucloud project list'")
	req.TargetImageName = flags.String("target-image-name", "", "Optional. Name of target image")
	req.TargetImageDescription = flags.String("target-image-desc", "", "Optional. Description of target image")
	async = flags.Bool("async", false, "Optional. Do not wait for the long-running operation to finish.")

	flags.SetFlagValuesFunc("source-image-id", func() []string {
		return getImageList([]string{status.IMAGE_AVAILABLE}, cli.IAMGE_CUSTOM, *req.ProjectId, *req.Region, *req.Zone)
	})
	flags.SetFlagValuesFunc("project-id", getProjectList)
	flags.SetFlagValuesFunc("region", getRegionList)
	flags.SetFlagValuesFunc("zone", getZoneList)
	flags.SetFlagValuesFunc("target-region", getRegionList)
	flags.SetFlagValuesFunc("target-project", getProjectList)

	cmd.MarkFlagRequired("source-image-id")

	return cmd
}

func getImageList(states []string, imageType, project, region, zone string) []string {
	req := base.BizClient.NewDescribeImageRequest()
	req.ProjectId = &project
	req.Region = &region
	req.Zone = &zone
	req.Limit = sdk.Int(1000)
	if imageType != cli.IMAGE_ALL {
		req.ImageType = sdk.String(imageType)
	}
	resp, err := base.BizClient.DescribeImage(req)
	if err != nil {
		return nil
	}
	list := []string{}
	for _, image := range resp.ImageSet {
		for _, s := range states {
			if image.State == s {
				list = append(list, image.ImageId+"/"+image.ImageName)
			}
		}
	}
	return list
}

func describeImageByID(imageID, project, region, zone string) (interface{}, error) {
	req := base.BizClient.NewDescribeImageRequest()
	req.ImageId = sdk.String(imageID)
	req.ProjectId = sdk.String(project)
	req.Region = sdk.String(region)
	req.Zone = sdk.String(zone)
	req.Limit = sdk.Int(50)
	resp, err := base.BizClient.DescribeImage(req)
	if err != nil {
		return nil, err
	}
	if len(resp.ImageSet) < 1 {
		return nil, nil
	}
	return &resp.ImageSet[0], nil
}
