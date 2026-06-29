# Firewall Manager

Windows 防火墙规则管理桌面工具，基于 Wails v2（Go + Vue 3）。

## 技术栈

- **后端**: Go + netsh advfirewall 命令封装
- **前端**: Vue 3 + Vite + Naive UI（深色主题）
- **打包**: Wails v2 → 单个 exe (~12MB)

## 开发命令

```bash
wails dev          # 开发模式（热重载）
wails build        # 生产构建 → build/bin/firewall-manager.exe
wails generate module  # 重新生成 JS 绑定
```

## 项目结构

- `main.go` / `app.go` — Wails 入口与前端绑定方法
- `internal/firewall/` — 防火墙操作封装（netsh 命令）
- `internal/admin/` — Windows 管理员权限检测
- `frontend/src/components/` — Vue 组件（AppLayout、RuleList、RuleEditor）
- `frontend/src/stores/rules.js` — 响应式状态管理
- `frontend/wailsjs/` — 自动生成的 Go↔JS 绑定（勿手动编辑）

## 注意事项

- 防火墙操作需要管理员权限，UAC manifest 已配置 `requireAdministrator`
- netsh 命令需隐藏控制台窗口（`SysProcAttr.HideWindow = true`）
- Naive UI 通过 `app.use(naive)` 全局注册
