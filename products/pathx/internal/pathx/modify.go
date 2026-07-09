package pathx

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	pathxsdk "github.com/ucloud/ucloud-sdk-go/services/pathx"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newModify ucloud pathx modify
func newModify(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, pathxsdk.NewClient)
	modifyBandwidthReq := client.NewModifyUGA3BandwidthRequest()
	modifyOriginInfoReq := client.NewModifyUGA3OriginInfoRequest()
	modifyInstanceReq := client.NewModifyUGA3InstanceRequest()
	modifyPortReq := client.NewModifyUGA3PortRequest()
	var tcpPorts, rsTcpPorts []string
	var instanceId string
	protocol := "TCP"
	modifyCmd := &cobra.Command{
		Use:     "modify",
		Short:   "Modify the pathx associated information. Example bandwidth or origin information or resource information",
		Long:    "Support modify bandwidth,origin information,resource information,port",
		Example: "ucloud pathx modify --id uga3-xxx --bandwidth 1 --origin-ip 127.0.0.1 --name Pathx测试 --remark 加速资源 --protocol TCP --port 30010 --origin-port 39999",
		Run: func(cmd *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			modifyBandwidthReq.InstanceId = &instanceId
			modifyInstanceReq.InstanceId = &instanceId
			modifyOriginInfoReq.InstanceId = &instanceId
			modifyPortReq.InstanceId = &instanceId
			results := []cli.OpResultRow{}
			if *modifyBandwidthReq.Bandwidth != 0 {
				fmt.Fprintf(w, "Starting modify the pathx[%s] bandwidth\n", instanceId)
				if *modifyBandwidthReq.Bandwidth < 1 || *modifyBandwidthReq.Bandwidth > 100 {
					ctx.HandleError(fmt.Errorf("The value of bandwidth size cannot be less than 1 and cannot be greater than 100"))
					return
				}
				modifyBandwidthReq.SetProjectIdRef(modifyInstanceReq.GetProjectIdRef())
				modifyBandwidthReq.SetRegionRef(modifyInstanceReq.GetRegionRef())
				modifyBandwidthReq.SetZoneRef(modifyInstanceReq.GetZoneRef())
				_, err := client.ModifyUGA3Bandwidth(modifyBandwidthReq)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				results = append(results, cli.OpResultRow{ResourceID: instanceId, Action: "modify-bandwidth", Status: "Modified"})
			}
			if *modifyOriginInfoReq.OriginIPList != "" || *modifyOriginInfoReq.OriginDomain != "" {
				fmt.Fprintf(w, "Starting modify the pathx[%s] origin information\n", instanceId)
				modifyOriginInfoReq.SetProjectIdRef(modifyInstanceReq.GetProjectIdRef())
				modifyOriginInfoReq.SetRegionRef(modifyInstanceReq.GetRegionRef())
				modifyOriginInfoReq.SetZoneRef(modifyInstanceReq.GetZoneRef())
				_, err := client.ModifyUGA3OriginInfo(modifyOriginInfoReq)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				results = append(results, cli.OpResultRow{ResourceID: instanceId, Action: "modify-origin", Status: "Modified"})
			}
			if *modifyInstanceReq.Name != "" || *modifyInstanceReq.Remark != "" {
				fmt.Fprintf(w, "Starting modify the pathx[%s] resource information\n", instanceId)
				_, err := client.ModifyUGA3Instance(modifyInstanceReq)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				results = append(results, cli.OpResultRow{ResourceID: instanceId, Action: "modify", Status: "Modified"})
			}
			tcpPortIntList := make([]int, 0)
			rsTcpPortIntList := make([]int, 0)
			if len(tcpPorts) > 0 || len(rsTcpPorts) > 0 {
				fmt.Fprintf(w, "Starting modify the pathx[%s] port\n", instanceId)
				if len(tcpPorts) == 0 {
					ctx.HandleError(fmt.Errorf("The port cannot be empty."))
					return
				} else if len(rsTcpPorts) == 0 {
					ctx.HandleError(fmt.Errorf("The origin-port cannot be empty."))
					return
				}
				if strings.EqualFold(protocol, "UDP") {
					ctx.HandleError(fmt.Errorf("The udp protocol is temporarily not supported for create"))
					return
				} else if !strings.EqualFold(protocol, "TCP") && !strings.EqualFold(protocol, "UDP") {
					ctx.HandleError(fmt.Errorf("The value of protocol input error,please input 'TCP' or 'UDP',and the value entered is not case sensitive"))
					return
				}
				tcpPortList, err := formatPortList(tcpPorts)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				for _, tcpPort := range tcpPortList {
					port, _ := strconv.Atoi(tcpPort)
					tcpPortIntList = append(tcpPortIntList, port)
				}
				rsTcpPortList, err := formatPortList(rsTcpPorts)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				for _, rsTcpPort := range rsTcpPortList {
					rsPort, _ := strconv.Atoi(rsTcpPort)
					rsTcpPortIntList = append(rsTcpPortIntList, rsPort)
				}
				if len(tcpPortIntList) != len(rsTcpPortIntList) {
					ctx.HandleError(fmt.Errorf("The number of port must be consistent with the number of origin-port."))
					return
				} else if len(tcpPortIntList) >= 10 {
					ctx.HandleError(fmt.Errorf("The number of port cannot greater than or equals to 10"))
					return
				}
			}
			if len(tcpPortIntList) > 0 && len(rsTcpPortIntList) > 0 {
				if strings.EqualFold(protocol, "TCP") {
					modifyPortReq.TCP = tcpPortIntList
					modifyPortReq.TCPRS = rsTcpPortIntList
				}
				modifyPortReq.SetProjectIdRef(modifyInstanceReq.GetProjectIdRef())
				modifyPortReq.SetRegionRef(modifyInstanceReq.GetRegionRef())
				modifyPortReq.SetZoneRef(modifyInstanceReq.GetZoneRef())
				_, err := client.ModifyUGA3Port(modifyPortReq)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				results = append(results, cli.OpResultRow{ResourceID: instanceId, Action: "modify-port", Status: "Modified"})
			}
			ctx.EmitResult(results...)
		},
	}
	flags := modifyCmd.Flags()
	flags.SortFlags = false
	ctx.BindProjectID(modifyCmd, modifyInstanceReq)
	ctx.BindRegion(modifyCmd, modifyInstanceReq)
	ctx.BindZone(modifyCmd, modifyInstanceReq)
	flags.StringVar(&instanceId, "id", "", "Required. It is the resource ID of the pathx")
	modifyBandwidthReq.Bandwidth = flags.Int("bandwidth", 0, "Optional. The bandwidth size. Its value range [1-100],no update if no value is specified")
	modifyOriginInfoReq.OriginIPList = flags.String("origin-ip", "", "Optional. Acceleration source IP. If multiple values exist,please split by ','")
	modifyOriginInfoReq.OriginDomain = flags.String("origin-domain", "", "Optional. Acceleration source domain name. Only 1 domain is supported")
	modifyInstanceReq.Name = flags.String("name", "", "Optional. Accelerate configuration resource name. If its value is not filled in or an empty string is not updated")
	modifyInstanceReq.Remark = flags.String("remark", "", "Optional. It will be modified if its value is not empty")
	flags.StringSliceVar(&tcpPorts, "port", nil, "Optional. Disable 65123 port,the port can be multiple,please split by ',' for example 80,3000-3010. The number of port must be consistent with the number of origin-port,and the number cannot greater than or equals to 10")
	flags.StringSliceVar(&rsTcpPorts, "origin-port", nil, "Optional. The origin-port can be multiple,please split by ',' for example 80,3000-3010.The number of origin-port must be consistent with the number of port")
	flags.StringVar(&protocol, "protocol", "TCP", "Its values can be TCP and UDP, but currently only supports TCP")
	modifyCmd.MarkFlagRequired("id")
	ctx.SetCompletion(modifyCmd, "id", func() []string {
		return getPathxList(ctx, *modifyInstanceReq.ProjectId, *modifyInstanceReq.Region, *modifyInstanceReq.Zone)
	})
	return modifyCmd
}
