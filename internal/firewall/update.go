package firewall

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
	"unicode/utf16"

	"firewall-manager/internal/admin"
	"firewall-manager/internal/diagnostics"
)

// 输入只通过环境变量传递，精确匹配避免 DisplayName 通配符及批量修改。
const uniqueRuleScript = `$request = $env:FWM_RULE_UPDATE | ConvertFrom-Json
$direction = if ($request.original.direction -eq 'in') { 'Inbound' } else { 'Outbound' }
$matches = @(Get-NetFirewallRule -PolicyStore ActiveStore -TracePolicyStore -ErrorAction Stop | Where-Object {
  $_.Name -ceq $request.original.id
})
if ($matches.Count -eq 0) { throw '原规则不存在，请刷新后重试' }
if ($matches.Count -ne 1) { throw '规则内部标识不唯一，无法唯一识别，已拒绝修改' }
$target = $matches[0]
if ([string]$target.PolicyStoreSourceType -ne 'Local' -or
    [string]$target.PolicyStoreSource -ne 'PersistentStore' -or
    [string]$target.PolicyStoreSource -cne $request.original.policyStoreSource -or
    [string]$target.PolicyStoreSourceType -cne $request.original.policyStoreSourceType -or
    $target.DisplayName -cne $request.original.name -or [string]$target.Direction -ne $direction) {
  throw '规则来源或身份已改变，原规则不是唯一可编辑的本地规则，请刷新后重试'
}
$localMatches = @(Get-NetFirewallRule -PolicyStore PersistentStore -ErrorAction Stop | Where-Object {
  $_.Name -ceq $request.original.id
})
if ($localMatches.Count -ne 1) { throw '原规则不是唯一可编辑的本地规则，请刷新或检查组策略' }
if ($localMatches[0].DisplayName -cne $request.original.name -or [string]$localMatches[0].Direction -ne $direction) {
  throw '本地规则身份已改变，请刷新后重试'
}
`

const updateRuleScript = uniqueRuleScript + `
$params = @{ InputObject = $localMatches[0]; ErrorAction = 'Stop' }
foreach ($property in $request.changes.PSObject.Properties) {
  $value = $property.Value
  switch ($property.Name) {
    'Direction' { $value = if ($value -eq 'in') { 'Inbound' } else { 'Outbound' } }
    'Action' { $value = if ($value -eq 'allow') { 'Allow' } else { 'Block' } }
    'Enabled' { $value = if ($value) { 'True' } else { 'False' } }
    'Profile' { $value = @($value.Split(',')) }
    'LocalAddress' { $value = @($value.Split(',')) }
    'RemoteAddress' { $value = @($value.Split(',')) }
    'LocalPort' { $value = @($value.Split(',')) }
    'RemotePort' { $value = @($value.Split(',')) }
    'IcmpType' { $value = @($value) }
  }
  $params[$property.Name] = $value
}
if ($params.Count -gt 2) { Set-NetFirewallRule @params | Out-Null }`

type updateRequest struct {
	Original FirewallRule           `json:"original"`
	Changes  map[string]interface{} `json:"changes"`
}

// UpdateRule 按内部标识和来源更新唯一可编辑本地规则。
func UpdateRule(original, updated FirewallRule) error {
	return updateRuleWith(original, updated, admin.IsAdmin, runUpdateScript)
}

func updateRuleWith(original, updated FirewallRule, isAdmin func() bool, run func(string, map[string]string) ([]byte, error)) error {
	if err := validateRuleIdentity(original); err != nil {
		return err
	}
	request, err := buildUpdateRequest(original, updated)
	if err != nil {
		return fmt.Errorf("更新规则失败: %w", err)
	}
	if !isAdmin() {
		return fmt.Errorf("更新规则失败: 此操作需要管理员权限")
	}
	payload, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("编码规则更新失败: %w", err)
	}
	output, err := run(updateRuleScript, map[string]string{"FWM_RULE_UPDATE": string(payload)})
	if err != nil {
		return fmt.Errorf("更新规则失败: %w\n%s；请刷新确认当前规则状态", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func buildUpdateRequest(original, updated FirewallRule) (updateRequest, error) {
	if original.ID != updated.ID || original.PolicyStoreSource != updated.PolicyStoreSource || original.PolicyStoreSourceType != updated.PolicyStoreSourceType {
		return updateRequest{}, fmt.Errorf("编辑不能改变规则内部标识或来源")
	}
	before, err := normalizeRuleWith(original, true)
	if err != nil {
		return updateRequest{}, fmt.Errorf("原规则参数不受支持: %w", err)
	}
	after, err := normalizeRuleWith(updated, true)
	if err != nil {
		return updateRequest{}, err
	}
	if before.Name != after.Name {
		return updateRequest{}, fmt.Errorf("编辑不支持更改规则名称")
	}
	changes := make(map[string]interface{})
	pairs := []struct{ key, before, after string }{
		{"Direction", string(before.Direction), string(after.Direction)},
		{"Action", string(before.Action), string(after.Action)},
		{"Program", before.Program, after.Program},
		{"LocalAddress", before.LocalAddr, after.LocalAddr},
		{"RemoteAddress", before.RemoteAddr, after.RemoteAddr},
		{"Protocol", before.Protocol, after.Protocol},
		{"LocalPort", before.LocalPort, after.LocalPort},
		{"RemotePort", before.RemotePort, after.RemotePort},
		{"Profile", before.Profile, after.Profile},
	}
	for _, pair := range pairs {
		if pair.before != pair.after {
			value := pair.after
			if pair.key == "Program" && value == "" {
				value = "Any"
			}
			changes[pair.key] = value
		}
	}
	if before.Enabled != after.Enabled {
		changes["Enabled"] = after.Enabled
	}
	if before.Protocol != after.Protocol && strings.Contains(after.Protocol, ":") {
		parts := strings.SplitN(after.Protocol, ":", 2)
		changes["Protocol"] = parts[0]
		codes := strings.Split(parts[1], ",")
		icmpType := codes[0]
		if codes[1] != "any" {
			icmpType += ":" + codes[1]
		}
		changes["IcmpType"] = icmpType
	} else if before.Protocol != after.Protocol && (after.Protocol == "icmpv4" || after.Protocol == "icmpv6") {
		changes["IcmpType"] = "Any"
	}
	// 切换到无端口协议时同时清除旧端口，避免留下不兼容过滤条件。
	if before.Protocol != after.Protocol && after.Protocol != "tcp" && after.Protocol != "udp" {
		changes["LocalPort"] = "any"
		changes["RemotePort"] = "any"
	}
	// 目标名称保持原始文本，已有规则名中的首尾空格也是身份的一部分。
	identity := before
	identity.Name = original.Name
	return updateRequest{Original: identity, Changes: changes}, nil
}

func runUpdateScript(script string, env map[string]string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	prefix := "$ErrorActionPreference = 'Stop'\n$ProgressPreference = 'SilentlyContinue'\n[Console]::OutputEncoding = New-Object System.Text.UTF8Encoding($false)\n"
	units := utf16.Encode([]rune(prefix + script))
	encoded := make([]byte, len(units)*2)
	for index, unit := range units {
		binary.LittleEndian.PutUint16(encoded[index*2:], unit)
	}
	cmd := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-NonInteractive", "-EncodedCommand", base64.StdEncoding.EncodeToString(encoded))
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Env = append([]string{}, os.Environ()...)
	for key, value := range env {
		cmd.Env = append(cmd.Env, key+"="+value)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if ctx.Err() != nil {
		return nil, fmt.Errorf("规则命令超时: %w", ctx.Err())
	}
	if err != nil {
		return []byte(diagnostics.PowerShell(stderr.Bytes())), err
	}
	return stdout.Bytes(), nil
}
