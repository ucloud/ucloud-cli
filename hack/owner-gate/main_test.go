package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// writeFile creates dir/name with content, making parent dirs (沿用 check-product 范式).
func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return path
}

// staticOwners builds a base-owners lookup from a fixed map (product -> owners);
// products absent from the map are treated as not existing at base.
func staticOwners(m map[string][]string) baseOwnersFunc {
	return func(product string) ([]string, bool, bool) {
		o, ok := m[product]
		return o, ok, false
	}
}

// staticOwnersParseErr returns a lookup where the named product exists at base
// but fails to parse (parseErr=true).
func staticOwnersParseErr(product string) baseOwnersFunc {
	return func(p string) ([]string, bool, bool) {
		if p == product {
			return nil, true, true
		}
		return nil, false, false
	}
}

func cf(path string) changedFile    { return changedFile{Path: path} }
func cfDel(path string) changedFile { return changedFile{Path: path, Deleted: true} }

// --------------------------------------------------------------------------
// productOf
// --------------------------------------------------------------------------

func TestProductOf(t *testing.T) {
	cases := []struct {
		path string
		want string
		ok   bool
	}{
		{"products/udb/internal/mysql/cmd.go", "udb", true},
		{"products/udb/product.yaml", "udb", true},
		{"products/eip/x.go", "eip", true},
		{"products/README.md", "", false}, // products/ 直属文件不属任何产品
		{"pkg/cli/context.go", "", false},
		{"go.mod", "", false},
		{".github/workflows/pr-gate.yml", "", false},
	}
	for _, c := range cases {
		got, ok := productOf(c.path)
		if got != c.want || ok != c.ok {
			t.Errorf("productOf(%q)=(%q,%v) want (%q,%v)", c.path, got, ok, c.want, c.ok)
		}
	}
}

// --------------------------------------------------------------------------
// decide — spec 要求的 8 个场景各一例(platformCleared=false:平台默认硬拦)
// --------------------------------------------------------------------------

func TestDecide_ProductAutonomous(t *testing.T) {
	d := decide(
		[]changedFile{cf("products/udb/internal/mysql/cmd.go")},
		"Episkey-G", false,
		staticOwners(map[string][]string{"udb": {"Episkey-G"}}),
	)
	if d.Type != "product" || !d.AutoMergeEligible || d.Blocking {
		t.Fatalf("expected product+eligible+nonblocking, got %+v", d)
	}
}

func TestDecide_PlatformFile(t *testing.T) {
	d := decide(
		[]changedFile{cf("pkg/cli/context.go")},
		"Episkey-G", false,
		staticOwners(map[string][]string{"udb": {"Episkey-G"}}),
	)
	if d.Type != "platform" || d.AutoMergeEligible || !d.Blocking {
		t.Fatalf("expected platform+ineligible+blocking, got %+v", d)
	}
}

func TestDecide_CrossProduct(t *testing.T) {
	d := decide(
		[]changedFile{cf("products/udb/x.go"), cf("products/eip/y.go")},
		"Episkey-G", false,
		staticOwners(map[string][]string{"udb": {"Episkey-G"}, "eip": {"Episkey-G"}}),
	)
	if d.Type != "platform" || d.AutoMergeEligible || !d.Blocking {
		t.Fatalf("expected platform (cross-product) blocking, got %+v", d)
	}
}

func TestDecide_NonOwnerEdit(t *testing.T) {
	d := decide(
		[]changedFile{cf("products/udb/internal/mysql/cmd.go")},
		"mallory", false,
		staticOwners(map[string][]string{"udb": {"Episkey-G"}}),
	)
	if d.Type != "platform" || d.AutoMergeEligible || !d.Blocking {
		t.Fatalf("expected platform (non-owner) blocking, got %+v", d)
	}
}

// base-vs-head 提权反例:HEAD 把 mallory 加进 owners,但 base 版只有 Episkey-G。
func TestDecide_SelfPromotionRejectedByBase(t *testing.T) {
	d := decide(
		[]changedFile{cf("products/udb/product.yaml")},
		"mallory", false,
		staticOwners(map[string][]string{"udb": {"Episkey-G"}}), // base: 无 mallory
	)
	if d.Type != "platform" || d.AutoMergeEligible || !d.Blocking {
		t.Fatalf("expected platform (self-promotion blocked by base), got %+v", d)
	}
}

// 改自己 owners:已是 base 版 owner 的人改 product.yaml(如加 co-owner)仍自治。
func TestDecide_OwnerEditsOwnOwners(t *testing.T) {
	d := decide(
		[]changedFile{cf("products/udb/product.yaml")},
		"Episkey-G", false,
		staticOwners(map[string][]string{"udb": {"Episkey-G"}}),
	)
	if d.Type != "product" || !d.AutoMergeEligible || d.Blocking {
		t.Fatalf("expected product+eligible (owner edits own owners), got %+v", d)
	}
}

// 新建产品:base 版无 products/newprod/ → 平台 PR(onboarding)。
func TestDecide_NewProductOnboarding(t *testing.T) {
	d := decide(
		[]changedFile{cf("products/newprod/product.yaml"), cf("products/newprod/product.go")},
		"Episkey-G", false,
		staticOwners(map[string][]string{}), // base: newprod 不存在
	)
	if d.Type != "platform" || d.AutoMergeEligible || !d.Blocking {
		t.Fatalf("expected platform (new product) blocking, got %+v", d)
	}
}

// 删产品:即便 author 是 owner,删除 product.yaml 也走平台审批(下线)。
func TestDecide_ProductOffboarding(t *testing.T) {
	d := decide(
		[]changedFile{cfDel("products/udb/product.yaml"), cfDel("products/udb/internal/mysql/cmd.go")},
		"Episkey-G", false,
		staticOwners(map[string][]string{"udb": {"Episkey-G"}}),
	)
	if d.Type != "platform" || d.AutoMergeEligible || !d.Blocking {
		t.Fatalf("expected platform (offboarding) blocking, got %+v", d)
	}
}

// --------------------------------------------------------------------------
// decide — 硬拦改造新增场景
// --------------------------------------------------------------------------

// 平台 PR 被放行(管理员自提或管理员批准)→ 仍 platform,但不再 blocking。
func TestDecide_PlatformClearedReleases(t *testing.T) {
	d := decide(
		[]changedFile{cf("base/biz_client.go")},
		"carol", true, // platformCleared
		staticOwners(map[string][]string{"udb": {"Episkey-G"}}),
	)
	if d.Type != "platform" || d.Blocking || d.AutoMergeEligible {
		t.Fatalf("expected platform+nonblocking+ineligible (cleared), got %+v", d)
	}
}

// cleared 只解硬拦,不把 non-owner 升级成 product 自治、不开 auto-merge。
func TestDecide_ClearedDoesNotPromoteNonOwner(t *testing.T) {
	d := decide(
		[]changedFile{cf("products/udb/internal/mysql/cmd.go")},
		"mallory", true, // cleared
		staticOwners(map[string][]string{"udb": {"Episkey-G"}}),
	)
	if d.Type != "platform" || d.AutoMergeEligible || d.Blocking {
		t.Fatalf("expected platform (non-owner) nonblocking+ineligible, got %+v", d)
	}
}

// product 自治路径不受 platformCleared 影响:两态都自动合、永不 blocking。
func TestDecide_ProductIgnoresCleared(t *testing.T) {
	for _, cleared := range []bool{false, true} {
		d := decide(
			[]changedFile{cf("products/udb/internal/mysql/cmd.go")},
			"Episkey-G", cleared,
			staticOwners(map[string][]string{"udb": {"Episkey-G"}}),
		)
		if d.Type != "product" || !d.AutoMergeEligible || d.Blocking {
			t.Fatalf("cleared=%v: expected product+eligible+nonblocking, got %+v", cleared, d)
		}
	}
}

// 跨产品 rename:old 端与 new 端分属两产品 → cross-product 平台红。
func TestDecide_CrossProductRename(t *testing.T) {
	d := decide(
		[]changedFile{
			{Path: "products/udb/cmd/shared.go", Deleted: true}, // rename old 端
			{Path: "products/uhost/cmd/shared.go", Deleted: false},
		},
		"Episkey-G", false,
		staticOwners(map[string][]string{"udb": {"Episkey-G"}, "uhost": {"Episkey-G"}}),
	)
	if d.Type != "platform" || !d.Blocking {
		t.Fatalf("expected platform (cross-product rename) blocking, got %+v", d)
	}
}

// rename-away product.yaml(搬离 canonical 路径)等同下线 → 平台红。
func TestDecide_RenameAwayProductYAMLIsOffboarding(t *testing.T) {
	d := decide(
		[]changedFile{
			{Path: "products/udb/product.yaml", Deleted: true}, // rename old 端
			{Path: "products/udb/product.yaml.bak", Deleted: false},
		},
		"Episkey-G", false, // 即便作者是 owner
		staticOwners(map[string][]string{"udb": {"Episkey-G"}}),
	)
	if d.Type != "platform" || !d.Blocking {
		t.Fatalf("expected platform (rename-away offboarding) blocking, got %+v", d)
	}
}

// 空 diff(仅 merge commit / 无改动)→ noop,不当平台拦。
func TestDecide_EmptyDiffNoop(t *testing.T) {
	d := decide(nil, "Episkey-G", false,
		staticOwners(map[string][]string{"udb": {"Episkey-G"}}))
	if d.Type != "noop" || d.Blocking || d.AutoMergeEligible {
		t.Fatalf("expected noop+nonblocking+ineligible, got %+v", d)
	}
}

// base 版 product.yaml 解析失败 → 独立平台文案(不误称 non-owner)。
func TestDecide_ParseFailureDistinctReason(t *testing.T) {
	d := decide(
		[]changedFile{cf("products/udb/internal/mysql/cmd.go")},
		"Episkey-G", false,
		staticOwnersParseErr("udb"),
	)
	if d.Type != "platform" || !d.Blocking {
		t.Fatalf("expected platform (parse error) blocking, got %+v", d)
	}
	if !strings.Contains(d.Reason, "解析失败") {
		t.Fatalf("expected reason to mention 解析失败, got %q", d.Reason)
	}
}

// --------------------------------------------------------------------------
// parseNameStatus — 解析 `git diff --name-status`(rename 两端均计入)
// --------------------------------------------------------------------------

func TestParseNameStatus(t *testing.T) {
	in := "M\tproducts/udb/product.go\n" +
		"A\tproducts/udb/internal/mysql/new.go\n" +
		"D\tproducts/udb/internal/mysql/old.go\n" +
		"R100\tproducts/udb/a.go\tproducts/udb/b.go\n" +
		"\n" // 空行应被跳过
	got, err := parseNameStatus(strings.NewReader(in))
	if err != nil {
		t.Fatalf("parseNameStatus: %v", err)
	}
	// M, A, D 各 1 条 + rename 产出 old/new 2 条 = 5。
	if len(got) != 5 {
		t.Fatalf("expected 5 entries, got %d: %+v", len(got), got)
	}
	// 普通改动非删除
	if got[0].Deleted || got[0].Path != "products/udb/product.go" {
		t.Errorf("entry[0] should be modified product.go, got %+v", got[0])
	}
	// 删除标记
	if !got[2].Deleted || got[2].Path != "products/udb/internal/mysql/old.go" {
		t.Errorf("entry[2] should be deleted old.go, got %+v", got[2])
	}
	// rename old 端:标记 Deleted
	if !got[3].Deleted || got[3].Path != "products/udb/a.go" {
		t.Errorf("entry[3] should be rename-old a.go (deleted), got %+v", got[3])
	}
	// rename new 端:非删除
	if got[4].Deleted || got[4].Path != "products/udb/b.go" {
		t.Errorf("entry[4] should be rename-new b.go (not deleted), got %+v", got[4])
	}
}

// 跨产品 rename 行:old/new 两端分属不同产品都要计入。
func TestParseNameStatus_CrossProductRename(t *testing.T) {
	in := "R100\tproducts/udb/cmd/shared.go\tproducts/uhost/cmd/shared.go\n"
	got, err := parseNameStatus(strings.NewReader(in))
	if err != nil {
		t.Fatalf("parseNameStatus: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 entries (old+new), got %d: %+v", len(got), got)
	}
	seen := map[string]bool{}
	for _, c := range got {
		if x, ok := productOf(c.Path); ok {
			seen[x] = true
		}
	}
	if !seen["udb"] || !seen["uhost"] {
		t.Fatalf("expected both udb and uhost in productSet, got %v", seen)
	}
}

// --------------------------------------------------------------------------
// gitShowOwners — 必须读 BASE 版,而非 HEAD(安全命门)
// --------------------------------------------------------------------------

func gitRun(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t",
		"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func TestGitShowOwners_ReadsBaseNotHead(t *testing.T) {
	repo := t.TempDir()
	gitRun(t, repo, "init", "-q", "-b", "main")

	// base commit: owners = [alice]
	writeFile(t, repo, "products/udb/product.yaml",
		"name: udb\nowners:\n  - alice\ncommands: [mysql]\nenabled: true\n")
	gitRun(t, repo, "add", "-A")
	gitRun(t, repo, "commit", "-qm", "base")
	baseSHA := gitRun(t, repo, "rev-parse", "HEAD")

	// head commit: PR 把 bob 加进 owners(自我提权尝试)
	writeFile(t, repo, "products/udb/product.yaml",
		"name: udb\nowners:\n  - alice\n  - bob\ncommands: [mysql]\nenabled: true\n")
	gitRun(t, repo, "add", "-A")
	gitRun(t, repo, "commit", "-qm", "head")

	t.Chdir(repo) // gitShowOwners 在进程 cwd 下跑 git

	owners, exists, parseErr := gitShowOwners(baseSHA, "udb")
	if !exists || parseErr {
		t.Fatalf("expected base product.yaml to exist and parse, got exists=%v parseErr=%v", exists, parseErr)
	}
	if len(owners) != 1 || owners[0] != "alice" {
		t.Fatalf("expected base owners [alice] (must NOT see head's bob), got %v", owners)
	}

	// 不存在的产品 → baseExists=false
	if _, exists, _ := gitShowOwners(baseSHA, "ghost"); exists {
		t.Fatal("expected ghost product to be absent at base")
	}
}

// base 版 product.yaml 存在但 YAML 解析失败 → parseErr=true、baseExists=true。
func TestGitShowOwners_ParseError(t *testing.T) {
	repo := t.TempDir()
	gitRun(t, repo, "init", "-q", "-b", "main")
	// 未定义 anchor 的别名引用 → yaml.v2 解析报错。
	writeFile(t, repo, "products/udb/product.yaml", "owners: *nope\n")
	gitRun(t, repo, "add", "-A")
	gitRun(t, repo, "commit", "-qm", "base")
	baseSHA := gitRun(t, repo, "rev-parse", "HEAD")

	t.Chdir(repo)
	owners, exists, parseErr := gitShowOwners(baseSHA, "udb")
	if !exists || !parseErr {
		t.Fatalf("expected exists=true parseErr=true, got exists=%v parseErr=%v", exists, parseErr)
	}
	if owners != nil {
		t.Fatalf("expected nil owners on parse error, got %v", owners)
	}
}
