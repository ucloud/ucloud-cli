package cmd

import (
	"strings"
	"testing"

	"github.com/ucloud/ucloud-sdk-go/services/uaccount"
)

// 回归：auth login 后已有 project_id 必须用新账号的项目列表校验。
// 跨账号/跨站点遗留的 project_id 若原样保留，后续业务命令全部 RetCode 292 "Project not exists"。
func TestResolveLoginProject(t *testing.T) {
	projects := []uaccount.ProjectListInfo{
		{ProjectId: "org-111", ProjectName: "Default", IsDefault: true},
		{ProjectId: "org-222", ProjectName: "Dev"},
	}

	// case 1: project_id 为空 → 选账号默认项目（原有首登补链行为）
	id, notice, err := resolveLoginProject("", projects)
	if err != nil {
		t.Fatalf("empty existing: unexpected error: %v", err)
	}
	if id != "org-111" {
		t.Errorf("empty existing: id = %q, want default org-111", id)
	}
	if !strings.Contains(notice, "org-111") || !strings.Contains(notice, "Default") {
		t.Errorf("empty existing: notice = %q, want it to mention default project id and name", notice)
	}

	// case 2: project_id 属于当前账号 → 保持不变且无提示（AP-2 不覆写用户设置）
	id, notice, err = resolveLoginProject("org-222", projects)
	if err != nil {
		t.Fatalf("existing in list: unexpected error: %v", err)
	}
	if id != "org-222" {
		t.Errorf("existing in list: id = %q, want kept org-222", id)
	}
	if notice != "" {
		t.Errorf("existing in list: notice = %q, want empty (no behavior change)", notice)
	}

	// case 3: project_id 不属于当前账号 → 切到默认项目并给出明确提示
	id, notice, err = resolveLoginProject("org-stale", projects)
	if err != nil {
		t.Fatalf("existing not in list: unexpected error: %v", err)
	}
	if id != "org-111" {
		t.Errorf("existing not in list: id = %q, want default org-111", id)
	}
	if !strings.Contains(notice, "org-stale") || !strings.Contains(notice, "org-111") {
		t.Errorf("existing not in list: notice = %q, want it to mention stale id and new default", notice)
	}

	// case 4: 列表里没有默认项目 → 返回错误（调用方仅告警，不中断登录）
	noDefault := []uaccount.ProjectListInfo{{ProjectId: "org-333", ProjectName: "Solo"}}
	if _, _, err = resolveLoginProject("org-stale", noDefault); err == nil {
		t.Error("no default project: want error, got nil")
	}
}
