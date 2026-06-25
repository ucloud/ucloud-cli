// hack/owner-gate decides whether a pull request can be auto-merged by a
// product owner (product-autonomous PR) or must go through platform review
// (platform PR). See docs/ROADMAP.md P2a.
//
// Ownership is ALWAYS judged against the BASE revision of
// products/X/product.yaml — a PR cannot grant itself ownership by adding the
// author to owners in the same PR.
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
	Type              string `json:"type"`              // "product" | "platform"
	AutoMergeEligible bool   `json:"autoMergeEligible"` // true only for product PRs
	Reason            string `json:"reason"`            // PR-visible explanation
}

// changedFile is one entry from `git diff --name-status`.
type changedFile struct {
	Path    string
	Deleted bool
}

// baseOwnersFunc returns the owners of products/<product>/product.yaml at the
// BASE revision, and whether that file existed at base.
type baseOwnersFunc func(product string) (owners []string, baseExists bool)

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

// productYAMLDeleted reports whether this PR deletes products/<x>/product.yaml.
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
// "<status>\t<path>"; renames/copies are "<status>\t<old>\t<new>" — we take the
// final path (the one present at HEAD). A status starting with 'D' marks a
// deletion.
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
		path := fields[len(fields)-1]
		out = append(out, changedFile{Path: path, Deleted: strings.HasPrefix(status, "D")})
	}
	return out, sc.Err()
}

// gitShowOwners reads products/<product>/product.yaml at baseSHA and returns
// its owners. baseExists is false when the file does not exist at that revision
// (new product). Runs in the process working directory (the repo).
func gitShowOwners(baseSHA, product string) ([]string, bool) {
	ref := fmt.Sprintf("%s:products/%s/product.yaml", baseSHA, product)
	raw, err := exec.Command("git", "show", ref).Output()
	if err != nil {
		// Non-zero exit ⇒ path absent at base ⇒ new product.
		return nil, false
	}
	var meta struct {
		Owners []string `yaml:"owners"`
	}
	if err := yaml.Unmarshal(raw, &meta); err != nil {
		return nil, true // exists but unparseable: existing, no owners
	}
	return meta.Owners, true
}

// decide is the pure admission decision (see plan / spec §①).
func decide(changed []changedFile, author string, baseOwners baseOwnersFunc) Decision {
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
		return Decision{
			Type:   "platform",
			Reason: fmt.Sprintf("⚠️ 平台 PR:改动触及平台/受保护路径(%s),需平台审批。", strings.Join(platformFiles, ", ")),
		}
	}

	dirs := sortedKeys(productSet)
	switch len(dirs) {
	case 0:
		return Decision{Type: "platform", Reason: "⚠️ 平台 PR:无产品目录改动。"}
	case 1:
		// handled below
	default:
		return Decision{
			Type:   "platform",
			Reason: fmt.Sprintf("⚠️ 平台 PR:跨产品改动(%s),请拆成每产品一个 PR 或走平台审批。", strings.Join(dirs, ", ")),
		}
	}

	x := dirs[0]
	owners, baseExists := baseOwners(x)
	if !baseExists {
		return Decision{Type: "platform", Reason: fmt.Sprintf("⚠️ 平台 PR:新产品 onboarding(base 版无 products/%s/),需平台审批。", x)}
	}
	if productYAMLDeleted(changed, x) {
		return Decision{Type: "platform", Reason: fmt.Sprintf("⚠️ 平台 PR:下线删除 products/%s/product.yaml,需平台审批。", x)}
	}
	for _, o := range owners {
		if strings.EqualFold(o, author) {
			return Decision{
				Type:              "product",
				AutoMergeEligible: true,
				Reason:            fmt.Sprintf("✅ products/%s 自治:%s 是 base 版 owner,过 CI 后自动合并。", x, author),
			}
		}
	}
	return Decision{
		Type:   "platform",
		Reason: fmt.Sprintf("⚠️ 平台 PR:%s 不是 products/%s 的 base 版 owner(non-owner 改动),需平台审批。", author, x),
	}
}

func main() {
	author := os.Getenv("OWNER_GATE_AUTHOR")
	baseSHA := os.Getenv("OWNER_GATE_BASE_SHA")
	if author == "" || baseSHA == "" {
		fmt.Fprintln(os.Stderr, "owner-gate: OWNER_GATE_AUTHOR and OWNER_GATE_BASE_SHA must be set")
		os.Exit(2)
	}

	changed, err := parseNameStatus(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "owner-gate: read diff: %v\n", err)
		os.Exit(2)
	}

	d := decide(changed, author, func(product string) ([]string, bool) {
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
	// 裁决本身从不让 check 失败 —— eligibility 由下游 auto-merge job 消费。
	// 仅上面的内部错误(env 缺失/读 diff 失败/写 output 失败)退非零。
}

// writeGitHubOutput appends type/autoMergeEligible/reason to the GitHub Actions
// step-output file. reason uses a heredoc to stay multiline-safe.
func writeGitHubOutput(path string, d Decision) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	fmt.Fprintf(f, "type=%s\n", d.Type)
	fmt.Fprintf(f, "autoMergeEligible=%t\n", d.AutoMergeEligible)
	fmt.Fprintf(f, "reason<<OWNER_GATE_EOF\n%s\nOWNER_GATE_EOF\n", d.Reason)
	return nil
}
