package firewall

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestRuleQueryFailurePropagates(t *testing.T) {
	failure := errors.New("mock query failure")
	rules, err := getRulesWith(Inbound, 0, func([]string) ([]byte, error) { return []byte("denied"), failure })
	if rules != nil || !errors.Is(err, failure) || !strings.Contains(err.Error(), "denied") {
		t.Fatalf("%v %v", rules, err)
	}
}

func TestSelectedRuleOperationsValidateAndIsolate(t *testing.T) {
	rule := validRule()
	rule.Name = ` 规则 '$(); [*] `
	failure := errors.New("mock operation failure")
	for _, script := range []string{deleteRuleScript, toggleRuleScript} {
		err := operateRuleWith(rule, false, "操作", script, func() bool { return true }, func(code string, env map[string]string) ([]byte, error) {
			if strings.Contains(code, rule.Name) {
				t.Fatal("input in script")
			}
			var request operationRequest
			if err := json.Unmarshal([]byte(env["FWM_RULE_UPDATE"]), &request); err != nil {
				t.Fatal(err)
			}
			if request.Original != rule || request.Enabled {
				t.Fatalf("%#v", request)
			}
			return []byte("access denied"), failure
		})
		if !errors.Is(err, failure) || !strings.Contains(err.Error(), "access denied") {
			t.Fatalf("%v", err)
		}
	}
	never := func(string, map[string]string) ([]byte, error) { t.Fatal("executed"); return nil, nil }
	if operateRuleWith(rule, false, "操作", deleteRuleScript, func() bool { return false }, never) == nil {
		t.Fatal("admin guard absent")
	}
	for _, invalid := range []FirewallRule{{Name: "", Direction: Inbound}, {Name: "rule", Direction: "wrong"}} {
		if operateRuleWith(invalid, false, "操作", deleteRuleScript, func() bool { return true }, never) == nil {
			t.Fatal("validation absent")
		}
	}
}

// 管理命令完全由 PowerShell 测试函数替换，不访问真实规则。
func TestSelectedRuleScriptUniqueTarget(t *testing.T) {
	for _, operation := range []struct{ name, script string }{{"delete", deleteRuleScript}, {"toggle", toggleRuleScript}} {
		for _, scenario := range []struct {
			count int
			local bool
		}{{0, true}, {1, true}, {2, true}, {1, false}} {
			t.Run(fmt.Sprintf("%s/%d/%v", operation.name, scenario.count, scenario.local), func(t *testing.T) {
				mock := `function Get-NetFirewallRule { param($PolicyStore,$ErrorAction,[switch]$TracePolicyStore)
  if ($PolicyStore -eq 'PersistentStore' -and $env:FWM_TEST_LOCAL -eq 'false') { return }
  for ($i=0; $i -lt [int]$env:FWM_TEST_COUNT; $i++) { [PSCustomObject]@{PolicyStoreSource='PersistentStore';PolicyStoreSourceType='Local';DisplayName='测试规则';Direction='Inbound';Name="id$i"} }
}
function Remove-NetFirewallRule { param($InputObject,$ErrorAction)
  if ($InputObject.Name -ne 'id0') { throw 'wrong target' }; $script:called=$true
}
function Set-NetFirewallRule { param($InputObject,$Enabled,$ErrorAction)
  if ($InputObject.Name -ne 'id0' -or $Enabled -ne 'False') { throw 'wrong target' }; $script:called=$true
}
`
				payload, _ := json.Marshal(operationRequest{Original: validRule(), Enabled: false})
				output, err := runUpdateScript(mock+operation.script+"\nif (-not $script:called) { throw 'not called' }", map[string]string{"FWM_RULE_UPDATE": string(payload), "FWM_TEST_COUNT": fmt.Sprint(scenario.count), "FWM_TEST_LOCAL": fmt.Sprint(scenario.local)})
				if (err == nil) != (scenario.count > 0 && scenario.local) {
					t.Fatalf("%v %s", err, output)
				}
			})
		}
	}
}

func TestSelectedRuleRejectsMissingOrAmbiguousInternalID(t *testing.T) {
	rule := validRule()
	rule.ID = ""
	never := func(string, map[string]string) ([]byte, error) { t.Fatal("executed without identity"); return nil, nil }
	if operateRuleWith(rule, false, "切换", toggleRuleScript, func() bool { return true }, never) == nil {
		t.Fatal("missing ID accepted")
	}
	if updateRuleWith(rule, rule, func() bool { return true }, never) == nil {
		t.Fatal("missing ID update accepted")
	}
	mock := `function Get-NetFirewallRule { param($PolicyStore,$ErrorAction,[switch]$TracePolicyStore)
 1..2 | ForEach-Object { [PSCustomObject]@{PolicyStoreSource='PersistentStore';PolicyStoreSourceType='Local';DisplayName='测试规则';Direction='Inbound';Name='id0'} }
}
function Remove-NetFirewallRule { throw 'unexpected removal' }
`
	payload, _ := json.Marshal(operationRequest{Original: validRule()})
	output, err := runUpdateScript(mock+deleteRuleScript, map[string]string{"FWM_RULE_UPDATE": string(payload)})
	if err == nil || !strings.Contains(string(output), "无法唯一识别") {
		t.Fatalf("ambiguous ID accepted: %s %v", output, err)
	}
}
