package http

import (
	"time"
)

// DefaultHeaders defined default http headers
var DefaultHeaders = map[string]string{
	"Content-Type": "application/x-www-form-urlencoded",
	// "X-SDK-VERSION": VERSION,
}

var DefaultTimeout = 30 * time.Second
