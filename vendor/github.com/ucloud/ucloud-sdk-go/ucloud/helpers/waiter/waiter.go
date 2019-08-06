/*
Package waiter is a helper package use for waiting remote state is transformed into target state.
*/
package waiter

import (
	"time"

	"github.com/ucloud/ucloud-sdk-go/ucloud/log"
)

const graceRefreshTimeout = 30 * time.Second

// RefreshFunc is the function to query remote resource state
type RefreshFunc func() (result interface{}, state string, err error)

// StateWaiter is the waiter that waiting for remote state achieve to target state
type StateWaiter struct {
	Pending      []string
	Target       []string
	Refresh      RefreshFunc
	Delay        time.Duration
	Timeout      time.Duration
	MinTimeout   time.Duration
	PollInterval time.Duration
}

// Wait watches an object and waits for it to achieve the state
func (waiter *StateWaiter) Wait() (interface{}, error) {
	if waiter.Timeout == 0 {
		return nil, errTimeoutConf
	}

	// it is a re-implementation of terraform StateChangeConf
	log.Debugf("waiting for state: %s", waiter.Target)

	resCh := make(chan refreshResult, 1)
	cancelCh := make(chan struct{})

	result := refreshResult{}

	go func() {
		defer close(resCh)

		// delay before refresh
		time.Sleep(waiter.Delay)

		var wait time.Duration

		for {
			// send an empty result to active channel
			resCh <- result

			select {
			case <-cancelCh:
				return
			case <-time.After(wait):
				// first round should not wait
				// initial wait with 100ms
				if wait == 0 {
					wait = 100 * time.Millisecond
				}
			}

			// refresh newest state
			res, currentState, err := waiter.Refresh()
			result = refreshResult{
				Result: res,
				State:  currentState,
				Error:  err,
			}

			if err != nil {
				resCh <- result
				return
			}

			// if target state is achieved, done
			for _, allowed := range waiter.Target {
				if currentState == allowed {
					result.Done = true
					resCh <- result
					return
				}
			}

			isPending := false
			for _, allowed := range waiter.Pending {
				if currentState == allowed {
					isPending = true
					break
				}
			}

			if !isPending && len(waiter.Pending) > 0 {
				resCh <- result
				return
			}

			// wait interval using exponential backoff, policy as follow:
			//
			// * 0 100ms 200ms 400ms 800ms 1.6s 3.2s 6.4s 10s 10s ... (0 <= MinTimeout <= 100ms)
			// * 0 3s 6s 10s 10s ... (100ms < MinTimeout <= 10s, eg. 3s)
			// * 0 11s 10s 10s 10s ... (10s < MinTimeout, eg. 11s)
			wait *= 2

			if waiter.PollInterval > 0 && waiter.PollInterval < 180*time.Second {
				wait = waiter.PollInterval
			} else {
				if wait < waiter.MinTimeout {
					wait = waiter.MinTimeout
				} else if wait > 10*time.Second {
					wait = 10 * time.Second
				}
			}
		}
	}()

	// store the last value result from the refresh loop
	var lastRes refreshResult

	timeout := time.After(waiter.Timeout)
	for {
		select {
		case r, ok := <-resCh:
			if !ok {
				return lastRes.Result, lastRes.Error
			}

			if r.Done {
				return r.Result, r.Error
			}

			lastRes = r
		case <-timeout:
			log.Debugf("wait for state timeout after %s", waiter.Timeout)

			// cancel the goroutine and start our grace period timer
			close(cancelCh)

			return waiter.waitForTimeout(resCh, lastRes)
		}
	}
}

type refreshResult struct {
	Result interface{}
	State  string
	Error  error
	Done   bool
}

func (waiter *StateWaiter) waitForTimeout(resCh <-chan refreshResult, last refreshResult) (interface{}, error) {
	timeout := time.After(graceRefreshTimeout)

	errTimeout := &TimeoutError{
		LastError:      last.Error,
		LastState:      last.State,
		Timeout:        waiter.Timeout,
		ExpectedStates: waiter.Target,
	}

	// we need wait until at lease once refresh is completed,
	// and close the cancel channel
	select {
	case r, ok := <-resCh:
		if r.Done {
			return r.Result, r.Error
		}

		if !ok {
			return nil, errTimeout
		}
	case <-timeout:
		return nil, errTimeout
	}

	return nil, nil
}
