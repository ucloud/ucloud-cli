## UHadoop CLI 测试清单

### 准备工作

```bash
# 1. 确认凭据已配置
ucloud config list

# 2. 查看可用地域
ucloud region

# 3. 切换 Go 版本
g use 1.25.12
alias go=/Users/user/.g/go/bin/go

# 4. 替换测试参数
# --region cn-bj2 --zone cn-bj2-02
```

---

### 一、查询命令

#### 1.1 list-framework-app — 查询框架和应用

```bash
go run . uhadoop list-framework-app --region <region> --zone <zone>
go run . uhadoop list-framework-app --region <region> --zone <zone> --output json
```

#### 1.2 list-node-type — 查询可用机型

```bash
go run . uhadoop list-node-type --region <region> --zone <zone>
go run . uhadoop list-node-type --region <region> --zone <zone> --node-role master
go run . uhadoop list-node-type --region <region> --zone <zone> --framework Hadoop --framework-version 3.3.4-udh3.2
```

#### 1.3 list — 列出集群

```bash
go run . uhadoop list --region <region>
go run . uhadoop list --region <region> --zone <zone> --limit 5
go run . uhadoop list --region <region> --id-only
go run . uhadoop list --all-region --output json
```

#### 1.4 describe — 查询集群详情

```bash
go run . uhadoop describe <instance-id> --region <region> --zone <zone>
go run . uhadoop describe <instance-id> --region <region> --zone <zone> --output json
```

---

### 二、修改命令

> ⚠️ 会对真实资源产生修改

#### 2.1 create — 创建集群

| 参数 | 必填 | 说明 |
|------|:--:|------|
| `--region` | ✅ | 地域 |
| `--zone` | ✅ | 可用区 |
| `--name` | ✅ | 集群名称 |
| `--framework` | ✅ | Hadoop / HDFS / MR / StarRocks-* |
| `--framework-version` | ✅ | 如 `3.3.4-udh3.2` |
| `--password` | ✅ | 登录密码（明文，自动 base64） |
| `--master-node-type` | | Master 机型，默认 `o.hadoop2m.xlarge` |
| `--core-node-type` | | Core 机型，默认 `o.hadoop2m.xlarge` |
| `--task-node-type` | | Task 机型（可选） |
| `--master-count` | | 默认 2（StarRocks 默认 3） |
| `--core-count` | | 默认 3 |
| `--task-count` | | 默认 0 |
| `--cluster-case` | | Spark / Hbase / Core-Hadoop |
| `--app-config` | | 手动指定组件，格式 `App#Version` |
| `--vpc-id` | | 可选 |
| `--subnet-id` | | 可选 |
| `--async` | | 异步模式，不等待创建完成 |

```bash
# Spark 模板创建
go run . uhadoop create \
  --region <region> --zone <zone> \
  --name test-spark --framework Hadoop \
  --framework-version 3.3.4-udh3.2 \
  --password 'YourPass123!' \
  --cluster-case Spark

# 手动指定 app-config
go run . uhadoop create \
  --region <region> --zone <zone> \
  --name test-manual --framework Hadoop \
  --framework-version 3.3.4-udh3.2 \
  --password 'YourPass123!' \
  --app-config Spark#3.5.3 --app-config Hive#3.1.3

# MR 集群（需要 --storage-cluster-id）
go run . uhadoop create \
  --region <region> --zone <zone> \
  --name test-mr --framework MR \
  --framework-version 3.3.4-udh3.2 \
  --password 'YourPass123!' \
  --storage-cluster-id <id> \
  --cluster-case Spark
```

#### 2.2 delete — 删除集群

| 参数 | 必填 | 说明 |
|------|:--:|------|
| `<instance-id>` | ✅ | 位置参数 |
| `--region` | ✅ | 地域 |
| `--zone` | ✅ | 可用区 |
| `--release-eip` | | 释放 EIP |
| `--yes` / `-y` | | 跳过确认 |

```bash
go run . uhadoop delete <instance-id> --region <region> --zone <zone>
go run . uhadoop delete <instance-id> --region <region> --zone <zone> --release-eip -y
```

#### 2.3 add-node — 扩容节点

| 参数 | 必填 | 说明 |
|------|:--:|------|
| `--region` | ✅ | 地域 |
| `--zone` | ✅ | 可用区 |
| `--instance-id` | ✅ | 集群 ID |
| `--node-role` | ✅ | core / task / client |
| `--node-type` | ✅ | 机型 |
| `--node-count` | | 默认 1 |
| `--async` | | 异步模式 |

```bash
go run . uhadoop add-node \
  --region <region> --zone <zone> \
  --instance-id <id> --node-role core \
  --node-type o.hadoop2m.xlarge
```

#### 2.4 restart-service — 启停服务

| 参数 | 必填 | 说明 |
|------|:--:|------|
| `--region` | ✅ | 地域 |
| `--zone` | ✅ | 可用区 |
| `--instance-id` | ✅ | 集群 ID |
| `--service-name` | ✅ | 服务名 |
| `--only-start` / `--only-stop` | | 只启/只停 |
| `--yes` / `-y` | | 跳过确认 |

```bash
go run . uhadoop restart-service \
  --region <region> --zone <zone> \
  --instance-id <id> --service-name Hive
```

#### 2.5 upgrade-node — 升级节点

| 参数 | 必填 | 说明 |
|------|:--:|------|
| `--region` | ✅ | 地域 |
| `--zone` | ✅ | 可用区 |
| `--instance-id` | ✅ | 集群 ID |
| `--node-role` | ✅ | master / core / task |
| `--node-type` | ✅ | 新机型 |
| `--node-name` | | 节点名（非 master 必填） |
| `--async` | | 异步模式 |
| `--yes` / `-y` | | 跳过确认 |

```bash
go run . uhadoop upgrade-node \
  --region <region> --zone <zone> \
  --instance-id <id> --node-role master \
  --node-type o.hadoop2m.xlarge
```

#### 2.6 upgrade-disk — 扩容磁盘

| 参数 | 必填 | 说明 |
|------|:--:|------|
| `--region` | ✅ | 地域 |
| `--zone` | ✅ | 可用区 |
| `--instance-id` | ✅ | 集群 ID |
| `--node-role` | ✅ | master / core / task |
| `--data-disk-size` | ✅ | 新数据盘大小 GB |
| `--boot-disk-size` | | 新系统盘大小 GB |
| `--node-name` | | 节点名（非 master 必填） |
| `--yes` / `-y` | | 跳过确认 |

```bash
go run . uhadoop upgrade-disk \
  --region <region> --zone <zone> \
  --instance-id <id> --node-role core \
  --data-disk-size 500 --node-name <id>-core1
```

---

### 三、校验

```bash
# 编译
go build ./...

# golden 测试
GOROOT=/Users/user/.g/go GOTOOLCHAIN=local go test ./hack/snapshot -v -run 'uhadoop'

# 缺参立即报错
go run . uhadoop create --name test
# → required flag(s) "framework", "framework-version", "password", "region", "zone" not set
```

---

### 测试结果

| # | 命令 | 状态 |
|---|------|------|
| 1 | `list-framework-app` | ⬜ |
| 2 | `list-node-type` | ⬜ |
| 3 | `list` | ⬜ |
| 4 | `describe` | ⬜ |
| 5 | `create --cluster-case Spark` | ⬜ |
| 6 | `create --app-config ...` | ⬜ |
| 7 | `delete` | ⬜ |
| 8 | `add-node` | ⬜ |
| 9 | `restart-service` | ⬜ |
| 10 | `upgrade-node` | ⬜ |
| 11 | `upgrade-disk` | ⬜ |
| 12 | `create` 缺参 | ⬜ |
| 13 | `go build ./...` | ⬜ |
| 14 | `go test ./hack/snapshot -v` | ⬜ |

> ⬜ = 待测试 &ensp; ✅ = 通过 &ensp; ❌ = 失败
