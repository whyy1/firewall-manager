# 同名防火墙规则身份迁移与验收

日期：2026-10-10。范围为现有规则管理，允许 Windows DisplayName 重复，不新增规则重命名功能。

## 实施顺序

1. 查询 ActiveStore 并使用 TracePolicyStore 读取来源；内部 Name、来源类型及来源组成身份。批量读取过滤器与 PersistentStore，歧义使用系统对象关联，关键字段缺失报错。
2. 仅来源明确的本地规则、唯一 PersistentStore 对象及界面支持的参数可修改。编辑、启停、删除复读验证身份、来源、原名称及方向，拒绝来源变化和目标歧义。
3. 界面展示可复制内部 ID 与完整来源、只读原因；行键使用复合身份，删除确认展示 ID。新增允许同名，编辑名称保持只读。
4. 补齐模拟后端、状态和浏览器测试；重新执行格式化、全部 Go 检查、前端测试、界面测试与生产构建，生成最新 Windows 构建及校验文件。
5. 用户执行管理员实机验收，包括同名各自操作、策略覆盖保护及完整查询耗时。通过后才准备 README、正式版本及发版材料；提交、推送、标签与发布等待明确授权。

## 安全边界

自动化测试仅使用模拟执行器、覆盖的 PowerShell cmdlet 或模拟 Wails API，不操作真实规则、服务、共享或注册表。保留其他未提交改动。

## 验收证据

模拟验证不能证明管理员实机行为。完整查询沿用 30 秒超时，需要在管理员环境记录规模与耗时；超时或关键项目未验证时不正式发布。具体结果记录于 smb-management-acceptance.md，实机步骤见 v0.4.0-manual-acceptance.md。

## 参考

Microsoft Get-NetFirewallRule 文档：https://learn.microsoft.com/en-us/powershell/module/netsecurity/get-netfirewallrule
内部 Name 在策略存储内唯一；TracePolicyStore 提供有效规则的来源信息。不同来源不能仅用显示名或内部 Name 判定为同一对象。