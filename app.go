package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"syscall"

	"firewall-manager/internal/admin"
	"firewall-manager/internal/firewall"
	"firewall-manager/internal/ports"
	"firewall-manager/internal/smb"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App 应用主结构体，方法会被绑定到前端
type App struct {
	ctx context.Context
}

// NewApp 创建 App 实例
func NewApp() *App {
	return &App{}
}

// startup Wails 启动回调
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// IsAdmin 检查是否以管理员权限运行
func (a *App) IsAdmin() bool {
	return admin.IsAdmin()
}

// GetRules 获取指定方向的防火墙规则（limit=0 获取全部）
func (a *App) GetRules(direction string, limit int) []firewall.FirewallRule {
	dir := firewall.Inbound
	if direction == "out" {
		dir = firewall.Outbound
	}
	rules, err := firewall.GetRules(dir, limit)
	if err != nil {
		fmt.Printf("获取规则失败: %v\n", err)
		return []firewall.FirewallRule{}
	}
	return rules
}

// AddRule 添加防火墙规则
func (a *App) AddRule(rule firewall.FirewallRule) error {
	return firewall.AddRule(rule)
}

// DeleteRule 删除防火墙规则
func (a *App) DeleteRule(name string) error {
	return firewall.DeleteRule(name)
}

// ToggleRule 启用/禁用规则
func (a *App) ToggleRule(name string, enabled bool) error {
	return firewall.ToggleRule(name, enabled)
}

// BlockApp 快捷阻止程序联网
func (a *App) BlockApp(programPath string) error {
	return firewall.BlockApp(programPath)
}

// AllowApp 快捷放行程序
func (a *App) AllowApp(programPath string) error {
	return firewall.AllowApp(programPath)
}

// GetFirewallStatus 获取防火墙是否开启
func (a *App) GetFirewallStatus() bool {
	on, err := firewall.GetFirewallStatus()
	if err != nil {
		fmt.Printf("获取防火墙状态失败: %v\n", err)
		return false
	}
	return on
}

// SetFirewallEnabled 开启或关闭防火墙
func (a *App) SetFirewallEnabled(enabled bool) error {
	return firewall.SetFirewallEnabled(enabled)
}

// ResetFirewall 重置防火墙为默认规则
func (a *App) ResetFirewall() error {
	return firewall.ResetFirewall()
}

// GetServicePorts 获取常用服务端口列表
func (a *App) GetServicePorts() []ports.ServicePort {
	return ports.GetServicePorts()
}

// ChangeServicePort 修改服务端口
func (a *App) ChangeServicePort(serviceName string, newPort int) error {
	return ports.ChangeServicePort(serviceName, newPort)
}

// GetSMBOverview 获取本机 SMB 服务和协议安全状态
func (a *App) GetSMBOverview() (smb.Overview, error) {
	return smb.GetOverview(a.ctx)
}

// GetSMBShares 获取本机 SMB 共享列表
func (a *App) GetSMBShares() ([]smb.Share, error) {
	return smb.GetShares(a.ctx)
}

// SetSMBServiceRunning 启动或停止本机 SMB 服务
func (a *App) SetSMBServiceRunning(running bool) error {
	return smb.SetServiceRunning(a.ctx, running)
}

// DisableSMB1 关闭已弃用的 SMB1 服务端协议
func (a *App) DisableSMB1() error {
	return smb.DisableSMB1(a.ctx)
}

// CreateSMBShare 创建本机 SMB 共享
func (a *App) CreateSMBShare(request smb.CreateShareRequest) error {
	return smb.CreateShare(a.ctx, request)
}

// RemoveSMBShare 删除本机普通 SMB 共享
func (a *App) RemoveSMBShare(name string) error {
	return smb.RemoveShare(a.ctx, name)
}

// DiscoverSMBNetworkShares 扫描局域网中可见的磁盘共享
func (a *App) DiscoverSMBNetworkShares() ([]smb.NetworkShare, error) {
	return smb.DiscoverNetworkShares(a.ctx)
}

// OpenSMBLocation 在资源管理器中打开本机目录或网络共享
func (a *App) OpenSMBLocation(location string) error {
	return smb.OpenLocation(location)
}

// GetSMBSessions 获取连接到本机的活动 SMB 会话
func (a *App) GetSMBSessions() ([]smb.Session, error) {
	return smb.GetSessions(a.ctx)
}

// CloseSMBSession 强制断开一条活动 SMB 会话
func (a *App) CloseSMBSession(sessionID string) error {
	return smb.CloseSession(a.ctx, sessionID)
}

// GetSMBFirewallStatus 获取文件和打印机共享防火墙规则状态
func (a *App) GetSMBFirewallStatus() (smb.FirewallStatus, error) {
	return smb.GetFirewallStatus(a.ctx)
}

// SetSMBFirewallEnabled 为所有网络配置文件启用或禁用文件和打印机共享规则组
func (a *App) SetSMBFirewallEnabled(enabled bool) error {
	return smb.SetFirewallEnabled(a.ctx, enabled)
}

// OpenDirectoryDialog 打开目录选择对话框
func (a *App) OpenDirectoryDialog() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择共享目录",
	})
}

// OpenExplorer 打开资源管理器并选中指定文件
func (a *App) OpenExplorer(path string) error {
	cmd := exec.Command("explorer.exe", "/select,", path)
	return cmd.Start()
}

// OpenFileDialog 打开文件选择对话框，返回选中的文件路径
func (a *App) OpenFileDialog() (string, error) {
	// 使用 PowerShell 打开文件选择对话框
	psScript := `
	Add-Type -AssemblyName System.Windows.Forms
	$dialog = New-Object System.Windows.Forms.OpenFileDialog
	$dialog.Filter = "可执行文件 (*.exe)|*.exe|所有文件 (*.*)|*.*"
	$dialog.Title = "选择程序"
	$dialog.ShowDialog() | Out-Null
	$dialog.FileName
	`
	cmd := exec.Command("powershell", "-Command", psScript)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
