package uphost

import (
	"fmt"

	"github.com/spf13/cobra"

	uphostsdk "github.com/ucloud/ucloud-sdk-go/services/uphost"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// NewCommand builds the `uphost` root command (list-only).
// Mirrors cmd/uphost.go NewCmdUPHost + NewCmdUPHostList.
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uphost",
		Short: "List UPHost instances",
		Long:  `List UPHost instances`,
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newList(ctx))

	return cmd
}

// newList ucloud uphost list
func newList(ctx *cli.Context) *cobra.Command {
	ids := []string{}
	client := cli.NewServiceClient(ctx, uphostsdk.NewClient)
	req := client.NewDescribePHostRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List UPHost instances",
		Long:  "List UPHost instances",
		Run: func(c *cobra.Command, args []string) {
			req.PHostId = ids
			resp, err := client.DescribePHost(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			list := make([]uphostRow, 0)
			for _, ins := range resp.PHostSet {
				row := uphostRow{
					ResourceID: ins.PHostId,
					Name:       ins.Name,
					Config:     fmt.Sprintf("core:%d memory:%dG", ins.CPUSet.CoreCount, ins.Memory/1024),
					Group:      ins.Tag,
					HostType:   ins.PHostType,
					Status:     ins.PMStatus,
					Image:      ins.ImageName,
				}
				for _, ip := range ins.IPSet {
					if ip.OperatorName == "Private" {
						row.PrivateIP = ip.IPAddr
					} else {
						row.PublicIP = ip.IPAddr + " " + ip.OperatorName
					}
				}
				for _, disk := range ins.DiskSet {
					if disk.Name == "data" {
						row.Config += fmt.Sprintf(" data-disk:%dG %s", disk.Space, disk.Type)
					}
				}
				list = append(list, row)
			}
			ctx.PrintList(list)
		},
	}
	flags := cmd.Flags()
	ctx.BindRegion(cmd, req)
	ctx.BindZoneEmpty(cmd, req)
	ctx.BindProjectID(cmd, req)
	ctx.BindOffset(cmd, req)
	ctx.BindLimit(cmd, req)
	flags.StringSliceVar(&ids, "uphost-id", nil, "Optional. Resource ID of uphost instances. List those specified uphost instances")

	return cmd
}
