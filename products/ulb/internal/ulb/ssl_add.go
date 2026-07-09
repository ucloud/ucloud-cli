package ulb

import (
	"fmt"

	"github.com/spf13/cobra"

	ulbsdk "github.com/ucloud/ucloud-sdk-go/services/ulb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newSSLAdd returns ucloud ulb ssl add.
func newSSLAdd(ctx *cli.Context) *cobra.Command {
	var allPath, sitePath, keyPath, caPath *string
	client := cli.NewServiceClient(ctx, ulbsdk.NewClient)
	req := client.NewCreateSSLRequest()
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add SSL Certificate",
		Long:  "Add SSL Certificate",
		Run: func(c *cobra.Command, args []string) {
			if *allPath == "" && (*sitePath == "" || *keyPath == "") {
				fmt.Fprintln(ctx.ProgressWriter(), "if all-in-one-file is omitted, site-certificate-file and private-key-file can't be empty")
				return
			}
			if *allPath != "" {
				content, err := readFile(*allPath)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				req.SSLContent = &content
			}
			if *sitePath != "" {
				content, err := readFile(*sitePath)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				req.UserCert = &content
			}
			if *keyPath != "" {
				content, err := readFile(*keyPath)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				req.PrivateKey = &content
			}
			if *caPath != "" {
				content, err := readFile(*caPath)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				req.CaCert = &content
			}

			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			resp, err := client.CreateSSL(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "ssl certificate[%s] added\n", resp.SSLId)
			ctx.EmitResult(cli.OpResultRow{ResourceID: resp.SSLId, Action: "add-ssl", Status: "Added"})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	req.SSLName = flags.String("name", "", "Required. Name of ssl certificate to add")
	req.SSLType = flags.String("format", "Pem", "Optional. Format of ssl certificate")
	allPath = flags.String("all-in-one-file", "", "Optional. Path of file which contain the complete content of the SSL certificate, including the content of site certificate, the private key which encrypted the site certificate, and the CA certificate. ")
	sitePath = flags.String("site-certificate-file", "", "Optional. Path of user's certificate file, *.crt. Required if all-in-one-file is omitted")
	keyPath = flags.String("private-key-file", "", "Optional. Path of private key file, *.key. Required if all-in-one-file is omitted")
	caPath = flags.String("ca-certificate-file", "", "Optional. Path of CA certificate file, *.crt")
	cmd.MarkFlagRequired("name")
	ctx.SetCompletion(cmd, "all-in-one-file", func() []string {
		return common.GetFileList("")
	})
	ctx.SetCompletion(cmd, "private-key-file", func() []string {
		return common.GetFileList(".key")
	})
	ctx.SetCompletion(cmd, "ca-certificate-file", func() []string {
		return common.GetFileList(".crt")
	})
	ctx.SetCompletion(cmd, "site-certificate-file", func() []string {
		return common.GetFileList(".crt")
	})
	return cmd
}
