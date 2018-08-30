package utils

import (
	"fmt"
	"net"
	"strings"
)

var avaliableHTTPMethod = []string{"GET", "POST", "PUT", "DELETE", "OPTION", "HEAD", "PATCH"}

// IsAvaliableMethod will check if a string is an avaliable http method.
func IsAvaliableMethod(method string) bool {
	for _, m := range avaliableHTTPMethod {
		if m == strings.ToUpper(method) {
			return true
		}
	}
	return false
}

// IsZeroValue will check any value if it is a zero value of it't type.
func IsZeroValue(expr interface{}) bool {
	if expr == nil {
		return true
	}

	switch v := expr.(type) {
	case bool:
		return false
	case string:
		return len(v) == 0
	case []byte:
		return len(v) == 0
	case int:
		return v == int(0)
	case int32:
		return v == int32(0)
	case int64:
		return v == int64(0)
	case uint:
		return v == uint(0)
	case uint32:
		return v == uint32(0)
	case uint64:
		return v == uint64(0)
	case float32:
		return v == float32(0.0)
	case float64:
		return v == float64(0.0)
	default:
		panic(fmt.Sprintf("unexpected type %T: %v", v, v))
	}
}

// IsErrorHTTPStatusCode will check a http status is error
func IsErrorHTTPStatusCode(code int) bool {
	if 400 <= code && code < 600 {
		return true
	}
	return false
}

// IsRetryableHTTPStatusCode will check a http status is retryable
func IsRetryableHTTPStatusCode(code int) bool {
	retryableCodes := [...]int{429, 502, 503, 504}
	for _, retryableCode := range retryableCodes {
		if code == retryableCode {
			return true
		}
	}
	return false
}

// IsTimeoutError will check if the error raise from network timeout
func IsTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	netError, isNetError := err.(net.Error)
	return isNetError && netError.Timeout()
}

// IsRetryableError will check if the error is retryable
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}
	_, isNetError := err.(net.Error)
	return isNetError
}
