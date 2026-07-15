package uk8s

import (
	"encoding/json"

	"github.com/ucloud/ucloud-sdk-go/ucloud/response"
)

// compatResponse keeps the full wire response for UK8S actions whose SDK
// response models lag behind the API. Embedding CommonBase preserves SDK
// RetCode handling while Payload retains every field without a type mismatch.
type compatResponse struct {
	response.CommonBase
	Payload map[string]interface{}
}

func (r *compatResponse) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &r.CommonBase); err != nil {
		return err
	}
	return json.Unmarshal(data, &r.Payload)
}

func (r *compatResponse) stringField(name string) string {
	value, _ := r.Payload[name].(string)
	return value
}
