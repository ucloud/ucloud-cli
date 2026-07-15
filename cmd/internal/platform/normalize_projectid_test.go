package platform

import (
	"testing"

	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
)

// projectIDOnWire 按 SDK 默认 form 编码器编码 req，返回真正会上行的 ProjectId。
// 断言 wire 值而不是 GetPayload() 中间态：产品踩的坑正是「中间态已归一化、
// wire 上却是原值」，只有编码结果能证伪。
func projectIDOnWire(t *testing.T, req request.Common) (string, bool) {
	t.Helper()
	form, err := request.EncodeForm(req)
	if err != nil {
		t.Fatalf("encode form: %v", err)
	}
	v, ok := form["ProjectId"]
	return v, ok
}

// TestNormalizeProjectIDAcrossProductShapes 覆盖 master 上各产品传 project-id 的全部
// 形态（每条 name 标注取样来源），断言 wire 上的最终值。
//
// 平台补全 getProjectList 给出的候选是 "org-xxx/ProjectName"（cmd/project.go），
// 平台 handler 负责还原成纯 id。把 ProjectId 放进 generic payload map 的产品，
// 归一化会被 SDK 的 payload 覆盖语义吃掉 —— 详见 normalizeProjectID 的注释。
func TestNormalizeProjectIDAcrossProductShapes(t *testing.T) {
	const idName = "org-x/MyProject"
	const bare = "org-x"

	tests := []struct {
		name     string
		build    func(t *testing.T) request.Common
		wantWire string
		wantSet  bool
	}{
		{
			// 取样：umongodb create_replset.go:61 / utidb api.go:49
			//      sqlserver create.go:89 / pgsql(#127) supabase params()
			// 这是本次修复的目标形态：修复前 wire 上是 "org-x/MyProject"，
			// 网关报 RetCode 292 Project [org-x/MyProject] not exists。
			name: "generic payload map + id/name (umongodb/utidb/sqlserver/pgsql-supabase)",
			build: func(t *testing.T) request.Common {
				gr := &request.BaseGenericRequest{}
				if err := gr.SetPayload(map[string]interface{}{"Action": "X", "ProjectId": idName}); err != nil {
					t.Fatal(err)
				}
				return gr
			},
			wantWire: bare, wantSet: true,
		},
		{
			// 同上形态但用户传的已是纯 id（不按 Tab 补全）—— 修复前后行为必须一致。
			name: "generic payload map + bare id (未按 Tab，最常见)",
			build: func(t *testing.T) request.Common {
				gr := &request.BaseGenericRequest{}
				if err := gr.SetPayload(map[string]interface{}{"Action": "X", "ProjectId": bare}); err != nil {
					t.Fatal(err)
				}
				return gr
			},
			wantWire: bare, wantSet: true,
		},
		{
			// 取样：cloudwatch query_metric_data.go:180 (BindProjectID 绑 CommonBase)
			//      ukafka list.go:60 (genReq.SetProjectId)
			// payload 不含 ProjectId → 一直是好的，本次修复不得改变它。
			name: "generic CommonBase only + id/name (cloudwatch/ukafka)",
			build: func(t *testing.T) request.Common {
				gr := &request.BaseGenericRequest{}
				if err := gr.SetPayload(map[string]interface{}{"Action": "X"}); err != nil {
					t.Fatal(err)
				}
				if err := gr.SetProjectId(idName); err != nil {
					t.Fatal(err)
				}
				return gr
			},
			wantWire: bare, wantSet: true,
		},
		{
			// 取样：mysql create.go:50 (payload 只有 DBVersion/Region/Zone) / uddos (不传 project)
			// 无 project-id → wire 上不该凭空出现该字段。
			name: "generic 无 project-id (mysql/uddos)",
			build: func(t *testing.T) request.Common {
				gr := &request.BaseGenericRequest{}
				if err := gr.SetPayload(map[string]interface{}{"Action": "X"}); err != nil {
					t.Fatal(err)
				}
				return gr
			},
			wantWire: "", wantSet: false,
		},
		{
			// 绝大多数产品：typed SDK 请求 —— 历史行为，必须不变。
			name: "typed request + id/name (绝大多数产品)",
			build: func(t *testing.T) request.Common {
				req := &request.CommonBase{}
				if err := req.SetProjectId(idName); err != nil {
					t.Fatal(err)
				}
				return req
			},
			wantWire: bare, wantSet: true,
		},
		{
			name: "typed request + bare id",
			build: func(t *testing.T) request.Common {
				req := &request.CommonBase{}
				if err := req.SetProjectId(bare); err != nil {
					t.Fatal(err)
				}
				return req
			},
			wantWire: bare, wantSet: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := normalizeProjectID(tt.build(t))
			if err != nil {
				t.Fatalf("normalizeProjectID: %v", err)
			}
			got, ok := projectIDOnWire(t, out)
			if ok != tt.wantSet {
				t.Fatalf("ProjectId present on wire = %v, want %v (got %q)", ok, tt.wantSet, got)
			}
			if got != tt.wantWire {
				t.Errorf("ProjectId on wire = %q, want %q", got, tt.wantWire)
			}
		})
	}
}

// 归一化不得殃及 payload 里的其它字段 —— SetPayload 是整表替换，回归保护。
func TestNormalizeProjectIDLeavesOtherPayloadFieldsIntact(t *testing.T) {
	gr := &request.BaseGenericRequest{}
	if err := gr.SetPayload(map[string]interface{}{
		"Action":     "CreateUMongoDBReplSet",
		"ProjectId":  "org-x/MyProject",
		"Region":     "cn-bj2",
		"Zone":       "cn-bj2-02",
		"Name":       "my/instance/with/slashes", // 业务字段里的 "/" 绝不能被 pick
		"DiskSpace":  100,                        // int 必须仍是 int（JSON 编码器依赖它）
		"IsMemoryDB": true,
	}); err != nil {
		t.Fatal(err)
	}
	out, err := normalizeProjectID(gr)
	if err != nil {
		t.Fatal(err)
	}
	payload := out.(request.GenericRequest).GetPayload()
	if payload["ProjectId"] != "org-x" {
		t.Errorf("ProjectId = %v, want org-x", payload["ProjectId"])
	}
	if payload["Name"] != "my/instance/with/slashes" {
		t.Errorf("business field with slashes was mangled: %v", payload["Name"])
	}
	if payload["DiskSpace"] != 100 {
		t.Errorf("DiskSpace = %#v, want int 100 (type must survive for JSON encoder)", payload["DiskSpace"])
	}
	if payload["IsMemoryDB"] != true {
		t.Errorf("IsMemoryDB = %#v, want bool true", payload["IsMemoryDB"])
	}
	if payload["Action"] != "CreateUMongoDBReplSet" || payload["Region"] != "cn-bj2" || payload["Zone"] != "cn-bj2-02" {
		t.Errorf("common fields changed: %v", payload)
	}
}
