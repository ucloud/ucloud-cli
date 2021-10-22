package http

import (
	"time"
)

const (
	MimeFormURLEncoded = "application/x-www-form-urlencoded"
	MimeJSON           = "application/json"
)

const (
	HeaderNameContentType = "Content-Type"
	HeaderNameUserAgent   = "User-Agent"
	HeaderUTimestampMs    = "U-Timestamp-Ms"
)

// DefaultHeaders defined default http headers
var DefaultHeaders = map[string]string{
	HeaderNameContentType: MimeFormURLEncoded,
	// "X-SDK-VERSION": VERSION,
}

// DefaultTimeout is the default timeout of each request
var DefaultTimeout = 30 * time.Second
