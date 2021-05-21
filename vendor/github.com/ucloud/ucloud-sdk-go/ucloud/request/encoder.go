package request

import (
	"encoding/base64"
	"github.com/ucloud/ucloud-sdk-go/private/protocol/http"
)

type Encoder interface {
	Encode(req Common) (*http.HttpRequest, error)
}

// ToBase64Query will encode a wrapped string as base64 wrapped string
func ToBase64Query(s *string) *string {
	return String(base64.StdEncoding.EncodeToString([]byte(StringValue(s))))
}

// Deprecated: ToQueryMap will convert a request to string map
func ToQueryMap(req Common) (map[string]string, error) {
	return EncodeForm(req)
}
