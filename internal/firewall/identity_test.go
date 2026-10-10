package firewall

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestQueryCompositeIdentityAndReadOnly(t *testing.T) {
	for _, scenario := range []struct {
		name, source, kind, action string
		editable                   bool
	}{
		{"local", "PersistentStore", "Local", "allow", true},
		{"policy", "domain\\policy", "GroupPolicy", "allow", false},
		{"unknown", "", "", "allow", false},
		{"configurable", "ConfigurableServiceStore", "Local", "allow", false},
		{"host", "computer-name", "Local", "allow", false},
		{"unsupported", "PersistentStore", "Local", "bypass", false},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			rule := validRule()
			rule.PolicyStoreSource, rule.PolicyStoreSourceType, rule.Action = scenario.source, scenario.kind, RuleAction(scenario.action)
			payload, _ := json.Marshal([]FirewallRule{rule})
			rules, err := queryRulesWith(Inbound, 0, func(string, map[string]string) ([]byte, error) { return payload, nil })
			if err != nil || len(rules) != 1 || rules[0].CanEdit != scenario.editable {
				t.Fatalf("%#v %v", rules, err)
			}
			if !scenario.editable && rules[0].ReadOnlyReason == "" {
				t.Fatal("missing read-only explanation")
			}
		})
	}
	first, second := validRule(), validRule()
	second.PolicyStoreSource, second.PolicyStoreSourceType = "domain\\policy", "GroupPolicy"
	payload, _ := json.Marshal([]FirewallRule{first, second})
	rules, err := queryRulesWith(Inbound, 0, func(string, map[string]string) ([]byte, error) { return payload, nil })
	if err != nil || len(rules) != 2 || rules[0].CanEdit || rules[1].CanEdit {
		t.Fatalf("cross-source ID %#v %v", rules, err)
	}
	payload, _ = json.Marshal([]FirewallRule{first, first})
	if _, err := queryRulesWith(Inbound, 0, func(string, map[string]string) ([]byte, error) { return payload, nil }); err == nil {
		t.Fatal("duplicate composite identity accepted")
	}
}

func TestWritesRejectUnknownOrChangedSource(t *testing.T) {
	never := func(string, map[string]string) ([]byte, error) { t.Fatal("unexpected command"); return nil, nil }
	for _, source := range []string{"", "ConfigurableServiceStore", "computer-name"} {
		rule := validRule()
		rule.PolicyStoreSource = source
		if operateRuleWith(rule, false, "删除", deleteRuleScript, func() bool { return true }, never) == nil {
			t.Fatal("unknown source accepted")
		}
	}
	original, updated := validRule(), validRule()
	updated.PolicyStoreSource = "other"
	if _, err := buildUpdateRequest(original, updated); err == nil {
		t.Fatal("source mutation accepted")
	}
	for _, sourceType := range []string{"GroupPolicy", "Dynamic", "Local"} {
		mock := `function Get-NetFirewallRule { param($PolicyStore,$ErrorAction,[switch]$TracePolicyStore)
  if ($PolicyStore -eq 'PersistentStore') { [pscustomobject]@{Name='id0';DisplayName='测试规则';Direction='Inbound';PolicyStoreSource='PersistentStore';PolicyStoreSourceType='Local'};return }
  [pscustomobject]@{ Name='id0';DisplayName='测试规则';Direction='Inbound';PolicyStoreSource=$env:FWM_TEST_SOURCE;PolicyStoreSourceType=$env:FWM_TEST_KIND }
}
function Remove-NetFirewallRule { throw 'unexpected removal' }
`
		request, _ := json.Marshal(operationRequest{Original: original})
		output, err := runUpdateScript(mock+deleteRuleScript, map[string]string{"FWM_RULE_UPDATE": string(request), "FWM_TEST_SOURCE": "changed-source", "FWM_TEST_KIND": sourceType})
		if err == nil || !strings.Contains(string(output), "来源或身份已改变") {
			t.Fatalf("source change accepted: %s %v", output, err)
		}
	}
}

func TestQueryMissingCriticalFieldsRejectsInsteadOfAny(t *testing.T) {
	for _, missing := range []string{"protocol", "localAddress", "remoteAddress", "localPort", "remotePort", "program", "profile", "icmp"} {
		mock := mockQueryCmdlets + `
function Get-NetFirewallPortFilter {
 param($PolicyStore,$AssociatedNetFirewallRule,$ErrorAction)
 foreach ($id in @('id-a','id-b')) {
  $protocol=58;$remotePort='Any';$localPort='Any';$icmp='128:Any'
  if ($env:FWM_TEST_MISSING -eq 'protocol') {$protocol=$null}
  if ($env:FWM_TEST_MISSING -eq 'localPort') {$localPort=$null}
  if ($env:FWM_TEST_MISSING -eq 'remotePort') {$remotePort=$null}
  if ($env:FWM_TEST_MISSING -eq 'icmp') {$icmp=$null}
  [pscustomobject]@{InstanceID=$id;Protocol=$protocol;LocalPort=$localPort;RemotePort=$remotePort;IcmpType=$icmp}
 }
}

function Get-NetFirewallAddressFilter {
 param($PolicyStore,$AssociatedNetFirewallRule,$ErrorAction)
 foreach ($id in @('id-a','id-b')) {
  $local='Any';$remote='Any'
  if ($env:FWM_TEST_MISSING -eq 'localAddress') {$local=$null}
  if ($env:FWM_TEST_MISSING -eq 'remoteAddress') {$remote=$null}
  [pscustomobject]@{InstanceID=$id;LocalAddress=$local;RemoteAddress=$remote}
 }
}
function Get-NetFirewallApplicationFilter {
 param($PolicyStore,$AssociatedNetFirewallRule,$ErrorAction)
 foreach ($id in @('id-a','id-b')) {
  $program='Any';if($env:FWM_TEST_MISSING -eq 'program'){$program=$null}
  [pscustomobject]@{InstanceID=$id;Program=$program}
 }
}
`
		if missing == "profile" {
			mock = strings.ReplaceAll(mock, "Profile='Domain, Private'", "Profile=$null")
		}
		_, err := queryRulesWith(Inbound, 0, func(script string, env map[string]string) ([]byte, error) {
			env["FWM_TEST_MISSING"] = missing
			return runUpdateScript(mock+script, env)
		})
		if err == nil || !strings.Contains(err.Error(), "关键过滤条件缺失") {
			t.Fatalf("%s accepted: %v", missing, err)
		}
	}
}

func TestQuerySharedInstanceUsesObjectAssociations(t *testing.T) {
	mock := `
function Get-NetFirewallRule {
 param($PolicyStore,$Direction,$ErrorAction,[switch]$TracePolicyStore)
 if ($PolicyStore -eq 'PersistentStore') {
  [pscustomobject]@{Name='shared-id';DisplayName='same';Direction='Inbound'};return
 }
 foreach($source in @('PersistentStore','domain-policy')) {
  $kind=if($source -eq 'PersistentStore'){'Local'}else{'GroupPolicy'}
  [pscustomobject]@{Name='shared-id';InstanceID='shared-id';DisplayName='same';Direction='Inbound';Action='Allow';Enabled='True';Profile='Any';PolicyStoreSource=$source;PolicyStoreSourceType=$kind}
 }
}
function Get-NetFirewallPortFilter {
 param($PolicyStore,$AssociatedNetFirewallRule,$ErrorAction)
 $port='9999'
 if($AssociatedNetFirewallRule){$port=if($AssociatedNetFirewallRule.PolicyStoreSource -eq 'PersistentStore'){'445'}else{'22'}}
 [pscustomobject]@{InstanceID='shared-id';Protocol=6;LocalPort=$port;RemotePort='Any'}
}
function Get-NetFirewallAddressFilter {
 param($PolicyStore,$AssociatedNetFirewallRule,$ErrorAction)
 if(-not $AssociatedNetFirewallRule){[pscustomobject]@{InstanceID='shared-id';LocalAddress='Any';RemoteAddress='Any'};return}
 [pscustomobject]@{LocalAddress='Any';RemoteAddress='Any'}
}
function Get-NetFirewallApplicationFilter {
 param($PolicyStore,$AssociatedNetFirewallRule,$ErrorAction)
 if(-not $AssociatedNetFirewallRule){[pscustomobject]@{InstanceID='shared-id';Program='wrong.exe'};return}
 [pscustomobject]@{Program='Any'}
}
`
	rules, err := queryRulesWith(Inbound, 0, func(script string, env map[string]string) ([]byte, error) { return runUpdateScript(mock+script, env) })
	if err != nil || len(rules) != 2 || rules[0].LocalPort != "445" || rules[1].LocalPort != "22" || rules[0].Program != "Any" || rules[0].CanEdit || rules[1].CanEdit {
		t.Fatalf("unsafe association: %#v %v", rules, err)
	}
}

func TestWriteRejectsDuplicatePersistentObject(t *testing.T) {
	mock := `function Get-NetFirewallRule {
 param($PolicyStore,$ErrorAction,[switch]$TracePolicyStore)
 $count=if($PolicyStore -eq 'ActiveStore'){1}else{2}
 1..$count | ForEach-Object {[pscustomobject]@{Name='id0';DisplayName='测试规则';Direction='Inbound';PolicyStoreSource='PersistentStore';PolicyStoreSourceType='Local'}}
}
function Set-NetFirewallRule { throw 'unexpected write' }
`
	request, _ := json.Marshal(operationRequest{Original: validRule()})
	output, err := runUpdateScript(mock+toggleRuleScript, map[string]string{"FWM_RULE_UPDATE": string(request)})
	if err == nil || !strings.Contains(string(output), "不是唯一可编辑的本地规则") {
		t.Fatalf("ambiguous persistent object accepted: %s %v", output, err)
	}
}

func TestQueryRejectsMissingOrAmbiguousAssociatedFilters(t *testing.T) {
	for _, count := range []string{"0", "2"} {
		mock := mockQueryCmdlets + `
function Get-NetFirewallPortFilter {
 param($PolicyStore,$AssociatedNetFirewallRule,$ErrorAction)
 if (-not $AssociatedNetFirewallRule) { return }
 for ($i=0;$i -lt [int]$env:FWM_TEST_COUNT;$i++) {
  [pscustomobject]@{InstanceID='id-a';Protocol=6;LocalPort='445';RemotePort='Any'}
 }
}

`
		_, err := queryRulesWith(Inbound, 0, func(script string, env map[string]string) ([]byte, error) {
			env["FWM_TEST_COUNT"] = count
			return runUpdateScript(mock+script, env)
		})
		if err == nil || !strings.Contains(err.Error(), "过滤器关联不唯一") {
			t.Fatalf("count %s accepted: %v", count, err)
		}
	}
}

func TestQueryCrossDirectionIdentityIsReadOnly(t *testing.T) {
	mock := mockQueryCmdlets + `
function Get-NetFirewallRule {
 param($PolicyStore,$Direction,$ErrorAction,[switch]$TracePolicyStore)
 [pscustomobject]@{Name='id-a';InstanceID='id-a';DisplayName='same';Direction='Inbound';Action='Allow';Enabled='True';Profile='Any';PolicyStoreSource='PersistentStore';PolicyStoreSourceType='Local'}
 if($PolicyStore -eq 'ActiveStore') {
  [pscustomobject]@{Name='id-a';InstanceID='outbound-instance';DisplayName='same';Direction='Outbound';Action='Allow';Enabled='True';Profile='Any';PolicyStoreSource='PersistentStore';PolicyStoreSourceType='Local'}
 }
}
`
	rules, err := queryRulesWith(Inbound, 0, func(script string, env map[string]string) ([]byte, error) { return runUpdateScript(mock+script, env) })
	if err != nil || len(rules) != 1 || rules[0].CanEdit || rules[0].Direction != Inbound {
		t.Fatalf("cross-direction ambiguity accepted: %#v %v", rules, err)
	}
}
