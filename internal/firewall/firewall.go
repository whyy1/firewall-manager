package firewall

import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"unicode/utf8"
	"unsafe"
)

// newHiddenCmd 创建隐藏控制台窗口的命令
func newHiddenCmd(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd
}

// gbkToUTF8 使用 Windows API 将 GBK 字节转为 UTF-8
func gbkToUTF8(data []byte) string {
	if utf8.Valid(data) {
		return string(data)
	}

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	MultiByteToWideChar := kernel32.NewProc("MultiByteToWideChar")
	WideCharToMultiByte := kernel32.NewProc("WideCharToMultiByte")

	// GBK (CP936) → UTF-16
	cp := uint32(936) // GBK
	srcLen := len(data)
	if srcLen == 0 {
		return ""
	}

	// 第一次调用：获取需要的 wchar 数量
	wcharLen, _, _ := MultiByteToWideChar.Call(
		uintptr(cp), 0,
		uintptr(unsafe.Pointer(&data[0])), uintptr(srcLen),
		0, 0,
	)
	if wcharLen == 0 {
		return string(data)
	}

	// 分配 UTF-16 缓冲区
	wcharBuf := make([]uint16, wcharLen)
	MultiByteToWideChar.Call(
		uintptr(cp), 0,
		uintptr(unsafe.Pointer(&data[0])), uintptr(srcLen),
		uintptr(unsafe.Pointer(&wcharBuf[0])), uintptr(wcharLen),
	)

	// UTF-16 → UTF-8
	utf8Len, _, _ := WideCharToMultiByte.Call(
		65001, 0, // CP_UTF8
		uintptr(unsafe.Pointer(&wcharBuf[0])), uintptr(wcharLen),
		0, 0, 0, 0,
	)
	if utf8Len == 0 {
		return string(data)
	}

	utf8Buf := make([]byte, utf8Len)
	WideCharToMultiByte.Call(
		65001, 0,
		uintptr(unsafe.Pointer(&wcharBuf[0])), uintptr(wcharLen),
		uintptr(unsafe.Pointer(&utf8Buf[0])), uintptr(utf8Len),
		0, 0,
	)

	return string(utf8Buf)
}

// GetRules 获取防火墙规则列表
func GetRules(direction RuleDirection, limit int) ([]FirewallRule, error) {
	dirStr := "In"
	if direction == Outbound {
		dirStr = "Out"
	}

	cmd := newHiddenCmd("netsh", "advfirewall", "firewall", "show", "rule", "name=all", "dir="+dirStr)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("执行 netsh 失败: %w\n%s", err, string(output))
	}

	// GBK → UTF-8
	utf8Output := gbkToUTF8(output)

	rules, err := parseRules(utf8Output, direction)
	if err != nil {
		return nil, err
	}

	if limit > 0 && len(rules) > limit {
		rules = rules[:limit]
	}

	return rules, nil
}

// AddRule 添加防火墙规则
func AddRule(rule FirewallRule) error {
	args := buildAddArgs(rule)
	cmd := newHiddenCmd("netsh", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("添加规则失败: %w\n%s", err, string(output))
	}
	return nil
}

// DeleteRule 删除防火墙规则
func DeleteRule(name string) error {
	cmd := newHiddenCmd("netsh", "advfirewall", "firewall", "delete", "rule", "name="+name)
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
	cmd := newHiddenCmd("netsh", "advfirewall", "firewall", "set", "rule",
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

// parseRules 解析 netsh 输出（已转为 UTF-8）
func parseRules(output string, direction RuleDirection) ([]FirewallRule, error) {
	var rules []FirewallRule
	var current *FirewallRule

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if isNewRule(line) {
			if current != nil {
				rules = append(rules, *current)
			}
			current = &FirewallRule{
				Direction: direction,
				Enabled:   true,
				Protocol:  "any",
				Profile:   "any",
			}
			current.Name = extractValue(line)
			continue
		}

		if current == nil {
			continue
		}

		key := normalizeKey(line)

		switch {
		case hasPrefixAny(key, "enabled:", "已启用:"):
			val := strings.TrimSpace(extractValue(line))
			current.Enabled = strings.ToLower(val) == "yes" || val == "是"

		case hasPrefixAny(key, "action:", "操作:"):
			val := strings.ToLower(strings.TrimSpace(extractValue(line)))
			if val == "block" || val == "阻止" {
				current.Action = Block
			} else {
				current.Action = Allow
			}

		case hasPrefixAny(key, "program:", "程序:"):
			current.Program = extractValue(line)

		case hasPrefixAny(key, "localip:", "本地ip:"):
			current.LocalAddr = strings.TrimSpace(extractValue(line))

		case hasPrefixAny(key, "remoteip:", "远程ip:"):
			current.RemoteAddr = strings.TrimSpace(extractValue(line))

		case hasPrefixAny(key, "protocol:", "协议:"):
			current.Protocol = strings.TrimSpace(extractValue(line))

		case hasPrefixAny(key, "localport:", "本地端口:"):
			current.LocalPort = strings.TrimSpace(extractValue(line))

		case hasPrefixAny(key, "remoteport:", "远程端口:"):
			current.RemotePort = strings.TrimSpace(extractValue(line))

		case hasPrefixAny(key, "profiles:", "配置文件:"):
			current.Profile = strings.TrimSpace(extractValue(line))
		}
	}

	if current != nil {
		rules = append(rules, *current)
	}

	return rules, nil
}

func isNewRule(line string) bool {
	return strings.HasPrefix(line, "Rule Name:") ||
		strings.HasPrefix(line, "规则名称:")
}

func normalizeKey(line string) string {
	// 去掉全角空格 　 和 ASCII 空格
	cleaned := strings.ReplaceAll(line, "　", "")
	cleaned = strings.TrimSpace(cleaned)
	lower := strings.ToLower(cleaned)
	idx := strings.Index(lower, ":")
	if idx < 0 {
		return lower
	}
	// 去掉冒号前所有空格
	return strings.ReplaceAll(lower[:idx], " ", "") + lower[idx:]
}

func hasPrefixAny(s string, prefixes ...string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}

func extractValue(line string) string {
	// 先尝试 ASCII 冒号
	idx := strings.Index(line, ":")
	if idx >= 0 {
		return strings.TrimSpace(line[idx+1:])
	}
	// 再尝试全角冒号
	idx = strings.Index(line, "：")
	if idx >= 0 {
		return strings.TrimSpace(line[idx+3:]) // "：" 占 3 字节
	}
	return ""
}

func extractFileName(path string) string {
	parts := strings.ReplaceAll(path, "\\", "/")
	segs := strings.Split(parts, "/")
	if len(segs) == 0 {
		return path
	}
	return segs[len(segs)-1]
}
