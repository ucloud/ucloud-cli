package waiter

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
)

var (
	errTimeoutConf = errors.New("timeout cannot be set zero")
)

// TimeoutError is returned when WaitForState times out
type TimeoutError struct {
	LastError      error
	LastState      string
	Timeout        time.Duration
	ExpectedStates []string
}

func (e *TimeoutError) Error() string {
	errs := []string{"cannot waiting for resource is completed"}

	if e.Timeout > 0 {
		errs = append(errs, fmt.Sprintf("timeout: %s", e.Timeout))
	}

	if e.LastState != "" {
		errs = append(errs, fmt.Sprintf("last state: %q", e.LastState))
	}

	if len(e.ExpectedStates) > 0 {
		errs = append(errs, fmt.Sprintf("want: %q", strings.Join(e.ExpectedStates, ",")))
	}

	if e.LastError != nil {
		errs = append(errs, fmt.Sprintf("err: %s", e.LastError))
	}

	return strings.Join(errs, ", ")
}
