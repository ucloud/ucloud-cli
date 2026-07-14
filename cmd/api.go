package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/ucloud/ucloud-sdk-go/services/uaccount"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/cmd/internal/platform"
	"github.com/ucloud/ucloud-cli/model/status"
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/ui"
)

type RepeatsConfig struct {
	Poller   cli.Poller
	IDInResp string
}

type repeatResult struct {
	success bool
	err     error
}

func repeatsSupportedAPI(out io.Writer) map[string]RepeatsConfig {
	return map[string]RepeatsConfig{
		"CreateULHostInstance": {Poller: newULHostPoller(out), IDInResp: "ULHostId"},
	}
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
		RunE: func(c *cobra.Command, args []string) error {
			if containHelp(args) {
				fmt.Fprintln(out, HelpInfo)
				return nil
			}
			params, err := parseParamsFromCmdLine(args)
			if err != nil {
				fmt.Fprintln(out, err)
				return err
			}

			if params["local-file"] != nil {
				file, ok := params["local-file"].(string)
				if !ok {
					err := fmt.Errorf("local-file should be a string")
					fmt.Fprintln(out, err)
					return err
				}
				params, err = parseParamsFromJSONFile(file)
				if err != nil {
					fmt.Fprintln(out, err)
					return err
				}
			}
			if action, actionOK := params[ActionField].(string); actionOK {
				if repeatsConfig, repeatsSupported := repeatsSupportedAPI(out)[action]; repeatsSupported {
					if repeats, repeatsOK := params[RepeatsField].(string); repeatsOK {
						var repeatsNum int
						var concurrentNum int
						repeatsNum, err = strconv.Atoi(repeats)
						if err != nil {
							fmt.Fprintf(out, "error: %v\n", err)
							return err
						}
						if concurrent, concurrentOK := params[ConcurrentField].(string); concurrentOK {
							concurrentNum, err = strconv.Atoi(concurrent)
							if err != nil {
								fmt.Fprintf(out, "error: %v\n", err)
								return err
							}
						} else {
							concurrentNum = DefaultConcurrent
						}
						delete(params, RepeatsField)
						delete(params, ConcurrentField)
						err = genericInvokeRepeatWrapper(&repeatsConfig, params, action, repeatsNum, concurrentNum, out)
						if err != nil {
							fmt.Fprintf(out, "error: %v\n", err)
							return err
						}
						return nil
					}
				}
			}
			client := newServiceClient(uaccount.NewClient)
			req := client.NewGenericRequest()
			err = req.SetPayload(params)
			if err != nil {
				fmt.Fprintf(out, "error: %v\n", err)
				return err
			}

			resp, err := client.GenericInvoke(req)
			if err != nil {
				fmt.Fprintf(out, "error: %v\n", err)
				return err
			}

			data, err := json.MarshalIndent(resp.GetPayload(), "", "  ")
			if err != nil {
				fmt.Fprintf(out, "error: %v\n", err)
				return err
			}
			fmt.Fprintln(out, string(data))
			return nil
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

func genericInvokeRepeatWrapper(repeatsConfig *RepeatsConfig, params map[string]interface{}, action string, repeats int, concurrent int, out io.Writer) error {
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
	retCh := make(chan repeatResult, repeats)

	wg.Add(repeats)
	doc := ui.NewDocument(out)
	refresh := ui.NewRefresh(out)
	printBlockLine := func(block *ui.Block, line string) {
		block.Append(line)
		if !ui.IsTTY(out) {
			fmt.Fprintln(out, line)
		}
	}

	client := newServiceClient(uaccount.NewClient)
	req := client.NewGenericRequest()
	err := req.SetPayload(params)
	if err != nil {
		return fmt.Errorf("fail to set payload: %w", err)
	}

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
			var resultErr error
			resp, err := client.GenericInvoke(req)
			block := ui.NewBlock()
			doc.Append(block)
			logs := []string{"=================================================="}
			logs = append(logs, fmt.Sprintf("api:%v, request:%v", action, platform.ToQueryMap(req)))
			if err != nil {
				logs = append(logs, fmt.Sprintf("err:%v", err))
				printBlockLine(block, platform.ParseError(err))
				success = false
				resultErr = err
			} else {
				logs = append(logs, fmt.Sprintf("resp:%#v", resp))
				resourceId, ok := resp.GetPayload()[repeatsConfig.IDInResp].(string)
				if !ok {
					resultErr = fmt.Errorf("expect %v in response, but not found", repeatsConfig.IDInResp)
					printBlockLine(block, resultErr.Error())
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
						resultErr = result.Err
						printBlockLine(block, result.Err.Error())
					}
				}
			}
			retCh <- repeatResult{success: success, err: resultErr}
			logs = append(logs, fmt.Sprintf("index:%d, result:%t", idx, success))
			platform.LogInfo(logs...)
		}(req, i)
	}

	var success, fail int
	var firstErr error
	block := ui.NewBlock()
	doc.Append(block)
	block.Append(fmt.Sprintf("creating, total:%d, success:%d, fail:%d", repeats, success, fail))
	blockCount := doc.GetBlockCount()
	for i := 0; i < repeats; i++ {
		ret := <-retCh
		if ret.success {
			success++
		} else {
			fail++
			if firstErr == nil {
				firstErr = ret.err
			}
		}
		text := fmt.Sprintf("creating, total:%d, success:%d, fail:%d", repeats, success, fail)
		if blockCount != doc.GetBlockCount() {
			block = ui.NewBlock()
			doc.Append(block)
			block.Append(text)
			blockCount = doc.GetBlockCount()
		} else {
			block.Update(text, 0)
		}
	}
	wg.Wait()
	if fail > 0 {
		fmt.Fprintf(out, "Check logs in %s\n", platform.GetLogFilePath())
	}
	refresh.Do(fmt.Sprintf("finally, total:%d, success:%d, fail:%d", repeats, success, fail))
	if firstErr != nil {
		return fmt.Errorf("repeat API %s failed: %w", action, firstErr)
	}
	return nil
}

func containHelp(args []string) bool {
	for _, arg := range args {
		if arg == "--help" {
			return true
		}
	}
	return false
}
