package service

import (
	"github.com/ucloud/ucloud-sdk-go/sdk"
	"github.com/ucloud/ucloud-sdk-go/sdk/auth"
	"github.com/ucloud/ucloud-sdk-go/service/pathx"
	"github.com/ucloud/ucloud-sdk-go/service/uaccount"
	"github.com/ucloud/ucloud-sdk-go/service/uhost"
	"github.com/ucloud/ucloud-sdk-go/service/ulb"
	"github.com/ucloud/ucloud-sdk-go/service/unet"
	"github.com/ucloud/ucloud-sdk-go/service/vpc"
)

type Client struct {
	uaccount.UAccountClient

	uhost.UHostClient

	unet.UNetClient

	ulb.ULBClient

	vpc.VPCClient

	pathx.PathXClient
}

func NewClient(config *sdk.ClientConfig, credential *auth.Credential) *Client {
	return &Client{
		*uaccount.NewClient(config, credential),

		*uhost.NewClient(config, credential),

		*unet.NewClient(config, credential),

		*ulb.NewClient(config, credential),

		*vpc.NewClient(config, credential),

		*pathx.NewClient(config, credential),
	}
}
