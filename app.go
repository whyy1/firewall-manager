package main

import (
	"context"
	"fmt"

	"firewall-manager/internal/admin"
	"firewall-manager/internal/firewall"
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
