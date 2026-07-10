package platform

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"unicode"

	"github.com/ucloud/ucloud-sdk-go/services/uaccount"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"
	"github.com/ucloud/ucloud-sdk-go/ucloud/response"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/model"
	"github.com/ucloud/ucloud-cli/pkg/ui"
)

// ConfigPath 配置文件路径
const ConfigPath = ".ucloud"

// GAP 表格列直接的间隔字符数
const GAP = 2

// Cxt 上下文
var Cxt = model.GetContext(os.Stdout)

// SdkClient 用于上报数据
var SdkClient *sdk.Client

// MosaicString 对字符串敏感部分打马赛克 如公钥私钥
func MosaicString(str string, beginChars, lastChars int) string {
	r := len(str) - lastChars - beginChars
	if r > 5 {
		return str[:beginChars] + strings.Repeat("*", 5) + str[(r+beginChars):]
	}
	return strings.Repeat("*", len(str))
}

// GetConfigDir 获取配置文件所在目录
func GetConfigDir() string {
	path := common.GetHomePath() + "/" + ConfigPath
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, 0755)
		if err != nil {
			panic(err)
		}
	}
	return path
}

// HandleBizError 处理RetCode != 0 的业务异常
func HandleBizError(resp response.Common) error {
	format := "Something wrong. RetCode:%d. Message:%s\n"
	LogError(fmt.Sprintf(format, resp.GetRetCode(), resp.GetMessage()))
	return fmt.Errorf(format, resp.GetRetCode(), resp.GetMessage())
}

// HandleError 处理错误，业务错误 和 HTTP错误. The console copy goes to the global
// writer (stdout); product code uses ctx.HandleError → HandleErrorTo(stderr).
func HandleError(err error) { HandleErrorTo(out, err) }

// HandleErrorTo is HandleError with a caller-chosen console writer w, so product
// commands can route errors to stderr and keep stdout machine-clean.
func HandleErrorTo(w io.Writer, err error) {
	if uErr, ok := err.(uerr.Error); ok && uErr.Code() != 0 {
		format := "Something wrong. RetCode:%d. Message:%s\n"
		LogErrorTo(w, fmt.Sprintf(format, uErr.Code(), uErr.Message()))
	} else {
		LogErrorTo(w, fmt.Sprintf("%v", err))
	}
}

// ParseError 解析错误为字符串
func ParseError(err error) string {
	if uErr, ok := err.(uerr.Error); ok && uErr.Code() != 0 {
		format := "Something wrong. RetCode:%d. Message:%s"
		message := uErr.Message()
		if uErr.Code() == -1 || uErr.Code() == -2 {
			message = "request timeout, retry later please"
		}
		return fmt.Sprintf(format, uErr.Code(), message)
	}
	return fmt.Sprintf("Error:%v", err)
}

// PrintJSON 以JSON格式打印数据集合
func PrintJSON(dataSet interface{}, out io.Writer) error {
	bytes, err := json.MarshalIndent(dataSet, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(out, string(bytes))
	if err != nil {
		return err
	}
	return nil
}

// PrintTableS 简化版表格打印，无需传表头，根据结构体反射解析
func PrintTableS(dataSet interface{}) {
	dataSetVal := reflect.ValueOf(dataSet)
	fieldNameList := make([]string, 0)
	if dataSetVal.Len() > 0 {
		elemType := dataSetVal.Index(0).Type()
		for i := 0; i < elemType.NumField(); i++ {
			fieldNameList = append(fieldNameList, elemType.Field(i).Name)
		}
	}
	if kind := dataSetVal.Kind(); kind == reflect.Slice || kind == reflect.Array {
		displaySlice(dataSetVal, fieldNameList)
	} else {
		panic(fmt.Sprintf("Internal error, PrintTableS expect array or slice, accept %T", dataSet))
	}
}

// PrintList 打印表格或者JSON
func PrintList(dataSet interface{}, out io.Writer) {
	if Global.JSON {
		PrintJSON(dataSet, out)
	} else {
		PrintTableS(dataSet)
	}
}

// PrintDescribe 打印详情
func PrintDescribe(attrs []DescribeTableRow, json bool) {
	if json {
		PrintJSON(attrs, os.Stdout)
	} else {
		for _, attr := range attrs {
			fmt.Println(attr.Attribute)
			fmt.Println(attr.Content)
			fmt.Println()
		}
	}
}

// PrintTable 以表格方式打印数据集合
func PrintTable(dataSet interface{}, fieldList []string) {
	dataSetVal := reflect.ValueOf(dataSet)
	switch dataSetVal.Kind() {
	case reflect.Slice, reflect.Array:
		displaySlice(dataSetVal, fieldList)
	default:
		panic(fmt.Sprintf("PrintTable expect array,slice or map, accept %T", dataSet))
	}
}

func displaySlice(listVal reflect.Value, fieldList []string) {
	showFieldMap := make(map[string]int)
	for _, field := range fieldList {
		showFieldMap[field] = len([]rune(field))
	}
	rowList := make([]map[string]interface{}, 0)
	for i := 0; i < listVal.Len(); i++ {
		elemVal := listVal.Index(i)
		elemType := elemVal.Type()
		var rows []map[string]interface{}
		for j := 0; j < elemVal.NumField(); j++ {
			field := elemVal.Field(j)
			fieldName := elemType.Field(j).Name
			if _, ok := showFieldMap[fieldName]; ok {
				if field.Kind() == reflect.Ptr {
					field = field.Elem()
				}
				text := fmt.Sprintf("%v", field.Interface())
				cells := strings.Split(text, "\n")
				for i, cell := range cells {
					width := calcWidth(cell)
					if showFieldMap[fieldName] < width {
						showFieldMap[fieldName] = width
					}
					if len(rows) == i {
						rows = append(rows, make(map[string]interface{}))
					}
					rows[i][fieldName] = cell
				}
			}
		}
		rowList = append(rowList, rows...)
	}
	printTable(rowList, fieldList, showFieldMap)
}

func printTable(rowList []map[string]interface{}, fieldList []string, fieldWidthMap map[string]int) {
	//打印表头
	for _, field := range fieldList {
		tmpl := "%-" + strconv.Itoa(fieldWidthMap[field]+GAP) + "s"
		fmt.Printf(tmpl, field)
	}
	if len(fieldList) != 0 {
		fmt.Printf("\n")
	}

	//打印数据
	for _, row := range rowList {
		for _, field := range fieldList {
			cutWidth := calcCutWidth(fmt.Sprintf("%v", row[field]))
			tmpl := "%-" + strconv.Itoa(fieldWidthMap[field]-cutWidth+GAP) + "v"
			if row[field] != nil {
				fmt.Printf(tmpl, row[field])
			} else {
				fmt.Printf(tmpl, "")
			}
		}
		fmt.Printf("\n")
	}
}

// DescribeTableRow 详情表格通用表格行
type DescribeTableRow struct {
	Attribute string
	Content   string
}

func calcCutWidth(text string) int {
	set := []*unicode.RangeTable{unicode.Han, unicode.Punct}
	width := 0
	for _, r := range text {
		if unicode.IsOneOf(set, r) && r > unicode.MaxLatin1 {
			width++
		}
	}
	return width
}

func calcWidth(text string) int {
	set := []*unicode.RangeTable{unicode.Han, unicode.Punct}
	width := 0
	for _, r := range text {
		if unicode.IsOneOf(set, r) && r > unicode.MaxLatin1 {
			width += 2
		} else {
			width++
		}
	}
	return width
}

// RegionLabel regionlable
var RegionLabel = map[string]string{
	"cn-bj1":       "Beijing1",
	"cn-bj2":       "Beijing2",
	"cn-sh2":       "Shanghai2",
	"cn-gd":        "Guangzhou",
	"cn-qz":        "Quanzhou",
	"hk":           "Hongkong",
	"us-ca":        "LosAngeles",
	"us-ws":        "Washington",
	"ge-fra":       "Frankfurt",
	"th-bkk":       "Bangkok",
	"kr-seoul":     "Seoul",
	"sg":           "Singapore",
	"tw-kh":        "Kaohsiung",
	"rus-mosc":     "Moscow",
	"jpn-tky":      "Tokyo",
	"tw-tp":        "TaiPei",
	"uae-dubai":    "Dubai",
	"idn-jakarta":  "Jakarta",
	"ind-mumbai":   "Bombay",
	"bra-saopaulo": "SaoPaulo",
	"uk-london":    "London",
	"afr-nigeria":  "Lagos",
}

// PickResourceID  uhost-xxx/uhost-name => uhost-xxx
func PickResourceID(str string) string {
	if strings.Index(str, "/") > -1 {
		return strings.SplitN(str, "/", 2)[0]
	}
	return str
}

// WriteJSONFileAtomic 原子写 json 文件：同目录临时文件 + Sync + Rename（D3；对照 botocore#3213 损坏事故）
func WriteJSONFileAtomic(list interface{}, filePath string) error {
	byts, err := json.Marshal(list)
	if err != nil {
		return err
	}
	dir := filepath.Dir(filePath)
	tmp, err := ioutil.TempFile(dir, "."+filepath.Base(filePath)+".tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name()) // rename 成功后此句为 no-op
	if err := tmp.Chmod(LocalFileMode); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(byts); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), filePath)
}

// Confirm 二次确认
func Confirm(yes bool, text string) bool {
	if yes {
		return true
	}
	sure, err := ui.Prompt(text)
	if err != nil {
		LogError(err.Error())
		return false
	}
	return sure
}

func curGoroutineID() int64 {
	var (
		buf [64]byte
		n   = runtime.Stack(buf[:], false)
		stk = strings.TrimPrefix(string(buf[:n]), "goroutine ")
	)

	idField := strings.Fields(stk)[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Errorf("can not get goroutine id: %v", err))
	}

	return int64(id)
}

func getDefaultRegion(cookie, csrfToken string) (string, string, error) {
	cfg := &AggConfig{
		Cookie:        cookie,
		BaseURL:       DefaultBaseURL,
		CSRFToken:     csrfToken,
		Timeout:       DefaultTimeoutSec,
		MaxRetryTimes: sdk.Int(DefaultMaxRetryTimes),
	}
	client, err := newUAccountClientForConfig(cfg)
	if err != nil {
		return "", "", err
	}
	req := client.NewGetRegionRequest()
	resp, err := client.GetRegion(req)
	if err != nil {
		return "", "", err
	}
	for _, r := range resp.Regions {
		if r.IsDefault {
			return r.Region, r.Zone, nil
		}
	}
	return "", "", fmt.Errorf("default region not found")
}

func getDefaultProject(cookie, csrfToken string) (string, string, error) {
	cfg := &AggConfig{
		Cookie:        cookie,
		BaseURL:       DefaultBaseURL,
		CSRFToken:     csrfToken,
		Timeout:       DefaultTimeoutSec,
		MaxRetryTimes: sdk.Int(DefaultMaxRetryTimes),
	}
	client, err := newUAccountClientForConfig(cfg)
	if err != nil {
		return "", "", err
	}

	req := client.NewGetProjectListRequest()
	resp, err := client.GetProjectList(req)
	if err != nil {
		return "", "", err
	}
	for _, project := range resp.ProjectSet {
		if project.IsDefault == true {
			return project.ProjectId, project.ProjectName, nil
		}
	}
	return "", "", fmt.Errorf("default project not found")
}

func newUAccountClientForConfig(cfg *AggConfig) (*uaccount.UAccountClient, error) {
	sdkConfig, credConfig, err := BuildClientRuntime(cfg)
	client := uaccount.NewClient(sdkConfig, BuildCredentialFrom(credConfig))
	AttachHandlersWith(client, credConfig, cfg, AggConfigListIns)
	return client, err
}
