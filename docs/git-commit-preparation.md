# Git 提交审核准备

日期：2026-10-10。当前在 main，修改基于 4c53b944d89d515afc4816d7638603605cd199e0；尚未暂存或提交。

## 建议提交说明

fix: harden firewall and SMB management operations

- Identify firewall rules by internal ID and policy provenance; preserve duplicate display names and reject ambiguous or read-only targets.
- Validate rule parameters and preserve untouched security properties; propagate read failures and prevent repeated writes.
- Correct RDP/WinRM error handling and make SMB TCP 445 read-only.
- Recover SMB service controls when protocol queries fail; validate share paths and permissions, protect special shares, and request visible Explorer windows.
- Rework management views and add isolated Go, state, and mocked Playwright coverage.
- Update bilingual README, acceptance material, and selected documentation screenshots.

## 验证证据

最新功能改动已运行 gofmt、go test ./... -count=1、go vet ./...、30 项前端单元测试、28 项模拟界面测试、生产构建与 Wails 构建，均通过。两个 SMB 系统集成开关关闭。前端保留包体积警告。

README/文档更新不改运行代码，检查 Markdown 图片路径、Git 空白差异及构建。用户确认 SMB 打开目录修复复测通过；其他管理员项目按现有验收记录审核，本文不宣称所有实机项目均完成。

## 排除与保留

- 不提交 build/bin、frontend/dist、node_modules、frontend/test-results、package-lock 或临时文件；六张选定 PNG 是文档资源，允许提交。
- .gitignore 中用户原有的 .*、*.test、AGENTS.md 等更改不随本次提交。项目新增的 frontend/test-results/ 忽略规则可仅暂存该末尾 hunk，保留用户其他行未暂存；审核前不自动操作索引。
- frontend/wailsjs 为 Wails 生成绑定；frontend/package.json.md5 已跟踪，是 Wails 生成的依赖校验记录。检查后随对应源码更新。
- 不修改版本配置，不创建提交、推送、标签或 Release。上述动作等待用户明确授权范围。

## 建议提交文件清单
- app.go
- docs/git-commit-preparation.md
- docs/rule-identity-plan.md
- docs/screenshots/ports-dark.png
- docs/screenshots/ports-light.png
- docs/screenshots/rules-dark.png
- docs/screenshots/rules-light.png
- docs/screenshots/smb-dark.png
- docs/screenshots/smb-light.png
- docs/smb-management-acceptance.md
- docs/smb-management-plan.md
- docs/v0.4.0-manual-acceptance.md
- frontend/package.json
- frontend/package.json.md5
- frontend/playwright.config.js
- frontend/src/App.vue
- frontend/src/components/AppLayout.vue
- frontend/src/components/PortManager.vue
- frontend/src/components/RuleEditor.vue
- frontend/src/components/RuleList.vue
- frontend/src/components/smb/SMBFirewallProfiles.vue
- frontend/src/components/smb/SMBLocalShares.vue
- frontend/src/components/smb/SMBNetworkShares.vue
- frontend/src/components/smb/SMBSessions.vue
- frontend/src/components/SMBManager.vue
- frontend/src/stores/operations.js
- frontend/src/stores/operations.test.js
- frontend/src/stores/port_state.js
- frontend/src/stores/port_state.test.js
- frontend/src/stores/ports.js
- frontend/src/stores/rule_form.js
- frontend/src/stores/rule_identity.js
- frontend/src/stores/rules_state.js
- frontend/src/stores/rules_state.test.js
- frontend/src/stores/rules.js
- frontend/src/stores/smb.js
- frontend/src/stores/smb.test.js
- frontend/src/styles/global.css
- frontend/tests/ui/actions.spec.js
- frontend/tests/ui/console.spec.js
- frontend/tests/ui/mock-app.js
- frontend/tests/ui/states.spec.js
- frontend/wailsjs/go/main/App.d.ts
- frontend/wailsjs/go/main/App.js
- frontend/wailsjs/go/models.ts
- internal/diagnostics/powershell.go
- internal/firewall/firewall_test.go
- internal/firewall/firewall.go
- internal/firewall/identity_test.go
- internal/firewall/operations_test.go
- internal/firewall/operations.go
- internal/firewall/query_test.go
- internal/firewall/query.go
- internal/firewall/types.go
- internal/firewall/update.go
- internal/firewall/validation.go
- internal/ports/operations_test.go
- internal/ports/operations.go
- internal/ports/ports_test.go
- internal/ports/ports.go
- internal/smb/backend.go
- internal/smb/diagnostics.go
- internal/smb/network_test.go
- internal/smb/network.go
- internal/smb/open_test.go
- internal/smb/open.go
- internal/smb/overview_test.go
- internal/smb/p1.go
- internal/smb/paths.go
- internal/smb/security_test.go
- internal/smb/smb_test.go
- internal/smb/smb.go
- README.md
