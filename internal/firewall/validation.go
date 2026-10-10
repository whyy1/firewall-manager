package firewall

import (
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"strings"
)

func normalizeRule(rule FirewallRule) (FirewallRule, error) {
	return normalizeRuleWith(rule, false)
}

func normalizeRuleWith(rule FirewallRule, extended bool) (FirewallRule, error) {
	result := rule
	result.Name = strings.TrimSpace(rule.Name)
	if result.Name == "" || strings.EqualFold(result.Name, "all") || strings.ContainsAny(result.Name, "\x00\r\n") {
		return result, fmt.Errorf("规则名称不能为空、all 或包含控制字符")
	}
	if result.Direction != Inbound && result.Direction != Outbound {
		return result, fmt.Errorf("规则方向必须为 in 或 out")
	}
	if result.Action != Allow && result.Action != Block {
		return result, fmt.Errorf("规则动作必须为 allow 或 block")
	}
	result.Program = strings.TrimSpace(rule.Program)
	if anyValue(result.Program) {
		result.Program = ""
	}
	if strings.ContainsAny(result.Program, "\x00\r\n") {
		return result, fmt.Errorf("程序路径包含控制字符")
	}
	var err error
	if result.LocalAddr, err = normalizeAddresses(rule.LocalAddr, false); err != nil {
		return result, fmt.Errorf("本地地址无效: %w", err)
	}
	if result.RemoteAddr, err = normalizeAddressesWith(rule.RemoteAddr, true, extended); err != nil {
		return result, fmt.Errorf("远程地址无效: %w", err)
	}
	if result.Profile, err = normalizeProfiles(rule.Profile); err != nil {
		return result, err
	}
	if result.Protocol, err = normalizeProtocol(rule.Protocol); err != nil {
		return result, err
	}
	if result.LocalPort, err = normalizePorts(rule.LocalPort, true); err != nil {
		return result, fmt.Errorf("本地端口无效: %w", err)
	}
	if result.RemotePort, err = normalizePorts(rule.RemotePort, false); err != nil {
		return result, fmt.Errorf("远程端口无效: %w", err)
	}
	if result.Protocol != "tcp" && result.Protocol != "udp" && (result.LocalPort != "any" || result.RemotePort != "any") {
		return result, fmt.Errorf("只有 TCP 或 UDP 协议可以指定端口")
	}
	if result.Protocol != "tcp" && (hasSpecialPort(result.LocalPort) || hasSpecialPort(result.RemotePort)) {
		return result, fmt.Errorf("RPC、RPC-EPMap 和 IPHTTPS 端口要求 TCP 协议")
	}
	return result, nil
}

func anyValue(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "any", "任意", "任何", "所有":
		return true
	default:
		return false
	}
}

func normalizeAddresses(value string, remote bool) (string, error) {
	return normalizeAddressesWith(value, remote, false)
}

func normalizeAddressesWith(value string, remote, extended bool) (string, error) {
	if anyValue(value) {
		return "any", nil
	}
	parts := strings.Split(strings.ReplaceAll(value, "，", ","), ",")
	for index, part := range parts {
		part = strings.TrimSpace(part)
		keyword := strings.ToLower(part)
		switch keyword {
		case "本地子网", "本地子网地址":
			keyword = "localsubnet"
		case "默认网关":
			keyword = "defaultgateway"
		}
		switch keyword {
		case "localsubnet", "dns", "dhcp", "wins", "defaultgateway":
			if !remote {
				return "", fmt.Errorf("%q 仅适用于远程地址", part)
			}
			parts[index] = keyword
			continue
		}
		if remote && extended && extendedAddressKeyword(keyword) {
			parts[index] = keyword
			continue
		}
		if anyValue(part) {
			return "", fmt.Errorf("任意地址不能与其他地址混合")
		}
		if address, err := netip.ParseAddr(part); err == nil && address.Zone() == "" {
			parts[index] = part
			continue
		}
		if _, err := netip.ParsePrefix(part); err == nil {
			parts[index] = part
			continue
		}
		if subnet, ok := dottedSubnet(part); ok {
			parts[index] = subnet
			continue
		}
		bounds := strings.Split(part, "-")
		if len(bounds) == 2 {
			first, firstErr := netip.ParseAddr(strings.TrimSpace(bounds[0]))
			last, lastErr := netip.ParseAddr(strings.TrimSpace(bounds[1]))
			if firstErr == nil && lastErr == nil && first.Zone() == "" && last.Zone() == "" && first.BitLen() == last.BitLen() && first.Compare(last) <= 0 {
				parts[index] = first.String() + "-" + last.String()
				continue
			}
		}
		return "", fmt.Errorf("%q 不是有效 IP、子网、范围或地址关键字", part)
	}
	return strings.Join(parts, ","), nil
}

func dottedSubnet(value string) (string, bool) {
	parts := strings.Split(value, "/")
	if len(parts) != 2 {
		return "", false
	}
	address, err := netip.ParseAddr(parts[0])
	mask := net.ParseIP(parts[1]).To4()
	if err != nil || !address.Is4() || mask == nil {
		return "", false
	}
	ones, bits := net.IPMask(mask).Size()
	if bits != 32 {
		return "", false
	}
	return address.String() + "/" + strconv.Itoa(ones), true
}

func normalizeProfiles(value string) (string, error) {
	if anyValue(value) {
		return "any", nil
	}
	parts := strings.Split(strings.ReplaceAll(value, "，", ","), ",")
	for index, part := range parts {
		switch strings.ToLower(strings.TrimSpace(part)) {
		case "domain", "域":
			parts[index] = "domain"
		case "private", "专用":
			parts[index] = "private"
		case "public", "公用":
			parts[index] = "public"
		default:
			return "", fmt.Errorf("配置文件无效: %q", part)
		}
	}
	return strings.Join(parts, ","), nil
}

func normalizeProtocol(value string) (string, error) {
	if anyValue(value) {
		return "any", nil
	}
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "tcp", "6":
		return "tcp", nil
	case "udp", "17":
		return "udp", nil
	case "icmpv4", "icmpv6":
		return value, nil
	}
	if numericRange(value, 255) {
		return value, nil
	}
	parts := strings.Split(value, ":")
	if len(parts) == 2 && (parts[0] == "icmpv4" || parts[0] == "icmpv6") {
		codes := strings.Split(parts[1], ",")
		if len(codes) == 2 && (codes[0] == "any" || numericRange(codes[0], 255)) && (codes[1] == "any" || numericRange(codes[1], 255)) {
			return value, nil
		}
	}
	return "", fmt.Errorf("协议无效: %q", value)
}

func numericRange(value string, max int) bool {
	if value == "" {
		return false
	}
	for _, char := range value {
		if char < '0' || char > '9' {
			return false
		}
	}
	number, err := strconv.Atoi(value)
	return err == nil && number >= 0 && number <= max
}

func normalizePorts(value string, local bool) (string, error) {
	if anyValue(value) {
		return "any", nil
	}
	parts := strings.Split(strings.ReplaceAll(value, "，", ","), ",")
	for index, part := range parts {
		part = strings.TrimSpace(part)
		if local {
			switch strings.ToLower(part) {
			case "rpc":
				parts[index] = "RPC"
				continue
			case "rpc-epmap":
				parts[index] = "RPC-EPMap"
				continue
			case "iphttps":
				parts[index] = "IPHTTPS"
				continue
			}
		}
		if numericRange(part, 65535) {
			parts[index] = part
			continue
		}
		bounds := strings.Split(part, "-")
		if len(bounds) == 2 && numericRange(bounds[0], 65535) && numericRange(bounds[1], 65535) {
			first, _ := strconv.Atoi(bounds[0])
			last, _ := strconv.Atoi(bounds[1])
			if first <= last {
				parts[index] = part
				continue
			}
		}
		return "", fmt.Errorf("%q 不是有效端口或范围", part)
	}
	return strings.Join(parts, ","), nil
}

func hasSpecialPort(value string) bool {
	return strings.ContainsAny(value, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
}

// Extended keywords are supported by Set-NetFirewallRule, not assumed for netsh add.
func extendedAddressKeyword(value string) bool {
	base := value
	if strings.HasSuffix(value, "4") || strings.HasSuffix(value, "6") {
		base = value[:len(value)-1]
	}
	switch base {
	case "any", "localsubnet", "dns", "dhcp", "wins", "defaultgateway", "internet", "intranet", "intranetremoteaccess", "playtodevice", "captiveportal":
		return value != "any"
	default:
		return false
	}
}
