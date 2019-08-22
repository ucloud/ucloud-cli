package base

import (
	ppathx "github.com/ucloud/ucloud-sdk-go/private/services/pathx"
	pudb "github.com/ucloud/ucloud-sdk-go/private/services/udb"
	puhost "github.com/ucloud/ucloud-sdk-go/private/services/uhost"
	pumem "github.com/ucloud/ucloud-sdk-go/private/services/umem"
	"github.com/ucloud/ucloud-sdk-go/services/pathx"
	"github.com/ucloud/ucloud-sdk-go/services/uaccount"
	"github.com/ucloud/ucloud-sdk-go/services/udb"
	"github.com/ucloud/ucloud-sdk-go/services/udisk"
	"github.com/ucloud/ucloud-sdk-go/services/udpn"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/services/ulb"
	"github.com/ucloud/ucloud-sdk-go/services/umem"
	"github.com/ucloud/ucloud-sdk-go/services/unet"
	"github.com/ucloud/ucloud-sdk-go/services/uphost"
	"github.com/ucloud/ucloud-sdk-go/services/vpc"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
)

//PrivateUHostClient 私有模块的uhost client 即未在官网开放的接口
type PrivateUHostClient = puhost.UHostClient

//PrivateUDBClient 私有模块的udb client 即未在官网开放的接口
type PrivateUDBClient = pudb.UDBClient

//PrivateUMemClient 私有模块的umem client 即未在官网开放的接口
type PrivateUMemClient = pumem.UMemClient

//PrivatePathxClient 私有模块的pathx client 即未在官网开放的接口
type PrivatePathxClient = ppathx.PathXClient

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
	umem.UMemClient
	uphost.UPHostClient
	PrivateUHostClient
	PrivateUDBClient
	PrivateUMemClient
	PrivatePathxClient
}

// NewClient will return a aggregate client
func NewClient(config *sdk.Config, credential *auth.Credential) *Client {
	var handler sdk.RequestHandler = func(c *sdk.Client, req request.Common) (request.Common, error) {
		err := req.SetProjectId(PickResourceID(req.GetProjectId()))
		return req, err
	}
	var (
		uaccountClient = *uaccount.NewClient(config, credential)
		uhostClient    = *uhost.NewClient(config, credential)
		unetClient     = *unet.NewClient(config, credential)
		vpcClient      = *vpc.NewClient(config, credential)
		udpnClient     = *udpn.NewClient(config, credential)
		pathxClient    = *pathx.NewClient(config, credential)
		udiskClient    = *udisk.NewClient(config, credential)
		ulbClient      = *ulb.NewClient(config, credential)
		udbClient      = *udb.NewClient(config, credential)
		umemClient     = *umem.NewClient(config, credential)
		uphostClient   = *uphost.NewClient(config, credential)
		puhostClient   = *puhost.NewClient(config, credential)
		pudbClient     = *pudb.NewClient(config, credential)
		pumemClient    = *pumem.NewClient(config, credential)
		ppathxClient   = *ppathx.NewClient(config, credential)
	)

	uaccountClient.Client.AddRequestHandler(handler)
	uhostClient.Client.AddRequestHandler(handler)
	unetClient.Client.AddRequestHandler(handler)
	vpcClient.Client.AddRequestHandler(handler)
	udpnClient.Client.AddRequestHandler(handler)
	pathxClient.Client.AddRequestHandler(handler)
	udiskClient.Client.AddRequestHandler(handler)
	ulbClient.Client.AddRequestHandler(handler)
	udbClient.Client.AddRequestHandler(handler)
	umemClient.Client.AddRequestHandler(handler)
	uphostClient.Client.AddRequestHandler(handler)
	puhostClient.Client.AddRequestHandler(handler)
	pudbClient.Client.AddRequestHandler(handler)
	pumemClient.Client.AddRequestHandler(handler)
	ppathxClient.Client.AddRequestHandler(handler)

	return &Client{
		uaccountClient,
		uhostClient,
		unetClient,
		vpcClient,
		udpnClient,
		pathxClient,
		udiskClient,
		ulbClient,
		udbClient,
		umemClient,
		uphostClient,
		puhostClient,
		pudbClient,
		pumemClient,
		ppathxClient,
	}
}
