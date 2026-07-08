package umodelverse

import (
	"encoding/json"
	"fmt"
	"sort"

	uai "github.com/ucloud/ucloud-sdk-go/services/uai_modelverse"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
	"github.com/ucloud/ucloud-sdk-go/ucloud/response"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

const productName = "umodelverse"

type apiResponse struct {
	response.BaseGenericResponse
}

func newClient(ctx *cli.Context) *uai.UAI_ModelverseClient {
	return cli.NewServiceClient(ctx, uai.NewClient)
}

func newRequest(client *uai.UAI_ModelverseClient, req request.Common, retryable bool) {
	client.Client.SetupRequest(req)
	req.SetRetryable(retryable)
}

func invokeUMAction(client *uai.UAI_ModelverseClient, action string, req request.Common) (*apiResponse, error) {
	var resp apiResponse
	err := client.Client.InvokeAction(action, req, &resp)
	return &resp, err
}

func printResponse(ctx *cli.Context, resp *apiResponse) {
	payload := resp.GetPayload()
	if ctx.Format() != cli.OutputTable {
		ctx.PrintList(payload)
		return
	}

	keys := make([]string, 0, len(payload))
	for key := range payload {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	rows := make([]fieldRow, 0, len(keys))
	for _, key := range keys {
		rows = append(rows, fieldRow{Field: key, Value: renderValue(payload[key])})
	}
	ctx.PrintList(rows)
}

func renderValue(v interface{}) string {
	switch val := v.(type) {
	case nil:
		return ""
	case string:
		return val
	case float64, bool:
		return fmt.Sprintf("%v", val)
	default:
		b, err := json.Marshal(val)
		if err != nil {
			return fmt.Sprintf("%v", val)
		}
		return string(b)
	}
}

type apiKeyRequest struct {
	request.CommonBase

	KeyId                            *string `required:"false"`
	Name                             *string `required:"false"`
	Status                           *int    `required:"false"`
	ModelverseDisabled               *int    `required:"false"`
	SandBoxDisabled                  *int    `required:"false"`
	DailyLimitAmount                 *string `required:"false"`
	MonthlyLimitAmount               *string `required:"false"`
	ExpireTime                       *int64  `required:"false"`
	GrantAllModels                   *bool   `required:"false"`
	GrantedModels                    *string `required:"false"`
	IPWhitelist                      *string `required:"false"`
	DailyQuotaAlertThreshold         *int    `required:"false"`
	MonthlyQuotaAlertThreshold       *int    `required:"false"`
	QuotaAlertChannels               *string `required:"false"`
	QuotaAlertEmail                  *string `required:"false"`
	QuotaAlertPhone                  *string `required:"false"`
	QuotaAlertEmailVerificationToken *string `required:"false"`
	QuotaAlertPhoneVerificationToken *string `required:"false"`
	Offset                           *int    `required:"false"`
	Limit                            *int    `required:"false"`
}

type squareModelRequest struct {
	request.CommonBase

	ModelType   *string `required:"false"`
	KeyWord     *string `required:"false"`
	Offset      *int    `required:"false"`
	Limit       *int    `required:"false"`
	OrderBy     *string `required:"false"`
	Order       *string `required:"false"`
	MaxModelLen *string `required:"false"`
	Language    *string `required:"false"`
}

type requestLogRequest struct {
	request.CommonBase

	StartTime  *int64  `required:"false"`
	EndTime    *int64  `required:"false"`
	Email      *string `required:"false"`
	RequestId  *string `required:"false"`
	ModelNames *string `required:"false"`
	ApiKeyIds  *string `required:"false"`
	Offset     *int    `required:"false"`
	Limit      *int    `required:"false"`
}

type logDetailRequest struct {
	request.CommonBase

	RequestId *string `required:"true"`
}

type orderRequest struct {
	request.CommonBase

	StartTime       *int64   `required:"false"`
	EndTime         *int64   `required:"false"`
	Page            *int     `required:"false"`
	PageSize        *int     `required:"false"`
	ResourceIds     []string `required:"false"`
	ModelIds        []string `required:"false"`
	PricingUnits    []int    `required:"false"`
	PricingSkus     []string `required:"false"`
	OrderTypes      []int    `required:"false"`
	ChargeTypes     []int    `required:"false"`
	OrganizationIds []int    `required:"false"`
	Regions         []string `required:"false"`
	ProductCodes    []string `required:"false"`
}

type filterOptionsRequest struct {
	request.CommonBase

	ProductCode *string `required:"false"`
}
