package firewall

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestQueryRuleIdentityAndLimit(t *testing.T) {
	run := func(script string, env map[string]string) ([]byte, error) {
		if script != queryRulesScript || env["FWM_RULE_DIRECTION"] != "in" {
			t.Fatal("unexpected query parameters")
		}
		return []byte(`[{"id":"first-id","name":"重名","direction":"in"},{"id":"second-id","name":"重名","direction":"in"}]`), nil
	}
	rules, err := queryRulesWith(Inbound, 0, run)
	if err != nil || len(rules) != 2 || rules[0].ID == rules[1].ID || rules[0].Name != rules[1].Name {
		t.Fatalf("identity not preserved: %#v, %v", rules, err)
	}
	rules, err = queryRulesWith(Inbound, 1, run)
	if err != nil || len(rules) != 1 || rules[0].ID != "first-id" {
		t.Fatalf("limit failed: %#v, %v", rules, err)
	}
}

func TestQueryRuleFailures(t *testing.T) {
	called := false
	_, err := queryRulesWith("invalid", 0, func(string, map[string]string) ([]byte, error) {
		called = true
		return nil, nil
	})
	if err == nil || called {
		t.Fatal("invalid direction executed query")
	}
	for _, output := range []string{`not json`, `[{"id":"","direction":"in"}]`, `[{"id":"id","direction":"out"}]`} {
		_, err := queryRulesWith(Inbound, 0, func(string, map[string]string) ([]byte, error) { return []byte(output), nil })
		if err == nil {
			t.Fatalf("accepted %q", output)
		}
	}
	_, err = queryRulesWith(Inbound, 0, func(string, map[string]string) ([]byte, error) {
		return []byte("mock access denied"), errors.New("failed")
	})
	if err == nil || !strings.Contains(err.Error(), "mock access denied") {
		t.Fatalf("lost diagnostic: %v", err)
	}
	for _, output := range []string{`[]`, `null`} {
		rules, err := queryRulesWith(Inbound, 0, func(string, map[string]string) ([]byte, error) { return []byte(output), nil })
		if err != nil || rules == nil || len(rules) != 0 {
			t.Fatalf("empty result failed: %#v, %v", rules, err)
		}
	}
}

// 所有 NetSecurity 查询由本地函数替换，测试不读取或写入真实防火墙。
const mockQueryCmdlets = `
function Get-NetFirewallRule {
  param($PolicyStore, $Direction, $ErrorAction, [switch]$TracePolicyStore)
  if ($env:FWM_TEST_EMPTY -eq 'yes') { return }
  if ($PolicyStore -eq 'ActiveStore' -and ($Direction -or -not $TracePolicyStore)) { throw 'unexpected rule query' }
  foreach ($id in @('id-a', 'id-b')) {
    [pscustomobject]@{ Name=$id; InstanceID=$id; DisplayName='same name'; Direction='Inbound'; Action='Allow'; Enabled='True'; Profile='Domain, Private'; PolicyStoreSource='PersistentStore'; PolicyStoreSourceType='Local' }
  }
}
function Get-NetFirewallPortFilter {
  param($PolicyStore, $AssociatedNetFirewallRule, $ErrorAction)
  if ($AssociatedNetFirewallRule) {
    if ($env:FWM_TEST_FALLBACK -eq 'no') { throw 'unexpected association query' }
    [pscustomobject]@{ InstanceID='different'; Protocol=58; LocalPort='Any'; RemotePort='Any'; IcmpType=@('128:Any', '129:0') }
    return
  }
  if ($env:FWM_TEST_FALLBACK -eq 'yes') { return }
  [pscustomobject]@{ InstanceID='id-b'; Protocol=6; LocalPort=@('445','139'); RemotePort='Any'; IcmpType='Any' }
  [pscustomobject]@{ InstanceID='id-a'; Protocol=58; LocalPort='Any'; RemotePort='Any'; IcmpType=@('128:Any','129:0') }
  if ($env:FWM_TEST_FALLBACK -eq 'duplicate') {
    [pscustomobject]@{ InstanceID='id-a'; Protocol=6; LocalPort='22'; RemotePort='Any'; IcmpType='Any' }
  }
}
function Get-NetFirewallAddressFilter {
  param($PolicyStore, $AssociatedNetFirewallRule, $ErrorAction)
  if ($AssociatedNetFirewallRule) { throw 'unexpected address association' }
  foreach ($id in @('id-b','id-a')) { [pscustomobject]@{ InstanceID=$id; LocalAddress='Any'; RemoteAddress=@('10.0.0.0/24','2001:db8::/64') } }
}
function Get-NetFirewallApplicationFilter {
  param($PolicyStore, $AssociatedNetFirewallRule, $ErrorAction)
  if ($AssociatedNetFirewallRule) { throw 'unexpected application association' }
  foreach ($id in @('id-a','id-b')) { [pscustomobject]@{ InstanceID=$id; Program='C:\Program Files\app.exe' } }
}
`

func TestQueryScriptMockedFilters(t *testing.T) {
	for _, fallback := range []string{"no", "yes", "duplicate"} {
		t.Run(fallback, func(t *testing.T) {
			rules, err := queryRulesWith(Inbound, 0, func(script string, env map[string]string) ([]byte, error) {
				env["FWM_TEST_FALLBACK"] = fallback
				return runUpdateScript(mockQueryCmdlets+script, env)
			})
			if err != nil {
				t.Fatal(err)
			}
			if len(rules) != 2 || rules[0].ID != "id-a" || rules[1].ID != "id-b" {
				t.Fatalf("incorrect identity: %#v", rules)
			}
			first := rules[0]
			if first.Protocol != "icmpv6" || first.IcmpType != "128:Any,129:0" || first.Profile != "domain,private" || first.RemoteAddr != "10.0.0.0/24,2001:db8::/64" || !first.Enabled {
				t.Fatalf("lost filter semantics: %#v", first)
			}
			if fallback == "no" && (rules[1].Protocol != "tcp" || rules[1].LocalPort != "445,139") {
				t.Fatalf("mapped filters to wrong rule: %#v", rules[1])
			}
		})
	}
}

func TestQueryScriptMockedEmpty(t *testing.T) {
	output, err := runUpdateScript(mockQueryCmdlets+queryRulesScript, map[string]string{"FWM_RULE_DIRECTION": "in", "FWM_TEST_EMPTY": "yes"})
	var rules []FirewallRule
	if err != nil || json.Unmarshal(output, &rules) != nil || len(rules) != 0 {
		t.Fatalf("empty array: %s, %v", output, err)
	}
}
