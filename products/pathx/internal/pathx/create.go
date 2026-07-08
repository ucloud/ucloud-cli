package pathx

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	pathxsdk "github.com/ucloud/ucloud-sdk-go/services/pathx"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newCreate ucloud pathx create
func newCreate(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, pathxsdk.NewClient)
	createPathxReq := client.NewCreateUGA3InstanceRequest()
	createPathxPortReq := client.NewCreateUGA3PortRequest()
	var ports, originPorts []string
	protocol := "tcp"
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create the pathx resource and port",
		Long:  "Create global unified access acceleration configuration item",
		Example: "ucloud pathx create --bandwidth 10 --area-code DXB" +
			"--charge-type Month --quantity 4 --accel Global --origin-ip 110.111.111.111" +
			"--protocol TCP --port 30654 --origin-port 30564",
		Run: func(cmd *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			fmt.Fprintln(w, "The pathx resource creating")
			if *createPathxReq.OriginIPList == "" && *createPathxReq.OriginDomain == "" {
				ctx.HandleError(fmt.Errorf("The origin-ip and origin-domain cannot be empty at the same time"))
				return
			}
			portIntList := make([]int, 0)
			originPortIntList := make([]int, 0)
			if len(ports) > 0 || len(originPorts) > 0 {
				if len(ports) == 0 {
					ctx.HandleError(fmt.Errorf("The port cannot be empty."))
					return
				} else if len(originPorts) == 0 {
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
				tcpPortList, err := formatPortList(ports)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				for _, tcpPort := range tcpPortList {
					port, _ := strconv.Atoi(tcpPort)
					portIntList = append(portIntList, port)
				}
				rsTcpPortList, err := formatPortList(originPorts)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				for _, rsTcpPort := range rsTcpPortList {
					rsPort, _ := strconv.Atoi(rsTcpPort)
					originPortIntList = append(originPortIntList, rsPort)
				}
				if len(portIntList) != len(originPortIntList) {
					ctx.HandleError(fmt.Errorf("The number of port must be consistent with the number of origin-port."))
					return
				} else if len(portIntList) >= 10 {
					ctx.HandleError(fmt.Errorf("The number of port cannot greater than or equals to 10"))
					return
				}
			}
			if strings.EqualFold(*createPathxReq.ChargeType, "Month") {
				*createPathxReq.Quantity = 0
			} else if *createPathxReq.Quantity <= 0 {
				ctx.HandleError(fmt.Errorf("If the value of charge-type is 'Year' or 'Hour',the value of quantity must be greater than 0"))
				return
			}
			switch strings.ToLower(*createPathxReq.ChargeType) {
			case "hour":
				*createPathxReq.ChargeType = "Dynamic"
			case "month":
				*createPathxReq.ChargeType = "Month"
			case "year":
				*createPathxReq.ChargeType = "Year"
			}
			createUGA3InstanceResp, err := client.CreateUGA3Instance(createPathxReq)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if createUGA3InstanceResp == nil || createUGA3InstanceResp.InstanceId == "" {
				ctx.HandleError(fmt.Errorf("An unknown error occurred and could not be created successfully."))
				return
			}
			if len(portIntList) > 0 && len(originPortIntList) > 0 {
				createPathxPortReq.InstanceId = &createUGA3InstanceResp.InstanceId
				createPathxPortReq.SetRegionRef(createPathxReq.GetRegionRef())
				createPathxPortReq.SetProjectIdRef(createPathxReq.GetProjectIdRef())
				createPathxPortReq.SetZoneRef(createPathxReq.GetZoneRef())
				fmt.Fprintln(w, "The pathx port creating")
				if strings.EqualFold(protocol, "TCP") {
					createPathxPortReq.TCP = portIntList
					createPathxPortReq.TCPRS = originPortIntList
				}
				_, err := client.CreateUGA3Port(createPathxPortReq)
				if err != nil {
					ctx.HandleError(err)
					return
				}
			}
			fmt.Fprintf(w, "The resource is created, and the resource ID is: %s\n", createUGA3InstanceResp.InstanceId)
			ctx.EmitResult(cli.OpResultRow{ResourceID: createUGA3InstanceResp.InstanceId, Action: "create", Status: "Created"})
		},
	}
	flags := createCmd.Flags()
	flags.SortFlags = false

	ctx.BindProjectID(createCmd, createPathxReq)
	ctx.BindRegion(createCmd, createPathxReq)
	ctx.BindZone(createCmd, createPathxReq)
	createPathxReq.Bandwidth = flags.String("bandwidth", "0", "Required. Shared bandwidth of the resource")
	createPathxReq.AreaCode = flags.String("area-code", "", "Optional. When it is empty,the nearest zone will be selected based on the origin-domain and origin-ip. Acceptable values:'BKK'(曼谷),'DXB'(迪拜),'FRA'(法兰克福),'SGN'(胡志明市),'HKG'(香港),'CGK'(雅加达),'LOS'(拉各斯),'LHR'(伦敦),'LAX'(洛杉矶),'MNL'(马尼拉),'DME'(莫斯科),'BOM'(孟买),'MSP'(圣保罗),'ICN'(首尔),'PVG'(上海),'SIN'(新加坡),'NRT'(东京),'IAD'(华盛顿),'TPE'(台北)")
	createPathxReq.ChargeType = flags.String("charge-type", "", "Optional. Payment method,its value is not case sensitive,acceptable values:'Year',pay yearly;'Month',pay monthly;'Hour', pay hourly")
	createPathxReq.Quantity = flags.Int("quantity", 1, "Optional. The duration of the pathx resource, the value cannot be less than or equal to 0. N years/months")
	createPathxReq.AccelerationArea = flags.String("accel", "", "Optional. The default value is 'Global'(全球). Other acceptable values:'AP'(亚太);'EU'(欧洲);'ME'(中东);'OA'(大洋洲);'AF'(非洲);'NA'(北美洲);'SA'(南美洲)")
	createPathxReq.OriginIPList = flags.String("origin-ip", "", "Optional. But when the origin-domain is empty,it cannot be empty. If multiple values exist,please split by ','. For example '0.0.0.0,110.110.100.100'")
	createPathxReq.OriginDomain = flags.String("origin-domain", "", "Optional. But when the origin-ip is empty,it cannot be empty")
	flags.StringSliceVar(&ports, "port", nil, "Optional. Disable 65123 port,the port can be multiple,please split by ',' for example 80,3000-3010. The number of port must be consistent with the number of origin-port,and the number cannot greater than or equals to 10")
	flags.StringSliceVar(&originPorts, "origin-port", nil, "Optional. The origin-port can be multiple,please split by ',' for example 80,3000-3010.The number of origin-port must be consistent with the number of port")
	flags.StringVar(&protocol, "protocol", "TCP", "Its values can be TCP and UDP, but currently only supports TCP")
	createCmd.MarkFlagRequired("bandwidth")
	command.SetFlagValues(createCmd, "area-code", "BKK", "DXB", "FRA", "SGN", "HKG", "CGK", "LOS", "LHR", "LAX", "MNL", "DME", "BOM", "MSP", "ICN", "PVG", "SIN", "NRT", "IAD", "TPE")
	command.SetFlagValues(createCmd, "charge-type", "Month", "Year", "Hour")
	command.SetFlagValues(createCmd, "accel", "Global", "AP", "EU", "ME", "OA", "AF", "NA", "SA")
	command.SetFlagValues(createCmd, "protocol", "TCP", "UDP")
	return createCmd
}
