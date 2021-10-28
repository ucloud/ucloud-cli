# UCloud PathX Example Test


## Create PathX instance with port

* Extract "id" by regexp("ID is: ([^\s]+)"): "ucloud pathx create --bandwidth 1 --area-code CAN --charge-type Dynamic --quantity 1 --accel Global --origin-domain www.ucloud.cn --port 8000-8001 --origin-port 8000-8001 --protocol TCP"
* Execute command: "ucloud pathx list"
* Execute command with "id": "ucloud pathx list --id $id"
* Execute command with "id": "ucloud pathx list --id $id --detail"
* Execute command: "ucloud pathx price list --bandwidth 10 --area-code BKK"
* Execute command: "ucloud pathx area list"
* Execute command: "ucloud pathx area list --origin-domain www.ucloud.cn"
* Execute command: "ucloud pathx area list --origin-domain www.ucloud.cn --no-accel"
* Execute command: "ucloud pathx area list --origin-domain www.ucloud.cn --accel Global"
* Execute command with "id": "ucloud pathx delete -y --id $id"

## Create PathX instance without port

* Extract "id" by regexp("ID is: ([^\s]+)"): "ucloud pathx create --bandwidth 1 --area-code BKK --charge-type Dynamic --quantity 1 --accel AP --origin-domain www.ucloud.cn"
* Execute command with "id": "ucloud pathx modify --bandwidth 2 --id $id"
* Execute command with "id": "ucloud pathx modify --origin-domain pathx.ucloud.cn --id $id"
* Execute command with "id": "ucloud pathx modify --name PathX产品测试 --remark 测试 --id $id"
* Execute command with "id": "ucloud pathx delete -y --id $id"
