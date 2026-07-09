package ulhost

import (
	"github.com/ucloud/ucloud-cli/base"
	internalulhost "github.com/ucloud/ucloud-cli/products/ulhost/internal/ulhost"
)

// Spoller is the package-level ulhost poller, replacing the old ulhostSpoller
// from cmd/ulhost.go. It is used by cmd/api.go's RepeatsSupportedAPI for
// CreateULHostInstance polling.
var Spoller = base.NewSpoller(internalulhost.SdescribeULHostByID, base.Cxt.GetWriter())
