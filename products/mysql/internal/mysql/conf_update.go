package mysql

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/udb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newUDBConfUpdate ucloud udb conf update
func newUDBConfUpdate(ctx *cli.Context) *cobra.Command {
	var confID, key, value, file string
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewUpdateUDBParamGroupRequest()
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update parameters of DB's configuration",
		Long:  "Update parameters of DB's configuration",
		Run: func(c *cobra.Command, args []string) {
			id, err := strconv.Atoi(ctx.PickResourceID(confID))
			if err != nil {
				ctx.HandleError(err)
				return
			}
			req.GroupId = &id

			w := ctx.ProgressWriter()
			updated := 0
			if key != "" && value != "" {
				req.Key = &key
				req.Value = &value
				_, err := client.UpdateUDBParamGroup(req)
				if err != nil {
					ctx.HandleError(err)
				} else {
					fmt.Fprintf(w, "conf[%s]'sparameter[%s = %s] updated\n", confID, key, value)
					updated++
				}
			}
			if file != "" {
				params, err := parseParam(file)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				for _, p := range params {
					req.Key = sdk.String(p.Key)
					req.Value = sdk.String(p.Value)
					_, err := client.UpdateUDBParamGroup(req)
					if err != nil {
						ctx.HandleError(fmt.Errorf("conf[%s]'s parameter[%s = %s] failed: %w", confID, p.Key, p.Value, err))
					} else {
						fmt.Fprintf(w, "conf[%s]'sparameter[%s = %s] updated\n", confID, p.Key, p.Value)
						updated++
					}
					fmt.Fprintln(w)
				}
			}
			results := []cli.OpResultRow{}
			if updated > 0 {
				results = append(results, cli.OpResultRow{ResourceID: strconv.Itoa(id), Action: "update", Status: "Updated"})
			}
			ctx.EmitResult(results...)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)

	flags.StringVar(&confID, "conf-id", "", "Required. ConfID of configuration to update")
	flags.StringVar(&key, "key", "", "Optional. Key of parameter")
	flags.StringVar(&value, "value", "", "Optional. Value of parameter")
	flags.StringVar(&file, "file", "", "Optional. Path of file in which each parameter occupies one line with format 'key = value'")

	command.SetCompletion(cmd, "conf-id", func() []string {
		return getModifiableConfIDList(ctx, "", *req.ProjectId, *req.Region, *req.Zone)
	})
	command.SetCompletion(cmd, "file", func() []string {
		return common.GetFileList("")
	})

	cmd.MarkFlagRequired("conf-id")
	return cmd
}

type confParam struct {
	Key   string
	Value string
}

func parseParam(filePath string) ([]confParam, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	params := []confParam{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		strs := strings.SplitN(line, "=", 2)
		if len(strs) < 2 {
			continue
		}
		param := confParam{
			Key:   strings.TrimSpace(strs[0]),
			Value: strings.TrimSpace(strs[1]),
		}
		params = append(params, param)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return params, nil
}
