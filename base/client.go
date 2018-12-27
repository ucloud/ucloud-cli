package base

import (
	pudisk "github.com/ucloud/ucloud-sdk-go/private/services/udisk"
	puhost "github.com/ucloud/ucloud-sdk-go/private/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/services/pathx"
	"github.com/ucloud/ucloud-sdk-go/services/uaccount"
	"github.com/ucloud/ucloud-sdk-go/services/udisk"
	"github.com/ucloud/ucloud-sdk-go/services/udpn"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/services/unet"
	"github.com/ucloud/ucloud-sdk-go/services/vpc"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

//PrivateUDiskClient 私有模块的udisk client 即未在官网开放的接口
type PrivateUDiskClient = pudisk.UDiskClient

//PrivateUHostClient 私有模块的udisk client 即未在官网开放的接口
type PrivateUHostClient = puhost.UHostClient

//Client aggregate client for business
type Client struct {
	uaccount.UAccountClient
	uhost.UHostClient
	unet.UNetClient
	vpc.VPCClient
	udpn.UDPNClient
	pathx.PathXClient
	udisk.UDiskClient
	PrivateUHostClient
}

// NewClient will return a aggregate client
func NewClient(config *ucloud.Config, credential *auth.Credential) *Client {
	return &Client{
		*uaccount.NewClient(config, credential),
		*uhost.NewClient(config, credential),
		*unet.NewClient(config, credential),
		*vpc.NewClient(config, credential),
		*udpn.NewClient(config, credential),
		*pathx.NewClient(config, credential),
		*udisk.NewClient(config, credential),
		*puhost.NewClient(config, credential),
	}
}
