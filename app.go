package main

import (
	"context"
	"fmt"
	"os/exec"

	"firewall-manager/internal/admin"
	"firewall-manager/internal/firewall"
	"firewall-manager/internal/ports"
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

// OpenExplorer 打开资源管理器并选中指定文件
func (a *App) OpenExplorer(path string) error {
	cmd := exec.Command("explorer.exe", "/select,", path)
	return cmd.Start()
}
