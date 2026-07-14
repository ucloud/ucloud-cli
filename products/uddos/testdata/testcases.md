# uddos CLI 测试用例

覆盖范围：`ucloud uddos mainland` 与 `ucloud uddos overseas` 下全部子命令。

---

## 一、国内高防 — mainland service

### 1.1 service create

**前置条件**：已登录，账号有购买高防服务的权限。

| TC# | 场景 | 命令 | 预期结果 |
|-----|------|------|----------|
| M-SVC-C-01 | 正常创建（华东，枣庄机房） | `ucloud uddos mainland service create --charge-type Month --quantity 1 --area-line EastChina --engine-room Zaozhuang --src-bandwidth 100 --defence-base-flow 30 --defence-max-flow 50 --name my-svc` | 输出包含新建服务的 ResourceID，状态 Created |
| M-SVC-C-02 | 正常创建（华东，扬州机房） | 同上，`--engine-room Yangzhou` | 正常返回 ResourceID |
| M-SVC-C-03 | 正常创建（华北，石家庄机房） | `--area-line NorthChina --engine-room Shijiazhuang` | 正常返回 ResourceID |
| M-SVC-C-04 | base-flow = max-flow | `--defence-base-flow 30 --defence-max-flow 30` | 正常创建，不报错 |
| M-SVC-C-05 | 按年计费 | `--charge-type Year --quantity 1` | 正常创建 |
| M-SVC-C-06 | src-bandwidth 为最小值 50 | `--src-bandwidth 50` | 正常创建 |
| M-SVC-C-07 | src-bandwidth 为 10 的整百倍 | `--src-bandwidth 200` | 正常创建 |
| M-SVC-C-08 | **charge-type 非法值** | `--charge-type Daily` | 报错：`invalid --charge-type "Daily", must be "Month" or "Year"` |
| M-SVC-C-09 | **area-line 非法值** | `--area-line SouthChina` | 报错：`invalid --area-line "SouthChina"` |
| M-SVC-C-10 | **engine-room 与 area-line 不匹配** | `--area-line EastChina --engine-room Shijiazhuang` | 报错：`invalid --engine-room "Shijiazhuang" for --area-line "EastChina"` |
| M-SVC-C-11 | **engine-room 与 NorthChina 不匹配** | `--area-line NorthChina --engine-room Zaozhuang` | 报错：`invalid --engine-room "Zaozhuang" for --area-line "NorthChina"` |
| M-SVC-C-12 | **src-bandwidth 低于下限** | `--src-bandwidth 40` | 报错：`--src-bandwidth minimum is 50` |
| M-SVC-C-13 | **src-bandwidth 不是 10 的倍数** | `--src-bandwidth 55` | 报错：`--src-bandwidth must be a multiple of 10` |
| M-SVC-C-14 | **defence-base-flow 不在白名单** | `--defence-base-flow 35` | 报错：`invalid --defence-base-flow 35` |
| M-SVC-C-15 | **defence-max-flow 不在白名单** | `--defence-max-flow 45` | 报错：`invalid --defence-max-flow 45` |
| M-SVC-C-16 | **max-flow 小于 base-flow** | `--defence-base-flow 50 --defence-max-flow 30` | 报错：`--defence-max-flow (30) must be >= --defence-base-flow (50)` |
| M-SVC-C-17 | **缺少 required flag** | 省略 `--name` | cobra 报错：required flag not set |

**defence flow 白名单**：30 / 40 / 50 / 60 / 70 / 80 / 100 / 200 / 300 / 400 / 500 / 600 / 700 / 800 (Gbps)

---

### 1.2 service list

| TC# | 场景 | 命令 | 预期结果 |
|-----|------|------|----------|
| M-SVC-L-01 | 列出全部服务 | `ucloud uddos mainland service list` | 表格输出所有国内高防服务，含 ResourceID / Name / DefenceStatus / ExpireTime |
| M-SVC-L-02 | 按 resource-id 过滤 | `ucloud uddos mainland service list --resource-id ghp-xxxxx` | 仅返回指定服务行，其余不显示 |
| M-SVC-L-03 | 分页 offset/limit | `--offset 10 --limit 5` | 返回从第 11 条起的最多 5 条记录 |
| M-SVC-L-04 | resource-id 不存在 | `--resource-id ghp-notexist` | 输出空表格，不报错 |
| M-SVC-L-05 | JSON 输出格式 | `... --output json` | 输出合法 JSON 数组，字段完整 |

---

## 二、国内高防 — mainland ip

### 2.1 ip create

| TC# | 场景 | 命令 | 预期结果 |
|-----|------|------|----------|
| M-IP-C-01 | 正常创建，仅必填参数 | `ucloud uddos mainland ip create --resource-id ghp-xxxxx` | 输出新建 BGP IP 地址，状态 Created |
| M-IP-C-02 | 指定 block-udp=yes | `... --block-udp yes` | 正常创建，UDP 流量被屏蔽 |
| M-IP-C-03 | 指定 type-ip=TypeCharge | `... --type-ip TypeCharge` | 正常创建计费类型 IP |
| M-IP-C-04 | 指定 remark 和 tag | `... --remark "test" --tag "biz-group"` | 正常创建，API 参数中含 Remark 和 Tag |
| M-IP-C-05 | 指定 eip-region | `... --eip-region cn-bj2` | 正常创建，API 参数中含 EIPRegion |
| M-IP-C-06 | **缺少 resource-id** | 省略 `--resource-id` | cobra 报错：required flag not set |
| M-IP-C-07 | resource-id 不存在 | `--resource-id ghp-notexist` | API 返回业务错误，CLI 打印错误信息 |

---

### 2.2 ip list

| TC# | 场景 | 命令 | 预期结果 |
|-----|------|------|----------|
| M-IP-L-01 | 列出指定服务下所有 IP | `ucloud uddos mainland ip list --resource-id ghp-xxxxx` | 表格输出所有 BGP IP，含 DefenceIP / UserIP / LineType / Status / RuleCnt 等 |
| M-IP-L-02 | 按 bgp-ip 过滤 | `... --bgp-ip 1.2.3.4` | 仅返回匹配行 |
| M-IP-L-03 | 分页 | `... --offset 0 --limit 5` | 最多返回 5 条 |
| M-IP-L-04 | 无 IP 时输出 | `--resource-id ghp-empty` | 输出空表格，不报错 |
| M-IP-L-05 | **缺少 resource-id** | 省略 `--resource-id` | cobra 报错：required flag not set |

---

### 2.3 ip delete

| TC# | 场景 | 命令 | 预期结果 |
|-----|------|------|----------|
| M-IP-D-01 | 正常删除（带 --yes） | `ucloud uddos mainland ip delete --resource-id ghp-xxxxx --defence-ip 1.2.3.4 --yes` | 跳过确认，输出 `BGP IP deleted: 1.2.3.4`，状态 Deleted |
| M-IP-D-02 | 交互确认输入 y | 不带 `--yes`，stdin 输入 `y` | 执行删除，输出同上 |
| M-IP-D-03 | 交互确认输入 n | 不带 `--yes`，stdin 输入 `n` | 取消删除，无输出，不报错 |
| M-IP-D-04 | **缺少 resource-id** | 省略 `--resource-id` | cobra 报错：required flag not set |
| M-IP-D-05 | **缺少 defence-ip** | 省略 `--defence-ip` | cobra 报错：required flag not set |
| M-IP-D-06 | defence-ip 不存在 | `--defence-ip 9.9.9.9` | API 返回业务错误，CLI 打印错误信息 |

---

## 三、国内高防 — mainland rule

### 3.1 rule create

| TC# | 场景 | 命令 | 预期结果 |
|-----|------|------|----------|
| M-RL-C-01 | 最小参数创建 IP 协议规则 | `ucloud uddos mainland rule create --resource-id ghp-xxxxx --bgp-ip 1.2.3.4` | 输出 `rule[N] created for service[ghp-xxxxx]`，状态 Created |
| M-RL-C-02 | 指定 source-ip | `... --source-ip 10.0.0.1` | API 参数中含 SourceAddrArr、SourcePortArr、SourceToaIDArr |
| M-RL-C-03 | TCP 协议 + 端口 | `... --fwd-type TCP --bgp-ip-port 80` | 正常创建 TCP 转发规则 |
| M-RL-C-04 | UDP 协议 + 端口 | `... --fwd-type UDP --bgp-ip-port 53` | 正常创建 UDP 转发规则 |
| M-RL-C-05 | 开启负载均衡 | `... --load-balance Yes` | API 参数 LoadBalance=Yes |
| M-RL-C-06 | 指定 remark | `... --remark "main rule"` | API 参数含 Remark |
| M-RL-C-07 | **缺少 resource-id** | 省略 `--resource-id` | cobra 报错：required flag not set |
| M-RL-C-08 | **缺少 bgp-ip** | 省略 `--bgp-ip` | cobra 报错：required flag not set |

---

### 3.2 rule list

| TC# | 场景 | 命令 | 预期结果 |
|-----|------|------|----------|
| M-RL-L-01 | 列出指定服务下所有规则 | `ucloud uddos mainland rule list --resource-id ghp-xxxxx` | 表格输出所有规则，含 RuleIndex / BgpIP / FwdType / SourceIP / LoadBalance 等 |
| M-RL-L-02 | 按 bgp-ip 过滤 | `... --bgp-ip 1.2.3.4` | 仅返回该 IP 下的规则 |
| M-RL-L-03 | 按 rule-index 过滤 | `... --rule-index 0` | 仅返回 index=0 的规则 |
| M-RL-L-04 | 分页 | `... --limit 10 --offset 0` | 最多 10 条，默认 limit=32 |
| M-RL-L-05 | **缺少 resource-id** | 省略 `--resource-id` | cobra 报错：required flag not set |

---

### 3.3 rule update

| TC# | 场景 | 命令 | 预期结果 |
|-----|------|------|----------|
| M-RL-U-01 | 更新 source-ip | `ucloud uddos mainland rule update --resource-id ghp-xxxxx --bgp-ip 1.2.3.4 --rule-index 0 --source-ip 10.0.0.2` | 输出 `rule[0] updated for service[ghp-xxxxx]`，状态 Updated |
| M-RL-U-02 | 切换转发协议 | `... --fwd-type TCP --bgp-ip-port 443` | 正常更新 |
| M-RL-U-03 | 开启/关闭源地址探测 | `... --source-detect 1` | API 参数 SourceDetect=1 |
| M-RL-U-04 | 通过 rule-id 定位规则 | `... --rule-id rule-abc` | API 参数含 RuleID |
| M-RL-U-05 | **缺少 resource-id** | 省略 `--resource-id` | cobra 报错：required flag not set |
| M-RL-U-06 | **缺少 bgp-ip** | 省略 `--bgp-ip` | cobra 报错：required flag not set |
| M-RL-U-07 | **缺少 rule-index** | 省略 `--rule-index` | cobra 报错：required flag not set |

---

### 3.4 rule delete

| TC# | 场景 | 命令 | 预期结果 |
|-----|------|------|----------|
| M-RL-D-01 | 正常删除（带 --yes） | `ucloud uddos mainland rule delete --resource-id ghp-xxxxx --rule-index 0 --yes` | 输出 `rule[0] deleted from service[ghp-xxxxx]`，状态 Deleted |
| M-RL-D-02 | 交互确认输入 y | 不带 `--yes`，stdin 输入 `y` | 执行删除 |
| M-RL-D-03 | 交互确认输入 n | 不带 `--yes`，stdin 输入 `n` | 取消删除，不报错 |
| M-RL-D-04 | **缺少 resource-id** | 省略 `--resource-id` | cobra 报错：required flag not set |
| M-RL-D-05 | **缺少 rule-index** | 省略 `--rule-index` | cobra 报错：required flag not set |
| M-RL-D-06 | rule-index 不存在 | `--rule-index 9999` | API 返回业务错误，CLI 打印错误信息 |

---

## 四、海外高防 — overseas service

### 4.1 service create

**src-bandwidth 步进规则**：≤300 Mbps 步进 50；300~1000 步进 100；1000~5000 步进 500。

| TC# | 场景 | 命令 | 预期结果 |
|-----|------|------|----------|
| O-SVC-C-01 | 亚太清洗中心（HongKong） | `ucloud uddos overseas service create --charge-type Month --quantity 1 --engine-room HongKong --src-bandwidth 100 --name my-svc` | 正常返回 ResourceID，状态 Created |
| O-SVC-C-02 | 亚太城市节点（Singapore） | `--engine-room Singapore` | 正常创建，AreaLine 自动映射为 HongKong 清洗中心 |
| O-SVC-C-03 | 欧洲（Frankfurt） | `--engine-room Frankfurt` | 正常创建 |
| O-SVC-C-04 | 欧洲城市（London） | `--engine-room London` | 正常创建，AreaLine 映射为 Frankfurt |
| O-SVC-C-05 | 北美（Ashburn） | `--engine-room Ashburn` | 正常创建 |
| O-SVC-C-06 | 北美城市（LosAngeles） | `--engine-room LosAngeles` | 正常创建，AreaLine 映射为 Ashburn |
| O-SVC-C-07 | src-bandwidth 最小值 50 | `--src-bandwidth 50` | 正常创建 |
| O-SVC-C-08 | src-bandwidth 300（步进边界） | `--src-bandwidth 300` | 正常创建 |
| O-SVC-C-09 | src-bandwidth 400（进入 100 步进） | `--src-bandwidth 400` | 正常创建 |
| O-SVC-C-10 | src-bandwidth 1000（步进边界） | `--src-bandwidth 1000` | 正常创建 |
| O-SVC-C-11 | src-bandwidth 最大值 5000 | `--src-bandwidth 5000` | 正常创建 |
| O-SVC-C-12 | **charge-type 非法值** | `--charge-type Daily` | 报错：`invalid --charge-type "Daily"` |
| O-SVC-C-13 | **engine-room 不在映射表** | `--engine-room Shanghai` | 报错：`invalid --engine-room "Shanghai"` |
| O-SVC-C-14 | **src-bandwidth 低于下限** | `--src-bandwidth 49` | 报错：`--src-bandwidth minimum is 50` |
| O-SVC-C-15 | **src-bandwidth 超过上限** | `--src-bandwidth 5001` | 报错：`--src-bandwidth maximum is 5000` |
| O-SVC-C-16 | **≤300 时不是 50 的倍数** | `--src-bandwidth 75` | 报错：`must be a multiple of 50 when <= 300` |
| O-SVC-C-17 | **300~1000 时不是 100 的倍数** | `--src-bandwidth 350` | 报错：`must be a multiple of 100 when 300~1000` |
| O-SVC-C-18 | **1000~5000 时不是 500 的倍数** | `--src-bandwidth 1200` | 报错：`must be a multiple of 500 when 1000~5000` |
| O-SVC-C-19 | **缺少 required flag** | 省略 `--engine-room` | cobra 报错：required flag not set |

---

### 4.2 service list

| TC# | 场景 | 命令 | 预期结果 |
|-----|------|------|----------|
| O-SVC-L-01 | 列出全部海外服务 | `ucloud uddos overseas service list` | 表格输出所有 NapType=2 的服务，含 ResourceID / Name / DefenceStatus / ExpireTime / Remark |
| O-SVC-L-02 | 按 resource-id 过滤 | `... --resource-id nap-xxxxx` | 仅返回指定服务 |
| O-SVC-L-03 | 分页 | `... --offset 0 --limit 10` | 最多返回 10 条 |
| O-SVC-L-04 | 无服务时输出 | （账号下无海外高防） | 输出空表格，不报错 |

---

## 五、海外高防 — overseas ip

### 5.1 ip create

> 海外高防为透传模式（Passthrough），自动通过 `DescribeNapServiceInfo` + `GetNapServiceConfig` 解析 EIPRegion；也可通过 `--eip-region` 手动指定。

| TC# | 场景 | 命令 | 预期结果 |
|-----|------|------|----------|
| O-IP-C-01 | 自动解析 EIPRegion（省略参数） | `ucloud uddos overseas ip create --resource-id nap-xxxxx` | 自动查询服务配置获取 EIPRegion，创建 IP，输出 `BGP IP created: x.x.x.x` |
| O-IP-C-02 | 手动指定 eip-region | `... --eip-region hk` | 跳过自动解析，直接使用 hk，正常创建 |
| O-IP-C-03 | 指定 block-udp=yes | `... --block-udp yes` | 正常创建，UDP 被屏蔽 |
| O-IP-C-04 | 指定 type-ip=TypeCharge | `... --type-ip TypeCharge` | 正常创建计费类型 IP |
| O-IP-C-05 | 指定 remark 和 tag | `... --remark "r1" --tag "grp1"` | API 参数含 Remark / Tag |
| O-IP-C-06 | **缺少 resource-id** | 省略 `--resource-id` | cobra 报错：required flag not set |
| O-IP-C-07 | resource-id 查不到服务 | `--resource-id nap-notexist` | 报错：`service nap-notexist not found` |
| O-IP-C-08 | 服务配置 IpInfo 为空 | （数据库无 IpInfo 配置的服务） | 报错：`IpInfo is empty... please specify --eip-region manually` |

---

### 5.2 ip list

| TC# | 场景 | 命令 | 预期结果 |
|-----|------|------|----------|
| O-IP-L-01 | 列出指定服务下所有 IP | `ucloud uddos overseas ip list --resource-id nap-xxxxx` | 调用 `DescribePassthroughNapIP`，输出 EIPIP / EIPID / Status / EIPRegion / Tag / Remark |
| O-IP-L-02 | 按 nap-ip 过滤 | `... --nap-ip 1.2.3.4` | 仅返回匹配行，API 参数含 NapIp |
| O-IP-L-03 | 分页 | `... --offset 0 --limit 5` | 最多返回 5 条 |
| O-IP-L-04 | 无 IP 时输出 | `--resource-id nap-empty` | 输出空表格，不报错 |
| O-IP-L-05 | **缺少 resource-id** | 省略 `--resource-id` | cobra 报错：required flag not set |

---

### 5.3 ip delete

| TC# | 场景 | 命令 | 预期结果 |
|-----|------|------|----------|
| O-IP-D-01 | 正常删除（带 --yes） | `ucloud uddos overseas ip delete --resource-id nap-xxxxx --defence-ip 1.2.3.4 --yes` | 输出 `BGP IP deleted: 1.2.3.4`，状态 Deleted |
| O-IP-D-02 | 交互确认输入 y | 不带 `--yes`，stdin 输入 `y` | 执行删除 |
| O-IP-D-03 | 交互确认输入 n | 不带 `--yes`，stdin 输入 `n` | 取消删除，不报错 |
| O-IP-D-04 | **缺少 resource-id** | 省略 `--resource-id` | cobra 报错：required flag not set |
| O-IP-D-05 | **缺少 defence-ip** | 省略 `--defence-ip` | cobra 报错：required flag not set |

---

## 六、通用 / 输出格式测试

| TC# | 场景 | 命令 | 预期结果 |
|-----|------|------|----------|
| G-01 | JSON 输出格式 | 任意 list 命令 + `--output json` | 输出合法 JSON 数组，字段名与结构体一致 |
| G-02 | YAML 输出格式 | 任意 list 命令 + `--output yaml` | 输出合法 YAML |
| G-03 | 进度提示去向（JSON 模式） | 带 `--output json` 的 create/delete 命令 | 进度文本输出到 stderr，ResourceID 结果输出到 stdout |
| G-04 | 进度提示去向（Table 模式） | 不带 `--output`（默认表格） | 进度文本与结果均输出到 stdout |
| G-05 | 帮助信息 | `ucloud uddos --help` / `ucloud uddos mainland --help` | 输出可读的帮助文本，Example 字段完整 |
| G-06 | 命令树完整性 | `ucloud uddos mainland rule --help` | 列出 create / list / update / delete 四个子命令 |
| G-07 | 海外命令树完整性 | `ucloud uddos overseas ip --help` | 列出 create / list / delete 三个子命令（bind/unbind 暂未启用） |
