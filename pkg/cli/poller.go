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
	describe       func(string, *request.CommonBase) (interface{}, error)
	out            io.Writer
	stateFields    []string
	commandTimeout time.Duration
	timeout        time.Duration
}

// builtinPollTimeout 是同步轮询的内置兜底总超时。
const builtinPollTimeout = 10 * time.Minute

// userPollTimeout 是用户经 --wait-timeout-sec 指定的轮询超时（cmd 层启动时注入）。
// 0 表示用户未指定。它是全局最高优先级，覆盖任何命令自设的默认。
var userPollTimeout time.Duration

// SetUserPollTimeout 设置用户层轮询超时（cmd 层接 --wait-timeout-sec）。
// 非正值被忽略：SDK 的 waiter 在 Timeout==0 时直接报错（errTimeoutConf），
// 故此时保留其余层级，不将其置 0。
func SetUserPollTimeout(d time.Duration) {
	if d > 0 {
		userPollTimeout = d
	}
}

// effectivePollTimeout 按 用户 > 命令自设 > 内置 的优先级裁决最终超时。
func effectivePollTimeout(commandTimeout time.Duration) time.Duration {
	if userPollTimeout > 0 {
		return userPollTimeout
	}
	if commandTimeout > 0 {
		return commandTimeout
	}
	return builtinPollTimeout
}

// PollerOption 定制单个 poller 的创建。
type PollerOption func(*poller)

// WithTimeout 让单个命令声明自己的默认轮询超时。非正值被忽略。
// 优先级低于用户 --wait-timeout-sec、高于内置默认（见 effectivePollTimeout）。
func WithTimeout(d time.Duration) PollerOption {
	return func(p *poller) {
		if d > 0 {
			p.commandTimeout = d
		}
	}
}

func NewPoller(describe func(string, *request.CommonBase) (interface{}, error), out io.Writer, opts ...PollerOption) Poller {
	p := &poller{
		describe:    describe,
		out:         out,
		stateFields: []string{"State", "Status"},
	}
	for _, o := range opts {
		o(p)
	}
	p.timeout = effectivePollTimeout(p.commandTimeout)
	return p
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
