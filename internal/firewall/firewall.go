package firewall

import (
	"fmt"
	"os/exec"
	"strings"
)

// GetRules 获取防火墙规则列表
func GetRules(direction RuleDirection) ([]FirewallRule, error) {
	dirStr := "In"
	if direction == Outbound {
		dirStr = "Out"
	}

	cmd := exec.Command("netsh", "advfirewall", "firewall", "show", "rule", "name=all", "dir="+dirStr)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("执行 netsh 失败: %w\n%s", err, string(output))
	}

	return parseRules(string(output), direction)
}

// AddRule 添加防火墙规则
func AddRule(rule FirewallRule) error {
	args := buildAddArgs(rule)
	cmd := exec.Command("netsh", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("添加规则失败: %w\n%s", err, string(output))
	}
	return nil
}

// DeleteRule 删除防火墙规则
func DeleteRule(name string) error {
	cmd := exec.Command("netsh", "advfirewall", "firewall", "delete", "rule", "name="+name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("删除规则失败: %w\n%s", err, string(output))
	}
	return nil
}

// ToggleRule 启用或禁用规则
func ToggleRule(name string, enabled bool) error {
	state := "yes"
	if !enabled {
		state = "no"
	}
	cmd := exec.Command("netsh", "advfirewall", "firewall", "set", "rule",
		"name="+name, "new", "enable="+state)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("切换规则状态失败: %w\n%s", err, string(output))
	}
	return nil
}

// BlockApp 快捷阻止程序联网
func BlockApp(programPath string) error {
	inRule := FirewallRule{
		Name:      fmt.Sprintf("FM_Block_%s", extractFileName(programPath)),
		Direction: Inbound,
		Action:    Block,
		Program:   programPath,
		Protocol:  "any",
		Enabled:   true,
		Profile:   "any",
	}
	if err := AddRule(inRule); err != nil {
		return err
	}

	outRule := inRule
	outRule.Direction = Outbound
	return AddRule(outRule)
}

// AllowApp 快捷放行程序
func AllowApp(programPath string) error {
	ruleName := fmt.Sprintf("FM_Block_%s", extractFileName(programPath))
	return DeleteRule(ruleName)
}

// buildAddArgs 构建 netsh add 命令参数
func buildAddArgs(rule FirewallRule) []string {
	args := []string{
		"advfirewall", "firewall", "add", "rule",
		"name=" + rule.Name,
		"dir=" + string(rule.Direction),
		"action=" + string(rule.Action),
	}

	if rule.Program != "" {
		args = append(args, "program="+rule.Program)
	}
	if rule.Protocol != "" {
		args = append(args, "protocol="+rule.Protocol)
	} else {
		args = append(args, "protocol=any")
	}
	if rule.LocalPort != "" {
		args = append(args, "localport="+rule.LocalPort)
	}
	if rule.RemotePort != "" {
		args = append(args, "remoteport="+rule.RemotePort)
	}
	if rule.Profile != "" {
		args = append(args, "profile="+rule.Profile)
	} else {
		args = append(args, "profile=any")
	}

	enable := "yes"
	if !rule.Enabled {
		enable = "no"
	}
	args = append(args, "enable="+enable)

	return args
}

// parseRules 解析 netsh 输出
func parseRules(output string, direction RuleDirection) ([]FirewallRule, error) {
	var rules []FirewallRule
	var current *FirewallRule

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 新规则开始（以 "Rule Name:" 开头）
		if strings.HasPrefix(line, "Rule Name:") || strings.HasPrefix(line, "规则名称:") {
			if current != nil {
				rules = append(rules, *current)
			}
			current = &FirewallRule{
				Direction: direction,
				Enabled:   true, // 默认启用
				Protocol:  "any",
				Profile:   "any",
			}
			current.Name = extractValue(line)
			continue
		}

		if current == nil {
			continue
		}

		lower := strings.ToLower(line)

		switch {
		case strings.HasPrefix(lower, "enabled:") || strings.HasPrefix(lower, "已启用:"):
			val := strings.TrimSpace(extractValue(line))
			current.Enabled = strings.ToLower(val) == "yes" || val == "是"

		case strings.HasPrefix(lower, "action:") || strings.HasPrefix(lower, "操作:"):
			val := strings.ToLower(strings.TrimSpace(extractValue(line)))
			if val == "block" || val == "阻止" {
				current.Action = Block
			} else {
				current.Action = Allow
			}

		case strings.HasPrefix(lower, "program:") || strings.HasPrefix(lower, "程序:"):
			current.Program = extractValue(line)

		case strings.HasPrefix(lower, "protocol:") || strings.HasPrefix(lower, "协议:"):
			current.Protocol = strings.TrimSpace(extractValue(line))

		case strings.HasPrefix(lower, "localport:") || strings.HasPrefix(lower, "本地端口:"):
			current.LocalPort = strings.TrimSpace(extractValue(line))

		case strings.HasPrefix(lower, "remoteport:") || strings.HasPrefix(lower, "远程端口:"):
			current.RemotePort = strings.TrimSpace(extractValue(line))

		case strings.HasPrefix(lower, "profiles:") || strings.HasPrefix(lower, "配置文件:"):
			current.Profile = strings.TrimSpace(extractValue(line))
		}
	}

	if current != nil {
		rules = append(rules, *current)
	}

	return rules, nil
}

// extractValue 提取冒号后面的值
func extractValue(line string) string {
	idx := strings.Index(line, ":")
	if idx < 0 {
		return ""
	}
	return strings.TrimSpace(line[idx+1:])
}

// extractFileName 从路径中提取文件名
func extractFileName(path string) string {
	parts := strings.ReplaceAll(path, "\\", "/")
	segs := strings.Split(parts, "/")
	if len(segs) == 0 {
		return path
	}
	return segs[len(segs)-1]
}
