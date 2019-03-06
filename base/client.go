package base

import (
	pudb "github.com/ucloud/ucloud-sdk-go/private/services/udb"
	puhost "github.com/ucloud/ucloud-sdk-go/private/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/services/pathx"
	"github.com/ucloud/ucloud-sdk-go/services/uaccount"
	"github.com/ucloud/ucloud-sdk-go/services/udb"
	"github.com/ucloud/ucloud-sdk-go/services/udisk"
	"github.com/ucloud/ucloud-sdk-go/services/udpn"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/services/ulb"
	"github.com/ucloud/ucloud-sdk-go/services/unet"
	"github.com/ucloud/ucloud-sdk-go/services/vpc"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

//PrivateUHostClient 私有模块的uhost client 即未在官网开放的接口
type PrivateUHostClient = puhost.UHostClient

//PrivateUDBClient 私有模块的udb client 即未在官网开放的接口
type PrivateUDBClient = pudb.UDBClient

//Client aggregate client for business
type Client struct {
	uaccount.UAccountClient
	uhost.UHostClient
	unet.UNetClient
	vpc.VPCClient
	udpn.UDPNClient
	pathx.PathXClient
	udisk.UDiskClient
	ulb.ULBClient
	udb.UDBClient
	PrivateUHostClient
	PrivateUDBClient
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
		*ulb.NewClient(config, credential),
		*udb.NewClient(config, credential),
		*puhost.NewClient(config, credential),
		*pudb.NewClient(config, credential),
	}
}
