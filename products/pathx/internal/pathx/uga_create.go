package pathx

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	ppathx "github.com/ucloud/ucloud-sdk-go/private/services/pathx"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newUGACreate ucloud pathx uga create
func newUGACreate(ctx *cli.Context) *cobra.Command {
	var protocol string
	var ports, lines []string
	client := cli.NewServiceClient(ctx, ppathx.NewClient)
	req := client.NewCreateUGAInstanceRequest()
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create uga instance",
		Long:    "Create uga instance",
		Example: "ucloud pathx uga create --name testcli1 --protocol tcp --origin-location 中国 --origin-domain lixiaojun.xyz --upath-id upath-auvfexxx/test_0 --port 80-90,100,110-115",
		Run: func(c *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			if *req.IPList == "" && *req.Domain == "" {
				fmt.Fprintln(w, "origin-ip and origin-domain can not be both empty")
				return
			}
			portList, err := formatPortList(ports)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			switch strings.ToLower(protocol) {
			case "tcp":
				req.TCP = portList
			case "udp":
				req.UDP = portList
			case "http":
				req.HTTP = portList
			case "https":
				req.HTTPS = portList
			default:
				fmt.Fprintf(w, "protocol should be one of %s, received:%s\n", strings.Join(protocols, ","), protocol)
			}
			resp, err := client.CreateUGAInstance(req)
			if err != nil {
				if uErr, ok := err.(uerr.Error); ok && uErr.Code() == 33756 {
					fmt.Fprintf(w, "The number of ports added exceeds the limit(50). We recommend that you could reduce the number of ports, then create an uga instance, \nand then add the remaining ports by executing 'ucloud pathx uga add-port --protocol %s --uga-id <uga-id> --port <PortList>'\n", protocol)
				}
				return
			}
			fmt.Fprintf(w, "uga[%s] created\n", resp.UGAId)
			results := []cli.OpResultRow{{ResourceID: resp.UGAId, Action: "create", Status: "Created"}}
			for _, path := range lines {
				p := ctx.PickResourceID(path)
				bindReq := client.NewUGABindUPathRequest()
				bindReq.ProjectId = req.ProjectId
				bindReq.UGAId = sdk.String(resp.UGAId)
				bindReq.UPathId = &p
				_, err := client.UGABindUPath(bindReq)
				if err != nil {
					fmt.Fprintf(w, "bind uga[%s] and upath[%s] failed: %v\n", resp.UGAId, p, err)
				} else {
					fmt.Fprintf(w, "bound uga[%s] and upath[%s]\n", resp.UGAId, p)
					results = append(results, cli.OpResultRow{ResourceID: p, Action: "bind-upath", Status: "Bound"})
				}
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	ctx.BindProjectID(cmd, req)
	req.Name = flags.String("name", "", "Required. Name of uga instance to create")
	req.IPList = flags.String("origin-ip", "", "Required if origin-domain is empty. IP address of origin. multiple IP address separated by ','")
	req.Domain = flags.String("origin-domain", "", "Required if origin-ip is empty.")
	req.Location = flags.String("origin-location", "", "Required. Location of origin ip or domain. accpet valeus:'中国','洛杉矶','法兰克福','中国香港','雅加达','孟买','东京','莫斯科','新加坡','曼谷','中国台北','华盛顿','首尔'")
	flags.StringVar(&protocol, "protocol", "", fmt.Sprintf("Required. accept values: %s", strings.Join(protocols, ",")))
	flags.StringSliceVar(&ports, "port", nil, "Required. Single port or port range, separated by ',', for example 80,3000-3010")
	flags.StringSliceVar(&lines, "upath-id", nil, "Required. Accelerated path to bind with the uga instance to create. multiple upath-id separated by ','; see 'ucloud pathx upath list")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("origin-location")
	cmd.MarkFlagRequired("protocol")
	cmd.MarkFlagRequired("port")
	cmd.MarkFlagRequired("upath-id")
	command.SetFlagValues(cmd, "origin-location", "中国", "洛杉矶", "法兰克福", "中国香港", "雅加达", "孟买", "东京", "莫斯科", "新加坡", "曼谷", "中国台北", "华盛顿", "首尔")
	command.SetFlagValues(cmd, "protocol", protocols...)
	ctx.SetCompletion(cmd, "upath-id", func() []string {
		return getUpathIDList(ctx, *req.ProjectId)
	})
	return cmd
}
