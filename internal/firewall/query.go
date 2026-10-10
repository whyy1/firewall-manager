package firewall

import (
	"encoding/json"
	"fmt"
	"strings"
)

// 过滤器批量缓存仅以实例标识关联；缺失或不唯一时使用系统关联查询。
const queryRulesScript = `$direction = if ($env:FWM_RULE_DIRECTION -eq 'in') { 'Inbound' } else { 'Outbound' }
function Index-Filters($filters) {
  $index = @{}
  foreach ($filter in $filters) {
    $key = [string]$filter.InstanceID
    if ($key) {
      if ($index.ContainsKey($key)) { $index[$key] = @($index[$key]) + @($filter) }
      else { $index[$key] = @($filter) }
    }
  }
  return $index
}
function Join-Values($values, $optional = $false) {
  $items = @($values | ForEach-Object { [string]$_ })
  if ($items.Count -eq 0 -or @($items | Where-Object { [string]::IsNullOrWhiteSpace($_) }).Count -gt 0) {
    if ($optional) { return '' }
    throw '规则关键过滤条件缺失，已拒绝生成不完整列表'
  }
  return ($items -join ',')
}
$activeRules = @(Get-NetFirewallRule -PolicyStore ActiveStore -TracePolicyStore -ErrorAction Stop)
$rules = @($activeRules | Where-Object { [string]$_.Direction -eq $direction })
if ($rules.Count -eq 0) { Write-Output '[]'; return }
$activeIDs = @{}
foreach ($rule in $activeRules) {
  $id = [string]$rule.Name
  if ($activeIDs.ContainsKey($id)) { $activeIDs[$id]++ } else { $activeIDs[$id] = 1 }
}
$localRules = @{}
foreach ($local in @(Get-NetFirewallRule -PolicyStore PersistentStore -ErrorAction Stop)) {
  $id = [string]$local.Name
  if ($localRules.ContainsKey($id)) { $localRules[$id] = @($localRules[$id]) + @($local) }
  else { $localRules[$id] = @($local) }
}
$instances = Index-Filters $activeRules
$ports = Index-Filters @(Get-NetFirewallPortFilter -PolicyStore ActiveStore -ErrorAction Stop)
$addresses = Index-Filters @(Get-NetFirewallAddressFilter -PolicyStore ActiveStore -ErrorAction Stop)
$applications = Index-Filters @(Get-NetFirewallApplicationFilter -PolicyStore ActiveStore -ErrorAction Stop)
$result = @(foreach ($rule in $rules) {
  $key = [string]$rule.InstanceID
  $port = @()
  $address = @()
  $application = @()
  if ($ports.ContainsKey($key)) { $port = @($ports[$key]) }
  if ($addresses.ContainsKey($key)) { $address = @($addresses[$key]) }
  if ($applications.ContainsKey($key)) { $application = @($applications[$key]) }
  $sharedInstance = -not $key -or @($instances[$key]).Count -ne 1
  if ($port.Count -ne 1 -or $sharedInstance) { $port = @(Get-NetFirewallPortFilter -AssociatedNetFirewallRule $rule -ErrorAction Stop) }
  if ($address.Count -ne 1 -or $sharedInstance) { $address = @(Get-NetFirewallAddressFilter -AssociatedNetFirewallRule $rule -ErrorAction Stop) }
  if ($application.Count -ne 1 -or $sharedInstance) { $application = @(Get-NetFirewallApplicationFilter -AssociatedNetFirewallRule $rule -ErrorAction Stop) }
  if ($port.Count -ne 1 -or $address.Count -ne 1 -or $application.Count -ne 1) { throw '规则过滤器关联不唯一，请刷新后重试' }
  $protocol = (Join-Values $port[0].Protocol).ToLowerInvariant()
  switch ($protocol) {
    '6' { $protocol = 'tcp' }
    '17' { $protocol = 'udp' }
    '1' { $protocol = 'icmpv4' }
    '58' { $protocol = 'icmpv6' }
    '256' { $protocol = 'any' }
  }
  $canEdit = $false
  $reason = '规则来自策略、动态或未知来源，仅供查看'
  $id = [string]$rule.Name
  if ([string]$rule.PolicyStoreSourceType -eq 'Local' -and [string]$rule.PolicyStoreSource -eq 'PersistentStore') {
    $reason = '没有唯一匹配的本地持久规则，仅供查看'
    if ($activeIDs[$id] -eq 1 -and $localRules.ContainsKey($id) -and @($localRules[$id]).Count -eq 1) {
      $local = @($localRules[$id])[0]
      if ($local.Name -ceq $rule.Name -and $local.DisplayName -ceq $rule.DisplayName -and [string]$local.Direction -eq [string]$rule.Direction) {
        $canEdit = $true
        $reason = ''
      }
    }
  }
  $enabled = [string]$rule.Enabled
  if ($enabled -ne 'True' -and $enabled -ne 'False') { throw '规则启用状态无效，已拒绝生成不完整列表' }
  [pscustomobject]@{
    id = [string]$rule.Name
    policyStoreSource = [string]$rule.PolicyStoreSource
    policyStoreSourceType = [string]$rule.PolicyStoreSourceType
    canEdit = $canEdit
    readOnlyReason = $reason
    name = [string]$rule.DisplayName
    direction = $env:FWM_RULE_DIRECTION
    action = (Join-Values $rule.Action).ToLowerInvariant()
    enabled = ([string]$rule.Enabled -eq 'True')
    profile = ((Join-Values $rule.Profile).Replace(' ', '')).ToLowerInvariant()
    program = Join-Values $application[0].Program
    localAddr = Join-Values $address[0].LocalAddress
    remoteAddr = Join-Values $address[0].RemoteAddress
    localPort = Join-Values $port[0].LocalPort
    remotePort = Join-Values $port[0].RemotePort
    protocol = $protocol
    icmpType = Join-Values $port[0].IcmpType ($protocol -ne 'icmpv4' -and $protocol -ne 'icmpv6')
  }
})
ConvertTo-Json -InputObject $result -Depth 5 -Compress`

func queryRulesWith(direction RuleDirection, limit int, run func(string, map[string]string) ([]byte, error)) ([]FirewallRule, error) {
	if direction != Inbound && direction != Outbound {
		return nil, fmt.Errorf("查询规则失败: 规则方向无效")
	}
	output, err := run(queryRulesScript, map[string]string{"FWM_RULE_DIRECTION": string(direction)})
	if err != nil {
		return nil, fmt.Errorf("查询规则失败: %w\n%s", err, strings.TrimSpace(string(output)))
	}
	var rules []FirewallRule
	if err := json.Unmarshal(output, &rules); err != nil {
		return nil, fmt.Errorf("解析规则列表失败: %w", err)
	}
	if rules == nil {
		rules = []FirewallRule{}
	}
	type identity struct{ sourceType, source, id string }
	seen := make(map[identity]bool, len(rules))
	idCounts := make(map[string]int, len(rules))
	for _, rule := range rules {
		idCounts[strings.ToLower(rule.ID)]++
	}
	for index, rule := range rules {
		if strings.TrimSpace(rule.ID) == "" || rule.Direction != direction {
			return nil, fmt.Errorf("查询规则失败: 返回的规则标识或方向无效")
		}
		key := identity{rule.PolicyStoreSourceType, rule.PolicyStoreSource, rule.ID}
		if seen[key] {
			return nil, fmt.Errorf("查询规则失败: 来源和内部标识重复，无法唯一识别规则")
		}
		seen[key] = true
		if rule.CanEdit {
			if idCounts[strings.ToLower(rule.ID)] != 1 {
				rules[index].CanEdit = false
				rules[index].ReadOnlyReason = "有效规则内部标识不唯一，仅供查看"
			} else if err := validateRuleIdentity(rule); err != nil {
				rules[index].CanEdit = false
				rules[index].ReadOnlyReason = err.Error()
			} else if _, err := normalizeRuleWith(rule, true); err != nil {
				rules[index].CanEdit = false
				rules[index].ReadOnlyReason = "规则参数尚不支持编辑: " + err.Error()
			}
		}
	}
	if limit > 0 && len(rules) > limit {
		rules = rules[:limit]
	}
	return rules, nil
}
