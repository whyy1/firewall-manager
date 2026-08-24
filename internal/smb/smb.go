package smb

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"

	"firewall-manager/internal/admin"
)

const (
	readTimeout  = 15 * time.Second
	writeTimeout = 30 * time.Second
)

// Overview describes the local SMB server service and security configuration.
type Overview struct {
	ServiceRunning           bool   `json:"serviceRunning"`
	ServiceStatus            string `json:"serviceStatus"`
	StartupType              string `json:"startupType"`
	SMB1Enabled              bool   `json:"smb1Enabled"`
	SMB2Enabled              bool   `json:"smb2Enabled"`
	RequireSecuritySignature bool   `json:"requireSecuritySignature"`
	EncryptData              bool   `json:"encryptData"`
}

// Share describes a local SMB share.
type Share struct {
	Name                string `json:"name"`
	Path                string `json:"path"`
	NetworkPath         string `json:"networkPath"`
	Description         string `json:"description"`
	Special             bool   `json:"special"`
	State               string `json:"state"`
	ConcurrentUserLimit uint64 `json:"concurrentUserLimit"`
	EncryptData         bool   `json:"encryptData"`
}

// CreateShareRequest contains validated input for a new SMB share.
type CreateShareRequest struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Description string `json:"description"`
	AccessMode  string `json:"accessMode"`
}

var reservedShareNames = map[string]struct{}{
	"admin$": {},
	"ipc$":   {},
	"print$": {},
}

const overviewScript = `$service = Get-Service -Name LanmanServer -ErrorAction Stop
$config = Get-SmbServerConfiguration -ErrorAction Stop
[PSCustomObject]@{
  serviceRunning = [string]$service.Status -eq 'Running'
  serviceStatus = [string]$service.Status
  startupType = [string]$service.StartType
  smb1Enabled = [bool]$config.EnableSMB1Protocol
  smb2Enabled = [bool]$config.EnableSMB2Protocol
  requireSecuritySignature = [bool]$config.RequireSecuritySignature
  encryptData = [bool]$config.EncryptData
} | ConvertTo-Json -Compress`

const sharesScript = `$items = @(Get-SmbShare -IncludeHidden -ErrorAction Stop | ForEach-Object {
  [PSCustomObject]@{
    name = [string]$_.Name
    path = [string]$_.Path
    networkPath = '\\' + $env:COMPUTERNAME + '\' + [string]$_.Name
    description = [string]$_.Description
    special = [bool]$_.Special
    state = [string]$_.ShareState
    concurrentUserLimit = [uint64]$_.ConcurrentUserLimit
    encryptData = [bool]$_.EncryptData
  }
})
ConvertTo-Json -InputObject $items -Compress`

const createShareScript = `$sid = New-Object System.Security.Principal.SecurityIdentifier('S-1-1-0')
$principal = $sid.Translate([System.Security.Principal.NTAccount]).Value
$params = @{
  Name = $env:FWM_SMB_SHARE_NAME
  Path = $env:FWM_SMB_SHARE_PATH
  Description = $env:FWM_SMB_SHARE_DESCRIPTION
  ErrorAction = 'Stop'
}
switch ($env:FWM_SMB_ACCESS_MODE) {
  'read' { $params.ReadAccess = $principal }
  'change' { $params.ChangeAccess = $principal }
  'full' { $params.FullAccess = $principal }
  default { throw 'Unsupported share access mode' }
}
New-SmbShare @params | Out-Null`

const removeShareScript = `$share = Get-SmbShare -Name $env:FWM_SMB_SHARE_NAME -ErrorAction Stop
if ($share.Special) { throw 'System special shares cannot be removed' }
$share | Remove-SmbShare -Force -Confirm:$false -ErrorAction Stop`

// GetOverview returns the local SMB service and protocol state.
func GetOverview(ctx context.Context) (Overview, error) {
	var result Overview
	output, err := runPowerShell(ctx, readTimeout, overviewScript, nil)
	if err != nil {
		return result, fmt.Errorf("查询 SMB 状态失败: %w", err)
	}
	if err := json.Unmarshal(output, &result); err != nil {
		return result, fmt.Errorf("解析 SMB 状态失败: %w", err)
	}
	return result, nil
}

// GetShares returns all local SMB shares, including protected system shares.
func GetShares(ctx context.Context) ([]Share, error) {
	output, err := runPowerShell(ctx, readTimeout, sharesScript, nil)
	if err != nil {
		return nil, fmt.Errorf("查询 SMB 共享失败: %w", err)
	}
	var result []Share
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("解析 SMB 共享失败: %w", err)
	}
	if result == nil {
		result = []Share{}
	}
	return result, nil
}

// SetServiceRunning starts or stops the Windows SMB server service.
func SetServiceRunning(ctx context.Context, running bool) error {
	if err := requireAdmin(); err != nil {
		return err
	}
	script := "Stop-Service -Name LanmanServer -ErrorAction Stop"
	operation := "停止"
	if running {
		script = "Start-Service -Name LanmanServer -ErrorAction Stop"
		operation = "启动"
	}
	if _, err := runPowerShell(ctx, writeTimeout, script, nil); err != nil {
		return fmt.Errorf("%s SMB 服务失败: %w", operation, err)
	}
	return nil
}

// DisableSMB1 disables the deprecated SMB1 server protocol.
func DisableSMB1(ctx context.Context) error {
	if err := requireAdmin(); err != nil {
		return err
	}
	script := "Set-SmbServerConfiguration -EnableSMB1Protocol $false -Force -Confirm:$false -ErrorAction Stop"
	if _, err := runPowerShell(ctx, writeTimeout, script, nil); err != nil {
		return fmt.Errorf("关闭 SMB1 失败: %w", err)
	}
	return nil
}

// CreateShare creates a local share without changing filesystem permissions.
func CreateShare(ctx context.Context, request CreateShareRequest) error {
	if err := requireAdmin(); err != nil {
		return err
	}
	request = normalizeRequest(request)
	if err := validateCreateRequest(request); err != nil {
		return err
	}
	env := map[string]string{
		"FWM_SMB_SHARE_NAME":        request.Name,
		"FWM_SMB_SHARE_PATH":        request.Path,
		"FWM_SMB_SHARE_DESCRIPTION": request.Description,
		"FWM_SMB_ACCESS_MODE":       request.AccessMode,
	}
	if _, err := runPowerShell(ctx, writeTimeout, createShareScript, env); err != nil {
		return fmt.Errorf("创建 SMB 共享失败: %w", err)
	}
	return nil
}

// RemoveShare removes a non-special local SMB share.
func RemoveShare(ctx context.Context, name string) error {
	if err := requireAdmin(); err != nil {
		return err
	}
	name = strings.TrimSpace(name)
	if err := validateShareName(name); err != nil {
		return err
	}
	if isReservedShareName(name) {
		return errors.New("不能删除系统保留共享")
	}
	env := map[string]string{"FWM_SMB_SHARE_NAME": name}
	if _, err := runPowerShell(ctx, writeTimeout, removeShareScript, env); err != nil {
		return fmt.Errorf("删除 SMB 共享失败: %w", err)
	}
	return nil
}

func requireAdmin() error {
	if !admin.IsAdmin() {
		return errors.New("此操作需要管理员权限")
	}
	return nil
}

func normalizeRequest(request CreateShareRequest) CreateShareRequest {
	return CreateShareRequest{
		Name:        strings.TrimSpace(request.Name),
		Path:        filepath.Clean(strings.TrimSpace(request.Path)),
		Description: strings.TrimSpace(request.Description),
		AccessMode:  strings.ToLower(strings.TrimSpace(request.AccessMode)),
	}
}

func validateCreateRequest(request CreateShareRequest) error {
	if err := validateShareName(request.Name); err != nil {
		return err
	}
	if isReservedShareName(request.Name) || strings.HasSuffix(request.Name, "$") {
		return errors.New("共享名属于系统保留名称")
	}
	if !filepath.IsAbs(request.Path) {
		return errors.New("共享目录必须是绝对路径")
	}
	info, err := os.Stat(request.Path)
	if err != nil {
		return fmt.Errorf("共享目录不可用: %w", err)
	}
	if !info.IsDir() {
		return errors.New("共享路径必须指向目录")
	}
	if utf8.RuneCountInString(request.Description) > 256 {
		return errors.New("共享说明不能超过 256 个字符")
	}
	switch request.AccessMode {
	case "read", "change", "full":
		return nil
	default:
		return errors.New("共享权限必须是 read、change 或 full")
	}
}

func validateShareName(name string) error {
	if name == "" || utf8.RuneCountInString(name) > 80 {
		return errors.New("共享名长度必须为 1 到 80 个字符")
	}
	if name != strings.TrimSpace(name) || strings.HasSuffix(name, ".") {
		return errors.New("共享名不能以空格或句点结尾")
	}
	if strings.ContainsAny(name, `"/\[]:|<>+=;,?*`) {
		return errors.New("共享名包含 Windows 不允许的字符")
	}
	for _, char := range name {
		if char < 32 {
			return errors.New("共享名不能包含控制字符")
		}
	}
	return nil
}

func isReservedShareName(name string) bool {
	lower := strings.ToLower(name)
	if _, found := reservedShareNames[lower]; found {
		return true
	}
	return len(lower) == 2 && lower[0] >= 'a' && lower[0] <= 'z' && lower[1] == '$'
}

func runPowerShell(ctx context.Context, timeout time.Duration, script string, extraEnv map[string]string) ([]byte, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	commandCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	prefix := "$ErrorActionPreference = 'Stop'\n$ProgressPreference = 'SilentlyContinue'\n[Console]::OutputEncoding = New-Object System.Text.UTF8Encoding($false)\n"
	encoded := base64.StdEncoding.EncodeToString([]byte(stringsToUTF16LE(prefix + script)))
	cmd := exec.CommandContext(commandCtx, "powershell.exe", "-NoProfile", "-NonInteractive", "-EncodedCommand", encoded)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Env = append([]string{}, os.Environ()...)
	for key, value := range extraEnv {
		cmd.Env = append(cmd.Env, key+"="+value)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if commandCtx.Err() == context.DeadlineExceeded {
		return nil, fmt.Errorf("命令执行超时（%s）", timeout)
	}
	return commandResult(stdout.Bytes(), stderr.Bytes(), err)
}

func commandResult(stdout, stderr []byte, runErr error) ([]byte, error) {
	if runErr == nil {
		return stdout, nil
	}
	messages := make([]string, 0, 2)
	if message := strings.TrimSpace(string(stderr)); message != "" {
		messages = append(messages, message)
	}
	if message := strings.TrimSpace(string(stdout)); message != "" {
		messages = append(messages, message)
	}
	if len(messages) == 0 {
		return nil, runErr
	}
	return nil, fmt.Errorf("%w: %s", runErr, strings.Join(messages, "\n"))
}

func stringsToUTF16LE(value string) string {
	runes := []rune(value)
	encoded := make([]byte, 0, len(runes)*2)
	for _, char := range runes {
		if char <= 0xffff {
			encoded = append(encoded, byte(char), byte(char>>8))
			continue
		}
		char -= 0x10000
		high := uint16(0xd800 + (char >> 10))
		low := uint16(0xdc00 + (char & 0x3ff))
		encoded = append(encoded, byte(high), byte(high>>8), byte(low), byte(low>>8))
	}
	return string(encoded)
}
