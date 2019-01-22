package base

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode"

	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"
	"github.com/ucloud/ucloud-sdk-go/ucloud/helpers/waiter"
	"github.com/ucloud/ucloud-sdk-go/ucloud/log"
	"github.com/ucloud/ucloud-sdk-go/ucloud/response"

	"github.com/ucloud/ucloud-cli/model"
	"github.com/ucloud/ucloud-cli/ux"
)

//ConfigPath 配置文件路径
const ConfigPath = ".ucloud"

//GAP 表格列直接的间隔字符数
const GAP = 2

//Cxt 上下文
var Cxt = model.GetContext(os.Stdout)

//SdkClient 用于上报数据
var SdkClient *sdk.Client

//BizClient 用于调用业务接口
var BizClient *Client

//GetHomePath 获取家目录
func GetHomePath() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

//MosaicString 对字符串敏感部分打马赛克 如公钥私钥
func MosaicString(str string, beginChars, lastChars int) string {
	r := len(str) - lastChars - beginChars
	if r > 5 {
		return str[:beginChars] + strings.Repeat("*", 5) + str[(r+beginChars):]
	}
	return strings.Repeat("*", len(str))
}

//AppendToFile 添加到文件中
func AppendToFile(name string, content string) error {
	f, err := os.OpenFile(name, os.O_RDWR|os.O_APPEND, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(fmt.Sprintf("\n%s\n", content))
	return err
}

//LineInFile 检查某一行是否在某文件中
func LineInFile(fileName string, lookFor string) bool {
	f, err := os.Open(fileName)
	if err != nil {
		return false
	}
	defer f.Close()
	r := bufio.NewReader(f)
	prefix := []byte{}
	for {
		line, isPrefix, err := r.ReadLine()
		if err == io.EOF {
			return false
		}
		if err != nil {
			return false
		}
		if isPrefix {
			prefix = append(prefix, line...)
			continue
		}
		line = append(prefix, line...)
		if string(line) == lookFor {
			return true
		}
		prefix = prefix[:0]
	}
}

//GetConfigPath 获取配置文件的绝对路径
func GetConfigPath() string {
	path := GetHomePath() + "/" + ConfigPath
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, 0755)
		if err != nil {
			panic(err)
		}
	}
	return path
}

//HandleBizError 处理RetCode != 0 的业务异常
func HandleBizError(resp response.Common) error {
	format := "Something wrong. RetCode:%d. Message:%s\n"
	Cxt.Printf(format, resp.GetRetCode(), resp.GetMessage())
	return fmt.Errorf(format, resp.GetRetCode(), resp.GetMessage())
}

//HandleError 处理错误，业务错误 和 HTTP错误
func HandleError(err error) {
	if uErr, ok := err.(uerr.Error); ok && uErr.Code() != 0 {
		format := "Something wrong. RetCode:%d. Message:%s\n"
		Cxt.Printf(format, uErr.Code(), uErr.Message())
	} else {
		Cxt.PrintErr(err)
	}
}

//PrintJSON 以JSON格式打印数据集合
func PrintJSON(dataSet interface{}) error {
	bytes, err := json.MarshalIndent(dataSet, "", "  ")
	if err != nil {
		return err
	}
	Cxt.Println(string(bytes))
	return nil
}

//PrintTableS 简化版表格打印，无需传表头，根据结构体反射解析
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

//PrintList 打印表格或者JSON
func PrintList(dataSet interface{}, json bool) {
	if json {
		PrintJSON(dataSet)
	} else {
		PrintTableS(dataSet)
	}
}

//PrintDescribe 打印详情
func PrintDescribe(attrs []DescribeTableRow, json bool) {
	if json {
		PrintJSON(attrs)
	} else {
		for _, attr := range attrs {
			fmt.Println(attr.Attribute)
			fmt.Println(attr.Content)
			fmt.Println()
		}
	}
}

//PrintTable 以表格方式打印数据集合
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
		rows := []map[string]interface{}{}
		for j := 0; j < elemVal.NumField(); j++ {
			field := elemVal.Field(j)
			fieldName := elemType.Field(j).Name
			if _, ok := showFieldMap[fieldName]; ok {
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
	fmt.Printf("\n")

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

//DescribeTableRow 详情表格通用表格行
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

//FormatDate 格式化时间,把以秒为单位的时间戳格式化未年月日
func FormatDate(seconds int) string {
	return time.Unix(int64(seconds), 0).Format("2006-01-02")
}

//FormatDateTime 格式化时间,把以秒为单位的时间戳格式化未年月日/时分秒
func FormatDateTime(seconds int) string {
	return time.Unix(int64(seconds), 0).Format("2006-01-02/15:04:05")
}

//RegionLabel regionlable
var RegionLabel = map[string]string{
	"cn-bj1":       "Beijing1",
	"cn-bj2":       "Beijing2",
	"cn-sh2":       "Shanghai2",
	"cn-gd":        "Guangzhou",
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

//Poller 轮询器
type Poller struct {
	stateFields   []string
	DescribeFunc  func(string, string, string, string) (interface{}, error)
	Out           io.Writer
	Timeout       time.Duration
	SdescribeFunc func(string) (interface{}, error)
}

//Spoll 简化版
func (p *Poller) Spoll(resourceID, pollText string, targetStates []string) {
	w := waiter.StateWaiter{
		Pending: []string{"pending"},
		Target:  []string{"avaliable"},
		Refresh: func() (interface{}, string, error) {
			inst, err := p.SdescribeFunc(resourceID)
			if err != nil {
				return nil, "", err
			}

			if inst == nil {
				return nil, "pending", nil
			}
			instValue := reflect.ValueOf(inst)
			instValue = reflect.Indirect(instValue)
			instType := instValue.Type()
			if instValue.Kind() != reflect.Struct {
				return nil, "", fmt.Errorf("Instance is not struct")
			}
			state := ""
			for i := 0; i < instValue.NumField(); i++ {
				for _, sf := range p.stateFields {
					if instType.Field(i).Name == sf {
						state = instValue.Field(i).String()
					}
				}
			}
			if state != "" {
				for _, t := range targetStates {
					if t == state {
						return inst, "avaliable", nil
					}
				}
			}
			return nil, "pending", nil

		},
		Timeout: p.Timeout,
	}

	done := make(chan bool)
	go func() {
		if resp, err := w.Wait(); err != nil {
			log.Error(err)
			if _, ok := err.(*waiter.TimeoutError); ok {
				done <- false
				return
			}
		} else {
			log.Infof("%#v", resp)
		}
		done <- true
	}()

	spinner := ux.NewDotSpinner(p.Out)
	spinner.Start(pollText)
	ret := <-done
	if ret {
		spinner.Stop()
	} else {
		spinner.Timeout()
	}
}

//Poll function
func (p *Poller) Poll(resourceID, projectID, region, zone, pollText string, targetState []string) {
	w := waiter.StateWaiter{
		Pending: []string{"pending"},
		Target:  []string{"avaliable"},
		Refresh: func() (interface{}, string, error) {
			inst, err := p.DescribeFunc(resourceID, projectID, region, zone)
			if err != nil {
				return nil, "", err
			}

			if inst == nil {
				return nil, "pending", nil
			}
			instValue := reflect.ValueOf(inst)
			instValue = reflect.Indirect(instValue)
			instType := instValue.Type()
			if instValue.Kind() != reflect.Struct {
				return nil, "", fmt.Errorf("Instance is not struct")
			}
			state := ""
			for i := 0; i < instValue.NumField(); i++ {
				for _, sf := range p.stateFields {
					if instType.Field(i).Name == sf {
						state = instValue.Field(i).String()
					}
				}
			}
			if state != "" {
				for _, t := range targetState {
					if t == state {
						return inst, "avaliable", nil
					}
				}
			}
			return nil, "pending", nil

		},
		Timeout: p.Timeout,
	}

	done := make(chan bool)
	go func() {
		if resp, err := w.Wait(); err != nil {
			log.Error(err)
			if _, ok := err.(*waiter.TimeoutError); ok {
				done <- false
				return
			}
		} else {
			log.Infof("%#v", resp)
		}
		done <- true
	}()

	spinner := ux.NewDotSpinner(p.Out)
	spinner.Start(pollText)
	ret := <-done
	if ret {
		spinner.Stop()
	} else {
		spinner.Timeout()
	}
}

//NewSpoller simple
func NewSpoller(describeFunc func(string) (interface{}, error), out io.Writer) *Poller {
	return &Poller{
		SdescribeFunc: describeFunc,
		Out:           out,
		stateFields:   []string{"State", "Status"},
		Timeout:       10 * time.Minute,
	}
}

//NewPoller 轮询
func NewPoller(describeFunc func(string, string, string, string) (interface{}, error), out io.Writer) *Poller {
	return &Poller{
		DescribeFunc: describeFunc,
		Out:          out,
		stateFields:  []string{"State", "Status"},
		Timeout:      10 * time.Minute,
	}
}

//PickResourceID  uhost-xxx/uhost-name => uhost-xxx
func PickResourceID(str string) string {
	if strings.Index(str, "/") > -1 {
		return strings.SplitN(str, "/", 2)[0]
	}
	return str
}

//WriteJSONFile 写json文件
func WriteJSONFile(list interface{}, fileName string) error {
	byts, err := json.Marshal(list)
	if err != nil {
		return err
	}
	filePath := GetConfigPath() + "/" + fileName
	err = ioutil.WriteFile(filePath, byts, 0600)
	if err != nil {
		return err
	}
	return nil
}
