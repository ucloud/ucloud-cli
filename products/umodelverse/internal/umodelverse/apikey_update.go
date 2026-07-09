package umodelverse

import (
	"fmt"

	"github.com/spf13/cobra"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func newAPIKeyUpdate(ctx *cli.Context) *cobra.Command {
	client := newClient(ctx)
	req := &apiKeyRequest{}
	newRequest(client, req, false)
	var grantedModels []string
	var quotaAlertChannels []string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a uModelVerse API key",
		Long:  "Update a uModelVerse API key.",
		Run: func(c *cobra.Command, args []string) {
			flags := c.Flags()
			req.KeyId = sdk.String(ctx.PickResourceID(*req.KeyId))
			clearStringIfUnchanged(flags, "name", &req.Name)
			clearIntIfUnchanged(flags, "status", &req.Status)
			clearIntIfUnchanged(flags, "modelverse-disabled", &req.ModelverseDisabled)
			clearIntIfUnchanged(flags, "sandbox-disabled", &req.SandBoxDisabled)
			clearStringIfUnchanged(flags, "daily-limit-amount", &req.DailyLimitAmount)
			clearStringIfUnchanged(flags, "monthly-limit-amount", &req.MonthlyLimitAmount)
			clearInt64IfUnchanged(flags, "expire-time", &req.ExpireTime)
			clearBoolIfUnchanged(flags, "grant-all-models", &req.GrantAllModels)
			req.GrantedModels = stringSliceJSONRef(grantedModels)
			clearStringIfUnchanged(flags, "ip-whitelist", &req.IPWhitelist)
			clearIntIfUnchanged(flags, "daily-quota-alert-threshold", &req.DailyQuotaAlertThreshold)
			clearIntIfUnchanged(flags, "monthly-quota-alert-threshold", &req.MonthlyQuotaAlertThreshold)
			req.QuotaAlertChannels = stringSliceJSONRef(quotaAlertChannels)
			clearStringIfUnchanged(flags, "quota-alert-email", &req.QuotaAlertEmail)
			clearStringIfUnchanged(flags, "quota-alert-phone", &req.QuotaAlertPhone)
			clearStringIfUnchanged(flags, "quota-alert-email-verification-token", &req.QuotaAlertEmailVerificationToken)
			clearStringIfUnchanged(flags, "quota-alert-phone-verification-token", &req.QuotaAlertPhoneVerificationToken)
			if req.IPWhitelist != nil {
				req.IPWhitelist = sdk.String(cleanMultilineFlag(*req.IPWhitelist))
			}
			resp, err := invokeUMAction(client, "UpdateUMInferAPIKey", req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "umodelverse apikey[%s] updated\n", *req.KeyId)
			printResponse(ctx, resp)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	req.KeyId = flags.String("key-id", "", "Required. API key ID to update.")
	req.Name = flags.String("name", "", "Optional. Updated API key name.")
	req.Status = flags.Int("status", 0, "Optional. API key status: 1 enabled, 2 disabled.")
	req.ModelverseDisabled = flags.Int("modelverse-disabled", 0, "Optional. Whether ModelVerse is disabled: 0 enabled, 1 disabled.")
	req.SandBoxDisabled = flags.Int("sandbox-disabled", 0, "Optional. Whether sandbox is disabled: 0 enabled, 1 disabled.")
	req.DailyLimitAmount = flags.String("daily-limit-amount", "", "Optional. Daily limit amount.")
	req.MonthlyLimitAmount = flags.String("monthly-limit-amount", "", "Optional. Monthly limit amount.")
	req.ExpireTime = flags.Int64("expire-time", 0, "Optional. API key expire time, Unix timestamp. Use -1 for never expire.")
	req.GrantAllModels = flags.Bool("grant-all-models", true, "Optional. Grant access to all models.")
	flags.StringSliceVar(&grantedModels, "granted-models", nil, "Optional. Granted model IDs when --grant-all-models=false. Can be repeated, comma-separated, or a JSON array string.")
	req.IPWhitelist = flags.String("ip-whitelist", "", "Optional. IP whitelist, newline-separated; literal \\n is also accepted.")
	req.DailyQuotaAlertThreshold = flags.Int("daily-quota-alert-threshold", 0, "Optional. Daily quota alert threshold.")
	req.MonthlyQuotaAlertThreshold = flags.Int("monthly-quota-alert-threshold", 0, "Optional. Monthly quota alert threshold.")
	flags.StringSliceVar(&quotaAlertChannels, "quota-alert-channel", nil, "Optional. Quota alert channel, e.g. email or sms. Can be repeated or comma-separated.")
	req.QuotaAlertEmail = flags.String("quota-alert-email", "", "Optional. Email address for quota alerts.")
	req.QuotaAlertPhone = flags.String("quota-alert-phone", "", "Optional. Phone number for quota alerts.")
	req.QuotaAlertEmailVerificationToken = flags.String("quota-alert-email-verification-token", "", "Optional. Email verification token for quota alerts.")
	req.QuotaAlertPhoneVerificationToken = flags.String("quota-alert-phone-verification-token", "", "Optional. Phone verification token for quota alerts.")
	bindProject(cmd, req, ctx.DefaultProjectID())

	cmd.MarkFlagRequired("key-id")
	return cmd
}
