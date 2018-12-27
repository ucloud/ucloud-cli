package udpn

import (
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

// UDPNClient is the client of UDPN
type UDPNClient struct {
	client *ucloud.Client
}

// NewClient will return a instance of UDPNClient
func NewClient(config *ucloud.Config, credential *auth.Credential) *UDPNClient {
	client := ucloud.NewClient(config, credential)
	return &UDPNClient{
		client: client,
	}
}
