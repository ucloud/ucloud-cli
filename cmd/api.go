package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/base"
)

//NewCmdAPI ucloud api --xkey xvalue
func NewCmdAPI(out io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "api",
		Short: "Call API",
		Long:  "Call API",
		Run: func(c *cobra.Command, args []string) {
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
