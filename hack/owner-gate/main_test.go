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
	return func(product string) ([]string, bool) {
		o, ok := m[product]
		return o, ok
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
// decide — spec 要求的 8 个场景各一例
// --------------------------------------------------------------------------

func TestDecide_ProductAutonomous(t *testing.T) {
	d := decide(
		[]changedFile{cf("products/udb/internal/mysql/cmd.go")},
		"Episkey-G",
		staticOwners(map[string][]string{"udb": {"Episkey-G"}}),
	)
	if d.Type != "product" || !d.AutoMergeEligible {
		t.Fatalf("expected product+eligible, got %+v", d)
	}
}

func TestDecide_PlatformFile(t *testing.T) {
	d := decide(
		[]changedFile{cf("pkg/cli/context.go")},
		"Episkey-G",
		staticOwners(map[string][]string{"udb": {"Episkey-G"}}),
	)
	if d.Type != "platform" || d.AutoMergeEligible {
		t.Fatalf("expected platform+ineligible, got %+v", d)
	}
}

func TestDecide_CrossProduct(t *testing.T) {
	d := decide(
		[]changedFile{cf("products/udb/x.go"), cf("products/eip/y.go")},
		"Episkey-G",
		staticOwners(map[string][]string{"udb": {"Episkey-G"}, "eip": {"Episkey-G"}}),
	)
	if d.Type != "platform" || d.AutoMergeEligible {
		t.Fatalf("expected platform (cross-product), got %+v", d)
	}
}

func TestDecide_NonOwnerEdit(t *testing.T) {
	d := decide(
		[]changedFile{cf("products/udb/internal/mysql/cmd.go")},
		"mallory",
		staticOwners(map[string][]string{"udb": {"Episkey-G"}}),
	)
	if d.Type != "platform" || d.AutoMergeEligible {
		t.Fatalf("expected platform (non-owner), got %+v", d)
	}
}

// base-vs-head 提权反例:HEAD 把 mallory 加进 owners,但 base 版只有 Episkey-G。
// 由于 baseOwners 只返回 base 版(不含 mallory),判定必须挡住。
func TestDecide_SelfPromotionRejectedByBase(t *testing.T) {
	d := decide(
		[]changedFile{cf("products/udb/product.yaml")},
		"mallory",
		staticOwners(map[string][]string{"udb": {"Episkey-G"}}), // base: 无 mallory
	)
	if d.Type != "platform" || d.AutoMergeEligible {
		t.Fatalf("expected platform (self-promotion blocked by base), got %+v", d)
	}
}

// 改自己 owners:已是 base 版 owner 的人改 product.yaml(如加 co-owner)仍自治。
func TestDecide_OwnerEditsOwnOwners(t *testing.T) {
	d := decide(
		[]changedFile{cf("products/udb/product.yaml")},
		"Episkey-G",
		staticOwners(map[string][]string{"udb": {"Episkey-G"}}),
	)
	if d.Type != "product" || !d.AutoMergeEligible {
		t.Fatalf("expected product+eligible (owner edits own owners), got %+v", d)
	}
}

// 新建产品:base 版无 products/newprod/ → 平台 PR(onboarding)。
func TestDecide_NewProductOnboarding(t *testing.T) {
	d := decide(
		[]changedFile{cf("products/newprod/product.yaml"), cf("products/newprod/product.go")},
		"Episkey-G",
		staticOwners(map[string][]string{}), // base: newprod 不存在
	)
	if d.Type != "platform" || d.AutoMergeEligible {
		t.Fatalf("expected platform (new product), got %+v", d)
	}
}

// 删产品:即便 author 是 owner,删除 product.yaml 也走平台审批(下线)。
func TestDecide_ProductOffboarding(t *testing.T) {
	d := decide(
		[]changedFile{cfDel("products/udb/product.yaml"), cfDel("products/udb/internal/mysql/cmd.go")},
		"Episkey-G",
		staticOwners(map[string][]string{"udb": {"Episkey-G"}}),
	)
	if d.Type != "platform" || d.AutoMergeEligible {
		t.Fatalf("expected platform (offboarding), got %+v", d)
	}
}

// --------------------------------------------------------------------------
// parseNameStatus — 解析 `git diff --name-status`
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
	if len(got) != 4 {
		t.Fatalf("expected 4 entries, got %d: %+v", len(got), got)
	}
	// 删除标记
	if !got[2].Deleted || got[2].Path != "products/udb/internal/mysql/old.go" {
		t.Errorf("entry[2] should be deleted old.go, got %+v", got[2])
	}
	// 改名取 HEAD 侧新路径,非删除
	if got[3].Deleted || got[3].Path != "products/udb/b.go" {
		t.Errorf("entry[3] should be rename->b.go (not deleted), got %+v", got[3])
	}
	// 普通改动非删除
	if got[0].Deleted || got[0].Path != "products/udb/product.go" {
		t.Errorf("entry[0] should be modified product.go, got %+v", got[0])
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

	owners, exists := gitShowOwners(baseSHA, "udb")
	if !exists {
		t.Fatal("expected base product.yaml to exist")
	}
	if len(owners) != 1 || owners[0] != "alice" {
		t.Fatalf("expected base owners [alice] (must NOT see head's bob), got %v", owners)
	}

	// 不存在的产品 → baseExists=false
	if _, exists := gitShowOwners(baseSHA, "ghost"); exists {
		t.Fatal("expected ghost product to be absent at base")
	}
}
