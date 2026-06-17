#!/usr/bin/env bash
# tests/oauth_cli_matrix.sh — CLI 环境矩阵（D8）：非 TTY / 无浏览器 / stdin pipe / init↔login 共存 / profile 切换
# 用法：bash tests/oauth_cli_matrix.sh
# 黑盒驱动构建产物，用独立 HOME 沙箱，不触碰真实 ~/.ucloud
set -u

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BIN="$ROOT/out/ucloud-matrix-test"

# 先用真实 HOME 构建（asdf/版本管理器的 go shim 依赖 $HOME 解析工具链），再切沙箱 HOME
go build -mod=vendor -o "$BIN" "$ROOT/main.go" || { echo "build failed"; exit 1; }

SANDBOX="$(mktemp -d)"
export HOME="$SANDBOX"
PASS=0; FAIL=0

check() { # check <desc> <expected_exit> <actual_exit> <output> <must_contain>
  local desc="$1" want="$2" got="$3" out="$4" needle="$5"
  if [ "$got" = "$want" ] && echo "$out" | grep -q "$needle"; then
    PASS=$((PASS+1)); echo "[OK]   $desc"
  else
    FAIL=$((FAIL+1)); echo "[FAIL] $desc (exit=$got want=$want; output: $out)"
  fi
}

# 1. 非 TTY login fail-fast：stderr + 非零退出
ERR=$(echo "" | "$BIN" auth login 2>&1 >/dev/null); RC=$?
check "non-tty login fail-fast to stderr" 1 "$RC" "$ERR" "interactive terminal"

# 2. stdin pipe 跑业务命令（无任何配置）：aksk 路径既有提示零回归（注意：历史行为 exit 0，保持）
OUT=$("$BIN" region 2>&1 </dev/null); RC=$?
check "aksk missing-key prompt unchanged (regression)" 0 "$RC" "$OUT" "private-key is empty"

# 3. oauth profile 且 token 缺失：stderr + 非零退出 + 指向 login（TTY）/ AK-SK（非 TTY）
mkdir -p "$HOME/.ucloud"
cat > "$HOME/.ucloud/config.json" <<'EOF'
[{"project_id":"org-x","region":"cn-bj2","zone":"cn-bj2-04","base_url":"https://api.ucloud.cn/","timeout_sec":15,"profile":"default","active":true,"max_retry_times":3}]
EOF
cat > "$HOME/.ucloud/credential.json" <<'EOF'
[{"public_key":"","private_key":"","cookie":"","csrf_token":"","profile":"default","auth_mode":"oauth"}]
EOF
# 注意：非 TTY 用 pipe 模拟而非 </dev/null —— /dev/null 是字符设备，会被 IsStdinTTY 误判为 TTY
ERR=$(echo "" | "$BIN" region 2>&1 >/dev/null); RC=$?
check "oauth missing token: stderr + nonzero + AK/SK pointer (non-tty)" 1 "$RC" "$ERR" "AK/SK"

# 4. logout 在未登录 profile 上幂等
cat > "$HOME/.ucloud/credential.json" <<'EOF'
[{"public_key":"pub","private_key":"pri","cookie":"","csrf_token":"","profile":"default"}]
EOF
OUT=$("$BIN" auth logout 2>&1); RC=$?
check "logout on non-oauth profile is a no-op" 0 "$RC" "$OUT" "not logged in"

# 5. profile 切换：oauth profile + aksk profile 并存，--profile 选中 aksk 的不受 oauth 影响
cat > "$HOME/.ucloud/config.json" <<'EOF'
[{"project_id":"org-x","region":"cn-bj2","zone":"cn-bj2-04","base_url":"https://api.ucloud.cn/","timeout_sec":15,"profile":"oa","active":true,"max_retry_times":3},
 {"project_id":"org-y","region":"cn-bj2","zone":"cn-bj2-04","base_url":"https://api.ucloud.cn/","timeout_sec":15,"profile":"ak","active":false,"max_retry_times":3}]
EOF
cat > "$HOME/.ucloud/credential.json" <<'EOF'
[{"public_key":"","private_key":"","cookie":"","csrf_token":"","profile":"oa","auth_mode":"oauth"},
 {"public_key":"pub","private_key":"pri","cookie":"","csrf_token":"","profile":"ak"}]
EOF
OUT=$("$BIN" config list 2>&1); RC=$?
check "config list shows AuthMode column" 0 "$RC" "$OUT" "AuthMode"
ERR=$(echo "" | "$BIN" region --profile oa 2>&1 >/dev/null); RC=$?
check "profile switch: oauth profile without token fails nonzero" 1 "$RC" "$ERR" "Profile 'oa'"

# 6. --no-browser flag 存在（help 可见；真实流程属手动 E2E）
OUT=$("$BIN" auth login --help 2>&1); RC=$?
check "login --help mentions --no-browser" 0 "$RC" "$OUT" "no-browser"

# 7. init 在 oauth profile（已存 AK/SK）上确认 y：auth_mode/token 必须清除并落盘
#    base_url/oauth_base_url 指向不可达地址，printHello/refresh 立刻失败，不出外网；只断言盘上状态
cat > "$HOME/.ucloud/config.json" <<'EOF'
[{"project_id":"org-x","region":"cn-bj2","zone":"cn-bj2-04","base_url":"http://127.0.0.1:1/","oauth_base_url":"http://127.0.0.1:1/","timeout_sec":15,"profile":"default","active":true,"max_retry_times":0}]
EOF
cat > "$HOME/.ucloud/credential.json" <<'EOF'
[{"public_key":"pub","private_key":"pri","cookie":"","csrf_token":"","profile":"default","auth_mode":"oauth","access_token":"at","refresh_token":"rt","expires_at":123}]
EOF
printf 'y\n' | "$BIN" init >/dev/null 2>&1
if ! grep -q '"auth_mode"' "$HOME/.ucloud/credential.json" && grep -q '"public_key": *"pub"' "$HOME/.ucloud/credential.json"; then
  PASS=$((PASS+1)); echo "[OK]   init on oauth profile with AK/SK: confirm y clears auth_mode on disk"
else
  FAIL=$((FAIL+1)); echo "[FAIL] init on oauth profile with AK/SK: confirm y clears auth_mode on disk (credential: $(cat "$HOME/.ucloud/credential.json"))"
fi

echo ""
echo "matrix result: $PASS passed, $FAIL failed"
rm -rf "$SANDBOX" "$BIN"
[ "$FAIL" -eq 0 ]
