package cmd

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

func NewCmdSignature() *cobra.Command {
	var (
		rawParams  []string
		privateKey string
		rawURL     string
	)
	cmd := &cobra.Command{
		Use:   "signature",
		Short: "Calculate ucloud signature",
		Long:  "Calculate ucloud signature",

		Aliases: []string{"sign"},

		Run: func(cmd *cobra.Command, args []string) {
			var params map[string]interface{}
			if rawURL != "" {
				// Parse params from exists url
				parsedURL, err := url.Parse(rawURL)
				if err != nil {
					fmt.Printf("error: failed to parse url %q: %v\n", rawURL, err)
					return
				}
				query := parsedURL.Query()
				params = make(map[string]interface{}, len(query))
				for key, values := range query {
					if key == "Signature" {
						fmt.Println("error: the `Signature` cannot be placed in url")
						return
					}
					if len(values) == 0 {
						continue
					}
					val := values[0]
					params[key] = val
				}
			}
			if len(rawParams) > 0 {
				if params == nil {
					params = make(map[string]interface{}, len(rawParams))
				}
				for _, rawParam := range rawParams {
					kv := strings.Split(rawParam, "=")
					if len(kv) != 2 {
						fmt.Printf("error: param %q is invalid\n", rawParam)
						return
					}
					params[kv[0]] = kv[1]
				}
			}
			if len(params) == 0 {
				fmt.Println("error: missing param")
				return
			}

			r := auth.CalculateSignature(params, privateKey)

			var colorParamBuf bytes.Buffer
			for _, key := range r.SortedKeys {
				val := params[key]
				colorParamBuf.WriteString(color.GreenString(key))
				colorParamBuf.WriteString(color.CyanString("%v", val))
			}
			colorParamBuf.WriteString(color.MagentaString(privateKey))
			fmt.Println("")
			fmt.Printf("ParamStr: %s\n", colorParamBuf.String())
			fmt.Println("")

			fmt.Printf("Signature: %s\n", color.BlueString(r.Sign))
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringArrayVarP(&rawParams, "param", "m", nil, "Request params")
	flags.StringVarP(&privateKey, "private-key", "k", "", "Private key")
	flags.StringVarP(&rawURL, "url", "u", "", "Request url without signature")
	cmd.MarkFlagRequired("private-key")

	return cmd
}
