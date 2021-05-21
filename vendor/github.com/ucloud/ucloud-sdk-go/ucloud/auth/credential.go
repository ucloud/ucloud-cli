/*
Package auth is the credential utilities of sdk
*/
package auth

import (
	"net/url"
	"time"
)

// Credential is the information of credential keys
type Credential struct {
	// UCloud Public Key
	PublicKey string

	// UCloud Private Key
	PrivateKey string

	// UCloud STS token for temporary usage
	SecurityToken string

	// Time the credentials will expire.
	CanExpire bool
	Expires   time.Time
}

// NewCredential will return credential config with default values
func NewCredential() Credential {
	return Credential{}
}

// CreateSign will encode query string to credential signature.
func (c *Credential) CreateSign(query string) string {
	payload := queryToMap(query)
	return c.VerifyAc(payload)
}

// BuildCredentialedQuery will build query string with signature query param.
func (c *Credential) BuildCredentialedQuery(params map[string]string) string {
	payload := make(map[string]interface{})
	for k, v := range params {
		payload[k] = v
	}
	return mapToQuery(c.Apply(payload))
}

// Apply will return payload with credential and signature
func (c *Credential) Apply(payload map[string]interface{}) map[string]interface{} {
	payload = c.applyDefaults(payload)
	payload["Signature"] = sign(c.applyDefaults(payload), c.PrivateKey)
	return payload
}

// VerifyAc will return payload with credential and signature
func (c *Credential) VerifyAc(payload map[string]interface{}) string {
	return sign(c.applyDefaults(payload), c.PrivateKey)
}

// IsExpired will return if the credential is expired
func (c *Credential) IsExpired() bool {
	return c.CanExpire && time.Now().After(c.Expires)
}

func (c *Credential) applyDefaults(payload map[string]interface{}) map[string]interface{} {
	values := make(map[string]interface{})
	for k, v := range payload {
		values[k] = v
	}
	if len(c.SecurityToken) != 0 {
		values["SecurityToken"] = c.SecurityToken
	}
	if len(c.PublicKey) != 0 {
		values["PublicKey"] = c.PublicKey
	}
	return values
}

func queryToMap(query string) map[string]interface{} {
	values := make(map[string]interface{})
	urlValues, err := url.ParseQuery(query)
	if err != nil {
		return values
	}
	for k, v := range urlValues {
		if len(v) != 0 {
			values[k] = v[0]
		}
	}
	return values
}

func mapToQuery(values map[string]interface{}) string {
	urlValues := url.Values{}
	for k, v := range values {
		urlValues.Set(k, any2String(v))
	}
	return urlValues.Encode()
}
