package base

import (
	"github.com/ucloud/ucloud-sdk-go/services/pathx"
	"github.com/ucloud/ucloud-sdk-go/services/uaccount"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/services/ulb"
	"github.com/ucloud/ucloud-sdk-go/services/unet"
	"github.com/ucloud/ucloud-sdk-go/services/vpc"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

//Client aggregate client for business
type Client struct {
	uaccount.UAccountClient
	uhost.UHostClient
	unet.UNetClient
	ulb.ULBClient
	vpc.VPCClient
	pathx.PathXClient
}

// NewClient will return a aggregate client
func NewClient(config *ucloud.Config, credential *auth.Credential) *Client {
	return &Client{
		*uaccount.NewClient(config, credential),
		*uhost.NewClient(config, credential),
		*unet.NewClient(config, credential),
		*ulb.NewClient(config, credential),
		*vpc.NewClient(config, credential),
		*pathx.NewClient(config, credential),
	}
}
