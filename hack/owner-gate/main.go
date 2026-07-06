// hack/owner-gate decides whether a pull request can be auto-merged by a
// product owner (product-autonomous PR) or must go through platform review
// (platform PR). See docs/ROADMAP.md P2a.
//
// Ownership is ALWAYS judged against the BASE revision of
// products/X/product.yaml — a PR cannot grant itself ownership by adding the
// author to owners in the same PR.
//
// owner-gate only ROUTES (product / platform / noop) and writes the verdict to
// GITHUB_OUTPUT; it never fails the check itself (exit 0 always, exit 2 only on
// an internal error). Whether a platform PR is BLOCKED is decided downstream in
// the workflow: a platform verdict is Blocking unless `platformCleared` is true
// (the PR author is an admin, OR an admin ≠ author has approved). The workflow
// computes platformCleared via the GitHub API and injects it; an admission step
// turns Blocking into a red required check (exit 1).
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"
)

// Decision is owner-gate's verdict.
type Decision struct {
	Type              string `json:"type"`              // "product" | "platform" | "noop"
	AutoMergeEligible bool   `json:"autoMergeEligible"` // true only for product PRs
	Blocking          bool   `json:"blocking"`          // true ⇒ admission step makes the check red
	Reason            string `json:"reason"`            // PR-visible explanation
}

// changedFile is one entry from `git diff --name-status`.
type changedFile struct {
	Path    string
	Deleted bool
}

// baseOwnersFunc returns the owners of products/<product>/product.yaml at the
// BASE revision, whether that file existed at base, and whether it existed but
// could not be parsed (parseErr) — the three cases drive distinct verdicts.
type baseOwnersFunc func(product string) (owners []string, baseExists bool, parseErr bool)

// productOf returns X when path is products/<X>/<...> (a file inside a product
// subtree). Files directly under products/ and files outside products/ are not
// product files.
func productOf(path string) (string, bool) {
	parts := strings.Split(path, "/")
	if len(parts) >= 3 && parts[0] == "products" {
		return parts[1], true
	}
	return "", false
}

// productYAMLDeleted reports whether this PR removes products/<x>/product.yaml —
// either by an outright delete or by renaming it away from its canonical path
// (parseNameStatus marks a rename's OLD side Deleted), both of which are
// offboarding and must go through platform review.
func productYAMLDeleted(changed []changedFile, x string) bool {
	want := fmt.Sprintf("products/%s/product.yaml", x)
	for _, c := range changed {
		if c.Deleted && c.Path == want {
			return true
		}
	}
	return false
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// parseNameStatus parses `git diff --name-status` output. Each line is
// "<status>\t<path>"; renames/copies are "<status>\t<old>\t<new>". For a
// rename/copy we record BOTH the old and the new path: the new path as the
// edited file, and the old path as removed-from (Deleted for a rename, kept for
// a copy). Recording both ends makes a cross-product move surface as touching
// two products, and a rename-away of product.yaml surface as a deletion.
func parseNameStatus(r io.Reader) ([]changedFile, error) {
	var out []changedFile
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimRight(sc.Text(), "\r\n")
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 2 {
			continue
		}
		status := fields[0]
		if (strings.HasPrefix(status, "R") || strings.HasPrefix(status, "C")) && len(fields) >= 3 {
			oldPath := fields[len(fields)-2]
			newPath := fields[len(fields)-1]
			// Rename removes the old path; copy leaves it in place.
			out = append(out, changedFile{Path: oldPath, Deleted: strings.HasPrefix(status, "R")})
			out = append(out, changedFile{Path: newPath, Deleted: false})
			continue
		}
		path := fields[len(fields)-1]
		out = append(out, changedFile{Path: path, Deleted: strings.HasPrefix(status, "D")})
	}
	return out, sc.Err()
}

// gitShowOwners reads products/<product>/product.yaml at baseSHA and returns its
// owners, whether the file existed at that revision (baseExists), and whether it
// existed but failed to parse (parseErr). Runs in the process working directory.
func gitShowOwners(baseSHA, product string) ([]string, bool, bool) {
	ref := fmt.Sprintf("%s:products/%s/product.yaml", baseSHA, product)
	raw, err := exec.Command("git", "show", ref).Output()
	if err != nil {
		// Non-zero exit ⇒ path absent at base ⇒ new product.
		return nil, false, false
	}
	var meta struct {
		Owners []string `yaml:"owners"`
	}
	if err := yaml.Unmarshal(raw, &meta); err != nil {
		return nil, true, true // exists but unparseable
	}
	return meta.Owners, true, false
}

// platformDecision builds a platform verdict. It is Blocking (→ red) unless the
// PR has been cleared (author is admin, or an admin ≠ author approved); the
// workflow computes `cleared` and the admission step consumes Blocking.
func platformDecision(detail string, cleared bool) Decision {
	if cleared {
		return Decision{
			Type:     "platform",
			Blocking: false,
			Reason:   "✅ 平台 PR 已放行(管理员自提或已获管理员批准),CI 通过后可合。判定:" + detail,
		}
	}
	return Decision{
		Type:     "platform",
		Blocking: true,
		Reason:   "🔴 平台 PR 默认硬拦,需管理员 Approve 放行(或由管理员提交)。判定:" + detail,
	}
}

// decide is the pure admission decision. platformCleared comes from the workflow
// (author admin OR admin≠author approved) and only affects platform verdicts.
func decide(changed []changedFile, author string, platformCleared bool, baseOwners baseOwnersFunc) Decision {
	if len(changed) == 0 {
		return Decision{Type: "noop", Reason: "空改动:无文件变更,无需准入裁决。"}
	}

	productSet := map[string]bool{}
	var platformFiles []string
	for _, c := range changed {
		if x, ok := productOf(c.Path); ok {
			productSet[x] = true
		} else {
			platformFiles = append(platformFiles, c.Path)
		}
	}

	if len(platformFiles) > 0 {
		sort.Strings(platformFiles)
		return platformDecision(fmt.Sprintf("改动触及平台/受保护路径(%s)", strings.Join(platformFiles, ", ")), platformCleared)
	}

	dirs := sortedKeys(productSet)
	switch len(dirs) {
	case 0:
		// Unreachable in practice (empty diff handled above; any non-empty file
		// is product or platform). Treat defensively as a no-op.
		return Decision{Type: "noop", Reason: "无产品目录改动。"}
	case 1:
		// handled below
	default:
		return platformDecision(fmt.Sprintf("跨产品改动(%s),请拆成每产品一个 PR", strings.Join(dirs, ", ")), platformCleared)
	}

	x := dirs[0]
	owners, baseExists, parseErr := baseOwners(x)
	if parseErr {
		return platformDecision(fmt.Sprintf("base 版 products/%s/product.yaml 解析失败,无法判定归属,需平台介入修复元数据", x), platformCleared)
	}
	if !baseExists {
		return platformDecision(fmt.Sprintf("新产品 onboarding(base 版无 products/%s/)", x), platformCleared)
	}
	if productYAMLDeleted(changed, x) {
		return platformDecision(fmt.Sprintf("下线删除/搬移 products/%s/product.yaml", x), platformCleared)
	}
	for _, o := range owners {
		if strings.EqualFold(o, author) {
			return Decision{
				Type:              "product",
				AutoMergeEligible: true,
				Blocking:          false,
				Reason:            fmt.Sprintf("✅ products/%s 自治:%s 是 base 版 owner,过 CI 后自动合并。", x, author),
			}
		}
	}
	return platformDecision(fmt.Sprintf("%s 不是 products/%s 的 base 版 owner(non-owner 改动)", author, x), platformCleared)
}

func main() {
	author := os.Getenv("OWNER_GATE_AUTHOR")
	baseSHA := os.Getenv("OWNER_GATE_BASE_SHA")
	if author == "" || baseSHA == "" {
		fmt.Fprintln(os.Stderr, "owner-gate: OWNER_GATE_AUTHOR and OWNER_GATE_BASE_SHA must be set")
		os.Exit(2)
	}
	// platformCleared: workflow-computed (author admin OR admin≠author approved).
	// Empty/absent ⇒ not cleared ⇒ platform stays blocked (fail-closed).
	platformCleared := strings.EqualFold(strings.TrimSpace(os.Getenv("OWNER_GATE_PLATFORM_CLEARED")), "true")

	changed, err := parseNameStatus(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "owner-gate: read diff: %v\n", err)
		os.Exit(2)
	}

	d := decide(changed, author, platformCleared, func(product string) ([]string, bool, bool) {
		return gitShowOwners(baseSHA, product)
	})

	enc, _ := json.Marshal(d)
	fmt.Println(string(enc))

	if gho := os.Getenv("GITHUB_OUTPUT"); gho != "" {
		if err := writeGitHubOutput(gho, d); err != nil {
			fmt.Fprintf(os.Stderr, "owner-gate: write GITHUB_OUTPUT: %v\n", err)
			os.Exit(2)
		}
	}
	// 裁决本身从不让 check 失败 —— Blocking 由下游 admission step 消费(exit 1)。
	// 仅内部错误(env 缺失/读 diff 失败/写 output 失败)退非零(exit 2)。
}

// writeGitHubOutput appends type/autoMergeEligible/blocking/reason to the GitHub
// Actions step-output file. reason uses a heredoc to stay multiline-safe.
func writeGitHubOutput(path string, d Decision) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	fmt.Fprintf(f, "type=%s\n", d.Type)
	fmt.Fprintf(f, "autoMergeEligible=%t\n", d.AutoMergeEligible)
	fmt.Fprintf(f, "blocking=%t\n", d.Blocking)
	fmt.Fprintf(f, "reason<<OWNER_GATE_EOF\n%s\nOWNER_GATE_EOF\n", d.Reason)
	return nil
}
