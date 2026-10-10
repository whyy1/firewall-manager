package firewall

import (
	"encoding/json"
	"fmt"
	"strings"

	"firewall-manager/internal/admin"
)

const deleteRuleScript = uniqueRuleScript + `Remove-NetFirewallRule -InputObject $localMatches[0] -ErrorAction Stop | Out-Null`
const toggleRuleScript = uniqueRuleScript + `$enabled = if ($request.enabled) { 'True' } else { 'False' }
Set-NetFirewallRule -InputObject $localMatches[0] -Enabled $enabled -ErrorAction Stop | Out-Null`

type operationRequest struct {
	Original FirewallRule `json:"original"`
	Enabled  bool         `json:"enabled"`
}

// DeleteRule 按内部标识和来源删除唯一匹配的本地规则。
func DeleteRule(rule FirewallRule) error {
	return operateRuleWith(rule, false, "删除规则", deleteRuleScript, admin.IsAdmin, runUpdateScript)
}

// ToggleRule 仅切换所选的唯一可编辑本地规则。
func ToggleRule(rule FirewallRule, enabled bool) error {
	return operateRuleWith(rule, enabled, "切换规则状态", toggleRuleScript, admin.IsAdmin, runUpdateScript)
}

func operateRuleWith(rule FirewallRule, enabled bool, label, script string, isAdmin func() bool, run func(string, map[string]string) ([]byte, error)) error {
	if err := validateRuleIdentity(rule); err != nil {
		return err
	}
	if strings.TrimSpace(rule.Name) == "" || strings.ContainsAny(rule.Name, "\x00\r\n") {
		return fmt.Errorf("%s失败: 规则名称无效", label)
	}
	if rule.Direction != Inbound && rule.Direction != Outbound {
		return fmt.Errorf("%s失败: 规则方向无效", label)
	}
	if !isAdmin() {
		return fmt.Errorf("%s失败: 此操作需要管理员权限", label)
	}
	payload, err := json.Marshal(operationRequest{Original: rule, Enabled: enabled})
	if err != nil {
		return fmt.Errorf("编码规则操作失败: %w", err)
	}
	output, err := run(script, map[string]string{"FWM_RULE_UPDATE": string(payload)})
	if err != nil {
		return fmt.Errorf("%s失败: %w\n%s；请刷新确认当前规则状态", label, err, strings.TrimSpace(string(output)))
	}
	return nil
}

func validateRuleIdentity(rule FirewallRule) error {
	if strings.TrimSpace(rule.ID) == "" || strings.ContainsAny(rule.ID, "\x00\r\n") {
		return fmt.Errorf("规则内部标识缺失或无效，请刷新后重试")
	}
	if !strings.EqualFold(rule.PolicyStoreSourceType, "Local") || !strings.EqualFold(rule.PolicyStoreSource, "PersistentStore") || strings.ContainsAny(rule.PolicyStoreSource+rule.PolicyStoreSourceType, "\x00\r\n") {
		return fmt.Errorf("规则来源缺失或不是可编辑的本地规则，请刷新后重试")
	}
	if !rule.CanEdit {
		return fmt.Errorf("规则当前只读，请刷新后检查规则来源和参数")
	}
	return nil
}
