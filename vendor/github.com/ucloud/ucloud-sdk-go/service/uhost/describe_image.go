//go:generate go run ../../private/cli/gen-api/main.go uhost DescribeImage

package uhost

import (
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
	. "github.com/ucloud/ucloud-sdk-go/service/uhost/types"
)

type DescribeImageRequest struct {
	request.CommonBase

	// Optional, 可用区。参见 [可用区列表](../summary/regionlist.html)
	Zone string

	// Optional, 镜像类型。标准镜像：Base，镜像市场：Business， 自定义镜像：Custom，默认返回所有类型
	ImageType string

	// Optional, 操作系统类型：Linux， Windows 默认返回所有类型
	OsType string

	// Optional, 镜像Id
	ImageId string

	// Optional, 列表起始位置偏移量，默认为0
	Offset int

	// Optional, 返回数据长度，默认为20
	Limit int

	// Optional, 是否返回价格：1返回，0不返回；默认不返回
	PriceSet int
}

type DescribeImageResponse struct {
	response.CommonBase

	// 满足条件的镜像总数
	TotalCount int

	// 镜像列表详见 UHostImageSet
	ImageSet []UHostImageSet
}

// NewDescribeImageRequest will create request of DescribeImage action.
func (c *UHostClient) NewDescribeImageRequest() *DescribeImageRequest {
	cfg := c.client.GetConfig()

	return &DescribeImageRequest{
		CommonBase: request.CommonBase{
			Region:    cfg.Region,
			ProjectId: cfg.ProjectId,
		},
	}
}

// DescribeImage - 获取指定数据中心镜像列表，用户可通过指定操作系统类型，镜像Id进行过滤。
func (c *UHostClient) DescribeImage(req *DescribeImageRequest) (*DescribeImageResponse, error) {
	var err error
	var res DescribeImageResponse

	err = c.client.InvokeAction("DescribeImage", req, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
