package pathx

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	ppathx "github.com/ucloud/ucloud-sdk-go/private/services/pathx"
	pathxsdk "github.com/ucloud/ucloud-sdk-go/services/pathx"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
)

var protocols = []string{"tcp", "udp"}

var regionLabel = map[string]string{
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

func formatPortList(userPorts []string) ([]string, error) {
	portList := make([]string, 0)
	for _, port := range userPorts {
		if strings.Contains(port, "-") {
			portRange := strings.Split(port, "-")
			if len(portRange) != 2 {
				return nil, fmt.Errorf("port %s is invalid, it's pattern should be like 3000-3100", port)
			}
			min, err := strconv.Atoi(portRange[0])
			if err != nil {
				return nil, fmt.Errorf("parse port failed: %v", err)
			}
			max, err := strconv.Atoi(portRange[1])
			if err != nil {
				return nil, fmt.Errorf("parse port failed: %v", err)
			}
			for i := min; i <= max; i++ {
				portList = append(portList, strconv.Itoa(i))
			}
		} else {
			portList = append(portList, port)
		}
	}
	return portList, nil
}

func getUpathStr(list []ppathx.UPathSet) string {
	paths := make([]string, 0)
	for _, p := range list {
		paths = append(paths, fmt.Sprintf("%s->%s %dM", p.LineFromName, p.LineToName, p.Bandwidth))
	}
	return strings.Join(paths, "\n")
}

func getOutIPStr(list []ppathx.OutPublicIpInfo) string {
	strs := make([]string, 0)
	for _, p := range list {
		strs = append(strs, fmt.Sprintf("%s %s", p.IP, regionLabel[p.Area]))
	}
	return strings.Join(strs, "\n")
}

func getPortStr(list []ppathx.UGAATask) string {
	strs := make([]string, 0)
	for _, t := range list {
		strs = append(strs, fmt.Sprintf("%s %d", t.Protocol, t.Port))
	}
	return strings.Join(strs, "\n")
}

func printPathxDetail(ctx *cli.Context, instanceInfo pathxsdk.ForwardInfo, out io.Writer) {
	attrs := []describeRow{
		{Attribute: "ResourceID", Content: instanceInfo.InstanceId},
		{Attribute: "CName", Content: instanceInfo.CName},
		{Attribute: "Name", Content: instanceInfo.Name},
		{Attribute: "AccelerationArea", Content: instanceInfo.AccelerationArea},
		{Attribute: "AccelerationAreaName", Content: instanceInfo.AccelerationAreaName},
		{Attribute: "OriginAreaCode", Content: instanceInfo.OriginAreaCode},
		{Attribute: "OriginArea", Content: instanceInfo.OriginArea},
		{Attribute: "Bandwidth", Content: strconv.Itoa(instanceInfo.Bandwidth)},
		{Attribute: "ChargeType", Content: instanceInfo.ChargeType},
		{Attribute: "IPList", Content: strings.Join(instanceInfo.IPList, ",")},
		{Attribute: "Domain", Content: instanceInfo.Domain},
		{Attribute: "Remark", Content: instanceInfo.Remark},
		{Attribute: "CreateTime", Content: common.FormatDateTime(instanceInfo.CreateTime)},
		{Attribute: "ExpireTime", Content: common.FormatDateTime(instanceInfo.ExpireTime)},
	}
	for _, attr := range attrs {
		fmt.Fprintf(out, "%-22s: %s\n", attr.Attribute, attr.Content)
	}
	if len(instanceInfo.AccelerationAreaInfos) > 0 {
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Acceleration area list:")
		for _, area := range instanceInfo.AccelerationAreaInfos {
			fmt.Fprintf(out, "%s:%5s\n", "Area", area.AccelerationArea)
			areaList := make([]PathxOptionalAreaRow, 0)
			for _, node := range area.AccelerationNodes {
				areaList = append(areaList, PathxOptionalAreaRow{
					AreaCode:    node.AreaCode,
					Area:        node.Area,
					FlagUnicode: node.FlagUnicode,
					FlagEmoji:   node.FlagEmoji,
				})
			}
			ctx.PrintList(areaList)
		}
	}
	if len(instanceInfo.EgressIpList) > 0 {
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Egress ip list:")
		egressIpList := make([]EgressIpInfoRow, 0)
		for _, egressIp := range instanceInfo.EgressIpList {
			egressIpList = append(egressIpList, EgressIpInfoRow{IP: egressIp.IP, Area: egressIp.Area})
		}
		ctx.PrintList(egressIpList)
	}
	if len(instanceInfo.PortSets) > 0 {
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Port list:")
		portList := make([]Uga3PortRow, 0)
		for _, portItem := range instanceInfo.PortSets {
			portList = append(portList, Uga3PortRow{
				Protocol: portItem.Protocol,
				Port:     portItem.Port,
				RSPort:   portItem.RSPort,
			})
		}
		ctx.PrintList(portList)
	}
}
