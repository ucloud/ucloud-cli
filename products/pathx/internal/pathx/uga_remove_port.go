package pathx

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	ppathx "github.com/ucloud/ucloud-sdk-go/private/services/pathx"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newUGARemovePort ucloud pathx uga delete-port
func newUGARemovePort(ctx *cli.Context) *cobra.Command {
	var ports []string
	var protocol string
	client := cli.NewServiceClient(ctx, ppathx.NewClient)
	req := client.NewDeleteUGATaskRequest()
	cmd := &cobra.Command{
		Use:   "delete-port",
		Short: "Delete port for uga instance",
		Long:  "Delete port for uga instance",
		Run: func(c *cobra.Command, args []string) {
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
				fmt.Fprintf(ctx.ProgressWriter(), "protocol should be one of %s, received:%s\n", strings.Join(protocols, ","), protocol)
			}
			*req.UGAId = ctx.PickResourceID(*req.UGAId)
			_, err = client.DeleteUGATask(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "port %v deleted\n", ports)
			ctx.EmitResult(cli.OpResultRow{ResourceID: *req.UGAId, Action: "delete-port", Status: "Deleted"})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	ctx.BindProjectID(cmd, req)
	req.UGAId = flags.String("uga-id", "", "Required. Resource ID of uga instance to delete port")
	flags.StringVar(&protocol, "protocol", "", fmt.Sprintf("Required. accept values: %s", strings.Join(protocols, ",")))
	flags.StringSliceVar(&ports, "port", nil, "Required. Single port or port range, separated by ',', for example 80,3000-3010")
	cmd.MarkFlagRequired("protocol")
	cmd.MarkFlagRequired("uga-id")
	cmd.MarkFlagRequired("port")
	command.SetFlagValues(cmd, "protocol", protocols...)
	ctx.SetCompletion(cmd, "uga-id", func() []string {
		return getUGAIDList(ctx, *req.ProjectId)
	})
	return cmd
}
