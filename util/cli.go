package util

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/ucloud/ucloud-sdk-go/sdk"
	uerr "github.com/ucloud/ucloud-sdk-go/sdk/error"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
	service "github.com/ucloud/ucloud-sdk-go/services"

	"github.com/ucloud/ucloud-cli/model"
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
var BizClient *service.Client

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
func MosaicString(s string, beginChars, lastChars int) string {
	r := len(s) - lastChars - beginChars
	if r > 0 {
		return s[:beginChars] + strings.Repeat("*", r) + s[(r+beginChars):]
	}
	return strings.Repeat("*", len(s))
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

//PrintTable 以表格方式打印数据集合
func PrintTable(dataSet interface{}, fieldList []string) error {
	dataSetVal := reflect.ValueOf(dataSet)

	switch dataSetVal.Kind() {
	case reflect.Slice, reflect.Array:
		displaySlice(dataSetVal, fieldList)
	case reflect.Map:
		displayMap(dataSetVal, fieldList)
	default:
		return fmt.Errorf("PrintTable expect array,slice or map, accept %T", dataSet)
	}
	return nil
}

func displayMap(mapVal reflect.Value, fieldList []string) {
	fmt.Println(mapVal, fieldList)
	//todo
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
		row := make(map[string]interface{})
		for j := 0; j < elemVal.NumField(); j++ {
			field := elemVal.Field(j)
			fieldName := elemType.Field(j).Name
			if _, ok := showFieldMap[fieldName]; ok {
				row[fieldName] = field.Interface()
				text := fmt.Sprintf("%v", field.Interface())
				width := calcWidth(text)
				if showFieldMap[fieldName] < width {
					showFieldMap[fieldName] = width
				}
			}
		}
		rowList = append(rowList, row)
	}

	for _, field := range fieldList {
		tmpl := "%-" + strconv.Itoa(showFieldMap[field]+GAP) + "s"
		fmt.Printf(tmpl, field)
	}
	fmt.Printf("\n")

	for _, row := range rowList {
		for _, field := range fieldList {
			cutWidth := calcCutWidth(fmt.Sprintf("%v", row[field]))
			tmpl := "%-" + strconv.Itoa(showFieldMap[field]-cutWidth+GAP) + "v"
			fmt.Printf(tmpl, row[field])
		}
		fmt.Printf("\n")
	}
}

func calcCutWidth(text string) int {
	set := []*unicode.RangeTable{unicode.Han}
	width := 0
	for _, r := range text {
		if unicode.IsOneOf(set, r) {
			width++
		}
	}
	return width
}

func calcWidth(text string) int {
	set := []*unicode.RangeTable{unicode.Han}
	width := 0
	for _, r := range text {
		if unicode.IsOneOf(set, r) {
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
