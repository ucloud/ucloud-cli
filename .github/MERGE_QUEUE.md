# Merge Queue — Setup & Process

This document describes the GitHub repository settings required to enable Merge Queue on `master`, and the engineering process for protected-path changes (new/modified products, platform config).

---

## Required Repository Settings Checklist

### 1. General

- [ ] **Settings → General → Allow auto-merge** — enable so PRs can be set to auto-merge and enter the queue automatically once approved.

### 2. Branch Protection / Ruleset on `master`

Navigate to **Settings → Branches → Add rule** (classic branch protection) or **Settings → Rules → Rulesets** (newer UI):

- [ ] **Require merge queue** — enables the GitHub Merge Queue. PRs cannot be merged directly; they must enter the queue.
- [ ] **Require status checks to pass before merging** — add ALL of the following required checks:
  - `scope-gate` (from `pr-gate.yml`)
  - `check-product` (from `pr-gate.yml`)
  - `test` (from `pr-gate.yml`)
  - `build-matrix` (6× platform matrix, from `pr-gate.yml`)
  - `build` (from `build.yml`)
- [ ] **Require branches to be up to date before merging** — the merge queue handles this automatically, but must be enabled.
- [ ] **Include administrators** — prevents bypass by repo admins.

> **Why `merge_group` trigger matters:** Every workflow listed as a required check must trigger on the `merge_group` event. Without it, the checks remain in "Pending" state inside the queue and PRs can never be merged. Both `pr-gate.yml` and `build.yml` now include `merge_group:` in their `on:` triggers.

---

## Merge Queue Availability Caveat

GitHub Merge Queue requires one of:

- **Public repository** on any plan (free tier included).
- **GitHub Enterprise Cloud (GHEC)** for private repositories.

If Merge Queue is unavailable for this repo's plan/visibility, use the fallback:

- Enable **"Require branches to be up to date"** + all required status checks (same list above).
- Merge PRs serially — only merge one at a time, waiting for CI to pass on each before merging the next.
- This avoids broken `master` from parallel merges but reduces velocity.

---

## New / Enable / Disable a Product — Platform-Serial PR Process

Changes to `products.yaml` and the regenerated artifacts (`cmd/products.gen.go`, golden test fixtures, generated docs) are **protected paths** and follow a separate, stricter process:

### Why a separate process?

`products.yaml` is the source of truth for all product scaffolding. Changes to it affect:

- `cmd/products.gen.go` — regenerated via `go run ./hack/gen-products`
- Golden test fixtures under `products/<name>/testdata/`
- Any auto-generated documentation

Hand-editing regenerated files in a product PR will cause the `check-product` gate to fail.

### Rules

1. **Product-local PRs** (feature work inside `products/<name>/`) MUST NOT touch:
   - `products.yaml`
   - `cmd/products.gen.go`
   - `hack/`
   - `.github/workflows/`
   - `go.mod` / `go.sum`
   - `.goreleaser.yaml` / `.svu.yaml`

   The `scope-gate` job enforces this automatically.

2. **Platform-serial PR** (one at a time, through platform review) is required for:
   - Adding a new product entry to `products.yaml`
   - Enabling or disabling an existing product
   - Any change to `hack/`, `go.mod`, or CI workflows

3. **Regeneration is mandatory** — after editing `products.yaml`, run:
   ```sh
   go run ./hack/gen-products
   ```
   Commit both `products.yaml` and all regenerated artifacts in the same platform PR.

4. **No hand-edits to generated files** — `cmd/products.gen.go` and golden fixtures are outputs, not inputs. The `check-product` job verifies this and will fail if they drift from `products.yaml`.

---

## `release.yml` Manual Approval Gate

The `release.yml` workflow's first run is guarded by a **GitHub Environment** named `release` that requires manual approval before the release job executes. This prevents accidental releases during initial setup.

After one verified release has shipped successfully, the manual approval requirement can be removed from the `release` environment (Settings → Environments → release → Protection rules).

Cross-reference: Task E8 (release workflow setup).

---

## Reference: `gh api` Commands (for humans, do not run in CI)

The following commands configure the branch protection rule via the GitHub API. Run manually by a repo admin after confirming settings in the UI.

```sh
# Replace ORG/REPO with the actual values
ORG=ucloud
REPO=ucloud-cli

# Enable branch protection on master with required status checks and merge queue
gh api --method PUT \
  repos/$ORG/$REPO/branches/master/protection \
  --input - <<'EOF'
{
  "required_status_checks": {
    "strict": true,
    "checks": [
      {"context": "scope-gate"},
      {"context": "check-product"},
      {"context": "test"},
      {"context": "build-matrix"},
      {"context": "build"}
    ]
  },
  "enforce_admins": true,
  "required_pull_request_reviews": null,
  "restrictions": null
}
EOF

# Enable merge queue (requires GHEC or public repo; uses GraphQL Rulesets API)
# For public repos this is done via UI: Settings → Branches → Edit rule → "Require merge queue"
```

> Note: The Merge Queue option in branch protection is not fully exposed in the REST API for all plan tiers. Use the Settings UI as the primary method; the commands above are a partial reference.
