package udisk

import (
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

// UDiskClient is the client of UDisk
type UDiskClient struct {
	client *ucloud.Client
}

// NewClient will return a instance of UDiskClient
func NewClient(config *ucloud.Config, credential *auth.Credential) *UDiskClient {
	client := ucloud.NewClient(config, credential)
	return &UDiskClient{
		client: client,
	}
}
