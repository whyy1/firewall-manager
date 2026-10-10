package firewall

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func validRule() FirewallRule {
	return FirewallRule{ID: "id0", PolicyStoreSource: "PersistentStore", PolicyStoreSourceType: "Local", CanEdit: true, Name: "测试规则", Direction: Inbound, Action: Allow, Protocol: "tcp", Enabled: true}
}

func TestLocalizedRuleRoundTrip(t *testing.T) {
	rules, err := parseRules("规则名称: 测试规则\n已启用: 是\n操作: 允许\n程序: 任何\n本地 IP: 任何\n远程 IP: 本地子网\n协议: TCP\n本地端口: 445\n远程端口: 任何\n配置文件: 域,专用,公用\n", Inbound)
	if err != nil || len(rules) != 1 {
		t.Fatalf("parseRules: %v, %#v", err, rules)
	}
	normalized, err := normalizeRule(rules[0])
	if err != nil {
		t.Fatal(err)
	}
	args := strings.Join(buildAddArgs(normalized), " ")
	for _, value := range []string{"localip=any", "remoteip=localsubnet", "profile=domain,private,public", "localport=445", "remoteport=any"} {
		if !strings.Contains(args, value) {
			t.Fatalf("missing %q: %s", value, args)
		}
	}
	if normalized.Program != "" {
		t.Fatalf("Program = %q", normalized.Program)
	}
}

func TestAddressValidation(t *testing.T) {
	for _, value := range []string{"Any", "任何", "192.168.1.2", "2001:db8::1", "10.0.0.0/24", "2001:db8::/64", "10.0.0.1-10.0.0.9", "10.0.0.1, 2001:db8::1", "LocalSubnet", "dns,dhcp,wins,defaultgateway", "10.0.0.0/255.255.255.0"} {
		t.Run(value, func(t *testing.T) {
			if _, err := normalizeAddresses(value, true); err != nil {
				t.Fatal(err)
			}
		})
	}
	for _, value := range []string{"unknown", "999.1.1.1", "10.0.0.1/33", "10.0.0.9-10.0.0.1", "10.0.0.1-::1", "any,10.0.0.1", "10.0.0.1,", "fe80::1%eth0", "10.0.0.0/255.0.255.0"} {
		t.Run(value, func(t *testing.T) {
			if _, err := normalizeAddresses(value, true); err == nil {
				t.Fatal("accepted invalid address")
			}
		})
	}
	if _, err := normalizeAddresses("localsubnet", false); err == nil {
		t.Fatal("local keyword accepted")
	}
}

func TestProtocolAndPorts(t *testing.T) {
	for _, test := range []struct {
		protocol, port string
		valid          bool
	}{
		{"TCP", "80,443,1000-2000", true}, {"6", "445", true}, {"UDP", "53", true},
		{"tcp", "RPC-EPMap", true}, {"udp", "RPC", false}, {"any", "445", false},
		{"icmpv4:8,any", "any", true}, {"256", "any", false}, {"tcp", "65536", false},
		{"tcp", "99-1", false}, {"tcp", "80,", false},
	} {
		rule := validRule()
		rule.Protocol = test.protocol
		rule.LocalPort = test.port
		_, err := normalizeRule(rule)
		if (err == nil) != test.valid {
			t.Errorf("%s/%s: %v", test.protocol, test.port, err)
		}
	}
}

func TestAddValidationPreventsExecution(t *testing.T) {
	rule := validRule()
	rule.RemoteAddr = "不是地址"
	err := addRuleWith(rule, func() bool { return true }, func([]string) ([]byte, error) { t.Fatal("command executed"); return nil, nil })
	if err == nil || !strings.Contains(err.Error(), "远程地址") {
		t.Fatalf("error = %v", err)
	}
}

func TestAddErrorDecodesGBK(t *testing.T) {
	runErr := errors.New("exit status 1")
	err := addRuleWith(validRule(), func() bool { return true }, func([]string) ([]byte, error) { return []byte{0xb4, 0xed, 0xce, 0xf3}, runErr })
	if !errors.Is(err, runErr) || !strings.Contains(err.Error(), "错误") {
		t.Fatalf("error = %v", err)
	}
}

func TestUpdateOnlyChangedFields(t *testing.T) {
	before := validRule()
	before.LocalAddr = "任何"
	before.Profile = "专用"
	before.LocalPort = "445"
	after := before
	after.LocalAddr = "any"
	after.Profile = "private"
	after.RemoteAddr = "192.168.1.0/24"
	request, err := buildUpdateRequest(before, after)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(request.Changes, map[string]interface{}{"RemoteAddress": "192.168.1.0/24"}) {
		t.Fatalf("changes = %#v", request.Changes)
	}
}

func TestICMPUpdateParameters(t *testing.T) {
	before := validRule()
	after := before
	after.Protocol = "icmpv4:3,4"
	request, err := buildUpdateRequest(before, after)
	if err != nil {
		t.Fatal(err)
	}
	if request.Changes["Protocol"] != "icmpv4" || request.Changes["IcmpType"] != "3:4" {
		t.Fatalf("changes = %#v", request.Changes)
	}
}

func TestUpdatePreservesOriginalIdentity(t *testing.T) {
	before := validRule()
	before.Name = " 测试规则 "
	after := before
	after.Enabled = false
	request, err := buildUpdateRequest(before, after)
	if err != nil || request.Original.Name != before.Name {
		t.Fatalf("request = %#v, error = %v", request, err)
	}
}

func TestUpdateScriptRejectsPolicyOnlyRule(t *testing.T) {
	mock := `function Get-NetFirewallRule { param($PolicyStore,$ErrorAction,[switch]$TracePolicyStore)
  if ($PolicyStore -eq 'ActiveStore') { [PSCustomObject]@{ PolicyStoreSource='PersistentStore';PolicyStoreSourceType='Local'; DisplayName = '测试规则'; Direction = 'Inbound'; Name = 'id0' } }
}
function Set-NetFirewallRule { throw 'unexpected update' }
`
	request, _ := json.Marshal(updateRequest{Original: validRule(), Changes: map[string]interface{}{"Enabled": false}})
	output, err := runUpdateScript(mock+updateRuleScript, map[string]string{"FWM_RULE_UPDATE": string(request)})
	if err == nil || !strings.Contains(string(output), "不是唯一可编辑的本地规则") {
		t.Fatalf("error = %v, output = %s", err, output)
	}
}

func TestUpdateInputIsolationAndErrors(t *testing.T) {
	before := validRule()
	before.Name = `规则 ' $(); [*]`
	after := before
	after.Enabled = false
	runErr := errors.New("mock failure")
	err := updateRuleWith(before, after, func() bool { return true }, func(script string, env map[string]string) ([]byte, error) {
		if strings.Contains(script, before.Name) || strings.Contains(script, "New-NetFirewallRule") || strings.Contains(script, "Remove-NetFirewallRule") {
			t.Fatal("unsafe script")
		}
		var request updateRequest
		if err := json.Unmarshal([]byte(env["FWM_RULE_UPDATE"]), &request); err != nil {
			t.Fatal(err)
		}
		if request.Original.Name != before.Name || request.Changes["Enabled"] != false {
			t.Fatalf("request: %#v", request)
		}
		return []byte("access denied"), runErr
	})
	if !errors.Is(err, runErr) || !strings.Contains(err.Error(), "access denied") {
		t.Fatalf("error = %v", err)
	}
	if err := updateRuleWith(before, after, func() bool { return false }, func(string, map[string]string) ([]byte, error) { t.Fatal("executed without admin"); return nil, nil }); err == nil {
		t.Fatal("missing admin error")
	}
}

// PowerShell 命令由测试函数完全替换，只运行脚本控制流，不访问真实防火墙。
func TestUpdateScriptUsesUniqueObject(t *testing.T) {
	for _, count := range []int{0, 1, 2} {
		t.Run(string(rune('0'+count)), func(t *testing.T) {
			mock := `function Get-NetFirewallRule { param($PolicyStore,$ErrorAction,[switch]$TracePolicyStore)
  for ($i = 0; $i -lt [int]$env:FWM_TEST_COUNT; $i++) { [PSCustomObject]@{ PolicyStoreSource='PersistentStore';PolicyStoreSourceType='Local'; DisplayName = '测试规则'; Direction = 'Inbound'; Name = "id$i" } }
}
function Set-NetFirewallRule { param($InputObject,$ErrorAction,$Enabled)
  if ($InputObject.Name -ne 'id0' -or $Enabled -ne 'False') { throw 'wrong target or parameters' }
  $script:called = $true
}
`
			request, _ := json.Marshal(updateRequest{Original: validRule(), Changes: map[string]interface{}{"Enabled": false}})
			output, err := runUpdateScript(mock+updateRuleScript+"\nif (-not $script:called) { throw 'update not called' }", map[string]string{"FWM_RULE_UPDATE": string(request), "FWM_TEST_COUNT": string(rune('0' + count))})
			if count > 0 && err != nil {
				t.Fatalf("unique update: %v %s", err, output)
			}
			if count == 0 && err == nil {
				t.Fatal("ambiguous or missing target accepted")
			}
			if count == 0 && !strings.Contains(string(output), "原规则不存在") {
				t.Fatalf("missing diagnostic: %s", output)
			}
		})
	}
}

func TestAddRuleRequiresAdminBeforeCommand(t *testing.T) {
	err := addRuleWith(validRule(), func() bool { return false }, func([]string) ([]byte, error) { t.Fatal("executed system command"); return nil, nil })
	if err == nil || !strings.Contains(err.Error(), "管理员权限") {
		t.Fatalf("missing admin guard: %v", err)
	}
}

func TestUpdateRetainsSupportedExtendedAddresses(t *testing.T) {
	for _, address := range []string{"Internet", "Intranet", "LocalSubnet4", "DNS6", "CaptivePortal", "IntranetRemoteAccess", "PlayToDevice"} {
		original := validRule()
		original.RemoteAddr = address
		updated := original
		updated.Enabled = !original.Enabled
		request, err := buildUpdateRequest(original, updated)
		if err != nil {
			t.Fatalf("%s: %v", address, err)
		}
		if _, ok := request.Changes["RemoteAddress"]; ok {
			t.Fatalf("unchanged address changed: %#v", request)
		}
		updated.RemoteAddr = "Internet6,LocalSubnet4"
		if _, err := buildUpdateRequest(original, updated); err != nil {
			t.Fatal(err)
		}
	}
	rule := validRule()
	rule.RemoteAddr = "UnknownNetwork"
	if _, err := buildUpdateRequest(rule, rule); err == nil {
		t.Fatal("unknown keyword accepted")
	}
	rule.RemoteAddr = "Internet"
	if _, err := normalizeRule(rule); err == nil {
		t.Fatal("netsh support assumed")
	}
}

func TestUpdateScriptReturnsReadableFailure(t *testing.T) {
	output, err := runUpdateScript("throw 'mock rule failure'", nil)
	if err == nil || !strings.Contains(string(output), "mock rule failure") || strings.Contains(string(output), "CLIXML") || strings.Contains(string(output), "_x000") {
		t.Fatalf("unreadable error: %s %v", output, err)
	}
	if _, err := normalizeAddressesWith("Internet64", true, true); err == nil {
		t.Fatal("invalid keyword suffix accepted")
	}
}
