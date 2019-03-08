//Code is generated by ucloud code generator, don't modify it by hand, it will cause undefined behaviors.
//go:generate ucloud-gen-go-api UMem CreateUMemSpace

package umem

import (
	"encoding/base64"

	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
	"github.com/ucloud/ucloud-sdk-go/ucloud/response"
)

// CreateUMemSpaceRequest is request schema for CreateUMemSpace action
type CreateUMemSpaceRequest struct {
	request.CommonBase

	// [公共参数] 地域。 参见 [地域和可用区列表](../summary/regionlist.html)
	// Region *string `required:"true"`

	// [公共参数] 可用区。参见 [可用区列表](../summary/regionlist.html)
	// Zone *string `required:"false"`

	// [公共参数] 项目ID。不填写为默认项目，子帐号必须填写。 请参考[GetProjectList接口](../summary/get_project_list.html)
	// ProjectId *string `required:"false"`

	// 内存大小, 单位:GB, 范围[1~1024]
	Size *int `required:"true"`

	// 空间名称,长度(6<=size<=63)
	Name *string `required:"true"`

	// 协议:memcache, redis (默认redis).注意:redis无single类型
	Protocol *string `required:"false"`

	// 空间类型:single(无热备),double(热备)(默认: double)
	Type *string `required:"false"`

	// Year , Month, Dynamic, Trial 默认: Month
	ChargeType *string `required:"false"`

	// 购买时长 默认: 1
	Quantity *int `required:"false"`

	//
	Tag *string `required:"false"`

	// URedis密码。请遵照[[api:uhost-api:specification|字段规范]]设定密码。密码需使用base64进行编码，举例如下：# echo -n Password1 | base64UGFzc3dvcmQx。
	Password *string `required:"false"`

	// 使用的代金券id
	CouponId *string `required:"false"`

	// VPC 的 ID
	VPCId *string `required:"false"`

	// Subnet 的 ID
	SubnetId *string `required:"false"`
}

// CreateUMemSpaceResponse is response schema for CreateUMemSpace action
type CreateUMemSpaceResponse struct {
	response.CommonBase

	// 创建内存空间ID列表
	SpaceId string
}

// NewCreateUMemSpaceRequest will create request of CreateUMemSpace action.
func (c *UMemClient) NewCreateUMemSpaceRequest() *CreateUMemSpaceRequest {
	req := &CreateUMemSpaceRequest{}

	// setup request with client config
	c.client.SetupRequest(req)

	// setup retryable with default retry policy (retry for non-create action and common error)
	req.SetRetryable(false)
	return req
}

// CreateUMemSpace - 创建UMem内存空间
func (c *UMemClient) CreateUMemSpace(req *CreateUMemSpaceRequest) (*CreateUMemSpaceResponse, error) {
	var err error
	var res CreateUMemSpaceResponse
	req.Password = ucloud.String(base64.StdEncoding.EncodeToString([]byte(ucloud.StringValue(req.Password))))

	err = c.client.InvokeAction("CreateUMemSpace", req, &res)
	if err != nil {
		return &res, err
	}

	return &res, nil
}
