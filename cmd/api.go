package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spf13/cobra"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/base"
	"github.com/ucloud/ucloud-cli/model/status"
	"github.com/ucloud/ucloud-cli/ux"
)

type RepeatsConfig struct {
	Poller   *base.Poller
	IDInResp string
}

var RepeatsSupportedAPI = map[string]RepeatsConfig{
	"CreateULHostInstance": {Poller: ulhostSpoller, IDInResp: "ULHostId"},
}

const ActionField = "Action"
const RepeatsField = "repeats"
const ConcurrentField = "concurrent"
const DefaultConcurrent = 20
const HelpField = "help"
const HelpInfo = `Usage: ucloud api [options] --Action actionName --param1 value1 --param2 value2 ...
Options:
      --local-file string  the path of the local file which contains the api parameters
      --repeats string     the number of repeats
      --concurrent string  the number of concurrent
      --help               show help`

// NewCmdAPI ucloud api --xkey xvalue
func NewCmdAPI(out io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "api",
		Short: "Call API",
		Long:  "Call API",
		Run: func(c *cobra.Command, args []string) {
			if slices.Contains(args, "--help") {
				fmt.Fprintln(out, HelpInfo)
				return
			}
			params, err := parseParamsFromCmdLine(args)
			if err != nil {
				fmt.Fprintln(out, err)
				return
			}

			if params["local-file"] != nil {
				file, ok := params["local-file"].(string)
				if !ok {
					fmt.Fprintf(out, "local-file should be a string\n")
				}
				params, err = parseParamsFromJSONFile(file)
				if err != nil {
					fmt.Fprintln(out, err)
					return
				}
			}
			if action, actionOK := params[ActionField].(string); actionOK {
				if repeatsConfig, repeatsSupported := RepeatsSupportedAPI[action]; repeatsSupported {
					if repeats, repeatsOK := params[RepeatsField].(string); repeatsOK {
						var repeatsNum int
						var concurrentNum int
						repeatsNum, err = strconv.Atoi(repeats)
						if err != nil {
							fmt.Fprintf(out, "error: %v\n", err)
							return
						}
						if concurrent, concurrentOK := params[ConcurrentField].(string); concurrentOK {
							concurrentNum, err = strconv.Atoi(concurrent)
							if err != nil {
								fmt.Fprintf(out, "error: %v\n", err)
								return
							}
						} else {
							concurrentNum = DefaultConcurrent
						}
						delete(params, RepeatsField)
						delete(params, ConcurrentField)
						err = genericInvokeRepeatWrapper(&repeatsConfig, params, action, repeatsNum, concurrentNum)
						if err != nil {
							fmt.Fprintf(out, "error: %v\n", err)
							return
						}
						return
					}
				}
			}
			req := base.BizClient.UAccountClient.NewGenericRequest()
			err = req.SetPayload(params)
			if err != nil {
				fmt.Fprintf(out, "error: %v\n", err)
				return
			}

			resp, err := base.BizClient.UAccountClient.GenericInvoke(req)
			if err != nil {
				fmt.Fprintf(out, "error: %v\n", err)
				return
			}

			data, err := json.MarshalIndent(resp.GetPayload(), "", "  ")
			if err != nil {
				fmt.Fprintf(out, "error: %v\n", err)
				return
			}
			fmt.Fprintln(out, string(data))
		},
	}
}

func parseParamsFromJSONFile(path string) (map[string]interface{}, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file error: %w", err)
	}
	params := make(map[string]interface{})
	err = json.Unmarshal(content, &params)
	if err != nil {
		return nil, fmt.Errorf("parse json error: %w", err)
	}
	return params, err
}

func parseParamsFromCmdLine(args []string) (map[string]interface{}, error) {
	if len(args)%2 != 0 {
		return nil, errors.New("the key value pairs of api parameters do not match")
	}
	params := make(map[string]interface{})
	for i := 0; i < len(args)-1; i += 2 {
		if strings.HasPrefix(args[i], "--") {
			args[i] = args[i][2:]
		}
		params[args[i]] = args[i+1]
	}
	return params, nil
}

func genericInvokeRepeatWrapper(repeatsConfig *RepeatsConfig, params map[string]interface{}, action string, repeats int, concurrent int) error {
	if repeatsConfig == nil {
		return fmt.Errorf("error: repeatsConfig is nil")
	}
	if repeats <= 0 {
		return fmt.Errorf("error: repeats should be a positive integer")
	}
	if concurrent <= 0 {
		return fmt.Errorf("error: concurrent should be a positive integer")
	}
	wg := &sync.WaitGroup{}
	tokens := make(chan struct{}, concurrent)
	retCh := make(chan bool, repeats)

	wg.Add(repeats)
	//ux.Doc.Disable()
	refresh := ux.NewRefresh()

	req := base.BizClient.UAccountClient.NewGenericRequest()
	err := req.SetPayload(params)
	if err != nil {
		return fmt.Errorf("fail to set payload: %w", err)
	}

	go func(req request.GenericRequest) {
		for i := 0; i < repeats; i++ {
			go func(req request.GenericRequest, idx int) {
				tokens <- struct{}{}
				defer func() {
					<-tokens
					//设置延时，使报错能渲染出来
					time.Sleep(time.Second / 5)
					wg.Done()
				}()
				success := true
				resp, err := base.BizClient.UAccountClient.GenericInvoke(req)
				block := ux.NewBlock()
				ux.Doc.Append(block)
				logs := []string{"=================================================="}
				logs = append(logs, fmt.Sprintf("api:%v, request:%v", action, base.ToQueryMap(req)))
				if err != nil {
					logs = append(logs, fmt.Sprintf("err:%v", err))
					block.Append(base.ParseError(err))
					success = false
				} else {
					logs = append(logs, fmt.Sprintf("resp:%#v", resp))
					resourceId, ok := resp.GetPayload()[repeatsConfig.IDInResp].(string)
					if !ok {
						block.Append(fmt.Sprintf("expect %v in response, but not found", repeatsConfig.IDInResp))
						success = false
					} else {
						text := fmt.Sprintf("the resource[%s] is initializing", resourceId)
						result := repeatsConfig.Poller.Sspoll(resourceId, text, []string{status.HOST_RUNNING, status.HOST_FAIL}, block, &request.CommonBase{
							Region:    ucloud.String(req.GetRegion()),
							Zone:      ucloud.String(req.GetZone()),
							ProjectId: ucloud.String(req.GetProjectId()),
						})
						if result.Err != nil {
							success = false
							block.Append(result.Err.Error())
						}
					}
					retCh <- success
					logs = append(logs, fmt.Sprintf("index:%d, result:%t", idx, success))
					base.LogInfo(logs...)
				}
			}(req, i)
		}
	}(req)

	var success, fail atomic.Int32
	go func() {
		block := ux.NewBlock()
		ux.Doc.Append(block)
		block.Append(fmt.Sprintf("creating, total:%d, success:%d, fail:%d", repeats, success.Load(), fail.Load()))
		blockCount := ux.Doc.GetBlockCount()
		for ret := range retCh {
			if ret {
				success.Add(1)
			} else {
				fail.Add(1)
			}
			text := fmt.Sprintf("creating, total:%d, success:%d, fail:%d", repeats, success.Load(), fail.Load())
			if blockCount != ux.Doc.GetBlockCount() {
				block = ux.NewBlock()
				ux.Doc.Append(block)
				block.Append(text)
				blockCount = ux.Doc.GetBlockCount()
			} else {
				block.Update(text, 0)
			}
			if repeats == int(success.Load())+int(fail.Load()) && fail.Load() > 0 {
				fmt.Printf("Check logs in %s\n", base.GetLogFilePath())
			}
		}
	}()
	wg.Wait()
	refresh.Do(fmt.Sprintf("finally, total:%d, success:%d, fail:%d", repeats, success.Load(), repeats-int(success.Load())))
	return nil
}
