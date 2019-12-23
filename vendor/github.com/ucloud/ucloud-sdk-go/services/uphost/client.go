package uphost

import (
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

// UPHostClient is the client of UPHost
type UPHostClient struct {
	*ucloud.Client
}

// NewClient will return a instance of UPHostClient
func NewClient(config *ucloud.Config, credential *auth.Credential) *UPHostClient {
	client := ucloud.NewClient(config, credential)
	return &UPHostClient{
		client,
	}
}
