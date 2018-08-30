package error

import (
	"errors"
)

var (
	InvalidRequestError = errors.New("client.InvalidRequestError")
	SendRequestError    = errors.New("client.SendRequestError")
	TimeoutError        = errors.New("client.TimeoutError")
)
