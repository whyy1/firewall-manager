package smb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Session describes an inbound client session to this SMB server.
type Session struct {
	SessionID          string `json:"sessionId"`
	ClientComputerName string `json:"clientComputerName"`
	ClientUserName     string `json:"clientUserName"`
	NumOpens           uint64 `json:"numOpens"`
	SecondsExists      uint64 `json:"secondsExists"`
	SecondsIdle        uint64 `json:"secondsIdle"`
	Dialect            string `json:"dialect"`
}

// FirewallStatus describes the built-in File and Printer Sharing rule group.
type FirewallStatus struct {
	Available bool                    `json:"available"`
	Profiles  []FirewallProfileStatus `json:"profiles"`
}

// FirewallProfileStatus describes sharing rules for one Windows profile.
type FirewallProfileStatus struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

const sessionsScript = `$items = @(Get-SmbSession -IncludeHidden -ErrorAction Stop | ForEach-Object {
  [PSCustomObject]@{
    sessionId = [string]$_.SessionId
    clientComputerName = [string]$_.ClientComputerName
    clientUserName = [string]$_.ClientUserName
    numOpens = [uint64]$_.NumOpens
    secondsExists = [uint64]$_.SecondsExists
    secondsIdle = [uint64]$_.SecondsIdle
    dialect = [string]$_.Dialect
  }
})
ConvertTo-Json -InputObject $items -Compress`

const closeSessionScript = `$sessionId = [uint64]::Parse($env:FWM_SMB_SESSION_ID)
Close-SmbSession -SessionId $sessionId -Force -Confirm:$false -ErrorAction Stop`

const firewallStatusScript = `$policy = New-Object -ComObject HNetCfg.FwPolicy2
$group = '@FirewallAPI.dll,-28502'
$available = @($policy.Rules | Where-Object { $_.Grouping -eq $group }).Count -gt 0
$profiles = @(
  @{ name = 'domain'; mask = 1 },
  @{ name = 'private'; mask = 2 },
  @{ name = 'public'; mask = 4 }
)
$items = @($profiles | ForEach-Object {
  [PSCustomObject]@{
    name = $_.name
    enabled = $available -and [bool]$policy.IsRuleGroupEnabled([int]$_.mask, $group)
  }
})
[PSCustomObject]@{
  available = $available
  profiles = $items
} | ConvertTo-Json -Compress`

const setFirewallScript = `$policy = New-Object -ComObject HNetCfg.FwPolicy2
$enabled = $env:FWM_SMB_FIREWALL_ENABLED -eq 'true'
$policy.EnableRuleGroup(2147483647, '@FirewallAPI.dll,-28502', $enabled)`

// GetSessions returns inbound SMB sessions to this computer.
func (b backend) GetSessions(ctx context.Context) ([]Session, error) {
	output, err := b.run(ctx, readTimeout, sessionsScript, nil)
	if err != nil {
		return nil, fmt.Errorf("查询 SMB 会话失败: %w", err)
	}
	var result []Session
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("解析 SMB 会话失败: %w", err)
	}
	if result == nil {
		result = []Session{}
	}
	return result, nil
}

// CloseSession forcibly closes one inbound SMB session.
func (b backend) CloseSession(ctx context.Context, sessionID string) error {
	if err := b.requireAdmin(); err != nil {
		return err
	}
	sessionID = strings.TrimSpace(sessionID)
	if err := validateSessionID(sessionID); err != nil {
		return err
	}
	if _, err := b.run(ctx, writeTimeout, closeSessionScript, map[string]string{"FWM_SMB_SESSION_ID": sessionID}); err != nil {
		return fmt.Errorf("断开 SMB 会话失败: %w", err)
	}
	return nil
}

func validateSessionID(sessionID string) error {
	if sessionID == "" || strings.IndexFunc(sessionID, func(char rune) bool { return char < '0' || char > '9' }) >= 0 {
		return errors.New("SMB 会话 ID 无效")
	}
	parsed, err := strconv.ParseUint(sessionID, 10, 64)
	if err != nil || parsed == 0 {
		return errors.New("SMB 会话 ID 无效")
	}
	return nil
}

// GetFirewallStatus returns the built-in sharing firewall group state.
func (b backend) GetFirewallStatus(ctx context.Context) (FirewallStatus, error) {
	var result FirewallStatus
	output, err := b.run(ctx, readTimeout, firewallStatusScript, nil)
	if err != nil {
		return result, fmt.Errorf("查询 SMB 防火墙规则失败: %w", err)
	}
	if err := json.Unmarshal(output, &result); err != nil {
		return result, fmt.Errorf("解析 SMB 防火墙状态失败: %w", err)
	}
	return result, nil
}

// SetFirewallEnabled toggles the built-in File and Printer Sharing group for all profiles.
func (b backend) SetFirewallEnabled(ctx context.Context, enabled bool) error {
	if err := b.requireAdmin(); err != nil {
		return err
	}
	env := map[string]string{
		"FWM_SMB_FIREWALL_ENABLED": strconv.FormatBool(enabled),
	}
	if _, err := b.run(ctx, writeTimeout, setFirewallScript, env); err != nil {
		return fmt.Errorf("设置 SMB 防火墙规则失败: %w", err)
	}
	return nil
}
