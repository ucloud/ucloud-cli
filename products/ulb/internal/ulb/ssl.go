package ulb

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newSSL returns ucloud ulb ssl.
func newSSL(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ssl",
		Short: "List and manipulate SSL Certificates for ULB",
		Long:  "List and manipulate SSL Certificates for ULB",
	}
	cmd.AddCommand(newSSLList(ctx))
	cmd.AddCommand(newSSLDescribe(ctx))
	cmd.AddCommand(newSSLAdd(ctx))
	cmd.AddCommand(newSSLDelete(ctx))
	cmd.AddCommand(newSSLBind(ctx))
	cmd.AddCommand(newSSLUnbind(ctx))
	return cmd
}
