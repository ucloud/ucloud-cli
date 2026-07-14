package cli

import (
	"fmt"
	"io"
	"reflect"
	"time"

	"github.com/ucloud/ucloud-sdk-go/ucloud/helpers/waiter"
	"github.com/ucloud/ucloud-sdk-go/ucloud/log"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/ui"
)

type PollResult struct {
	Done    bool
	Timeout bool
	Err     error
}

type Poller interface {
	Spoll(resourceID, pollText string, targetStates []string)
	Sspoll(resourceID, pollText string, targetStates []string, block *Block, common *request.CommonBase) *PollResult
}

type poller struct {
	describe    func(string, *request.CommonBase) (interface{}, error)
	out         io.Writer
	stateFields []string
	timeout     time.Duration
}

// defaultPollTimeout 是一次同步轮询的总等待预算。默认 10 分钟；可经
// SetDefaultPollTimeout 覆盖（cmd 层接到 --wait-timeout-sec）。
var defaultPollTimeout = 10 * time.Minute

// SetDefaultPollTimeout 覆盖此后由 NewPoller 创建的 poller 的总超时。
// 非正值被忽略：SDK 的 waiter 在 Timeout==0 时直接报错（errTimeoutConf），
// 故此时保留内置默认，不将其置 0。
func SetDefaultPollTimeout(d time.Duration) {
	if d > 0 {
		defaultPollTimeout = d
	}
}

func NewPoller(describe func(string, *request.CommonBase) (interface{}, error), out io.Writer) Poller {
	return &poller{
		describe:    describe,
		out:         out,
		stateFields: []string{"State", "Status"},
		timeout:     defaultPollTimeout,
	}
}

func (p *poller) Spoll(resourceID, pollText string, targetStates []string) {
	done := make(chan bool)
	go func() {
		if _, err := p.wait(resourceID, targetStates, nil); err != nil {
			log.Error(err)
			if _, ok := err.(*waiter.TimeoutError); ok {
				done <- false
				return
			}
		}
		done <- true
	}()

	if !ui.IsTTY(p.out) {
		if <-done {
			fmt.Fprintf(p.out, "%s...done\n", pollText)
		} else {
			fmt.Fprintf(p.out, "%s...timeout\n", pollText)
		}
		return
	}
	spinner := ui.NewDotSpinner(p.out)
	spinner.Start(pollText)
	ret := <-done
	if ret {
		spinner.Stop()
	} else {
		spinner.Timeout()
	}
}

func (p *poller) Sspoll(resourceID, pollText string, targetStates []string, block *Block, common *request.CommonBase) *PollResult {
	pollRetChan := make(chan PollResult)
	go func() {
		ret := PollResult{Done: true}
		if _, err := p.wait(resourceID, targetStates, common); err != nil {
			ret.Done = false
			ret.Err = err
			if _, ok := err.(*waiter.TimeoutError); ok {
				ret.Timeout = true
			}
		}
		pollRetChan <- ret
	}()

	if !ui.IsTTY(p.out) {
		ret := <-pollRetChan
		if ret.Timeout {
			fmt.Fprintf(p.out, "%s...timeout\n", pollText)
		} else {
			fmt.Fprintf(p.out, "%s...done\n", pollText)
		}
		return &ret
	}

	spin := ui.NewDotSpin(p.out, pollText)
	if block != nil {
		_ = block.SetSpin(spin)
	}
	ret := <-pollRetChan
	if ret.Timeout {
		spin.Timeout()
	} else {
		spin.Stop()
	}
	return &ret
}

func (p *poller) wait(resourceID string, targetStates []string, common *request.CommonBase) (interface{}, error) {
	w := waiter.StateWaiter{
		Pending: []string{"pending"},
		Target:  []string{"avaliable"},
		Refresh: func() (interface{}, string, error) {
			inst, err := p.describe(resourceID, common)
			if err != nil {
				return nil, "", err
			}
			if inst == nil {
				return nil, "pending", nil
			}
			state, err := p.state(inst)
			if err != nil {
				return nil, "", err
			}
			for _, target := range targetStates {
				if target == state {
					return inst, "avaliable", nil
				}
			}
			return nil, "pending", nil
		},
		Timeout: p.timeout,
	}
	return w.Wait()
}

func (p *poller) state(inst interface{}) (string, error) {
	instValue := reflect.Indirect(reflect.ValueOf(inst))
	if instValue.Kind() != reflect.Struct {
		return "", fmt.Errorf("Instance is not struct")
	}
	instType := instValue.Type()
	for i := 0; i < instValue.NumField(); i++ {
		for _, sf := range p.stateFields {
			if instType.Field(i).Name == sf {
				return instValue.Field(i).String(), nil
			}
		}
	}
	return "", nil
}
