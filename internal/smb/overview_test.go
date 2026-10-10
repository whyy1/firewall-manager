package smb

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// Both system cmdlets are replaced; this test never reads or changes real services.
func TestOverviewKeepsServiceWhenConfigurationUnavailable(t *testing.T) {
	for _, scenario := range []struct {
		name, status string
		fail         bool
	}{
		{"stopped", "Stopped", true}, {"configuration failure", "Running", true}, {"running", "Running", false},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			mock := `function Get-Service { param($Name,$ErrorAction)
 [PSCustomObject]@{Status=$env:FWM_TEST_STATUS;StartType='Manual'}
}
function Get-SmbServerConfiguration { param($ErrorAction)
 if ($env:FWM_TEST_STATUS -eq 'Stopped') { throw 'stopped configuration must not be queried' }
 if ($env:FWM_TEST_FAIL -eq 'true') { throw 'Windows System Error 111' }
 [PSCustomObject]@{EnableSMB1Protocol=$true;EnableSMB2Protocol=$true;RequireSecuritySignature=$true;EncryptData=$false}
}
`
			failure := "false"
			if scenario.fail {
				failure = "true"
			}
			output, err := runPowerShell(context.Background(), readTimeout, mock+overviewScript, map[string]string{"FWM_TEST_STATUS": scenario.status, "FWM_TEST_FAIL": failure})
			if err != nil {
				t.Fatal(err)
			}
			var result Overview
			if err := json.Unmarshal(output, &result); err != nil {
				t.Fatal(err)
			}
			if result.ServiceRunning != (scenario.status == "Running") || result.ServiceStatus != scenario.status || result.StartupType != "Manual" {
				t.Fatalf("service state lost: %#v", result)
			}
			if scenario.fail != (result.ConfigurationError != "") {
				t.Fatalf("configuration availability: %#v", result)
			}
			if scenario.status == "Stopped" && strings.Contains(result.ConfigurationError, "must not be queried") {
				t.Fatal("queried stopped provider")
			}
			if !scenario.fail && !result.SMB1Enabled {
				t.Fatal("successful configuration lost")
			}
		})
	}
}

func TestPowerShellCLIXMLErrorIsReadable(t *testing.T) {
	cause := errors.New("exit status 1")
	payload := `#< CLIXML
<Objs xmlns="http://schemas.microsoft.com/powershell/2004/04"><S S="Error">Get-SmbServerConfiguration : The file name is too long._x000D__x000A_</S><S S="Error">Windows System Error 111_x000D__x000A_</S><S S="Warning">ignored</S></Objs>`
	_, err := commandResult(nil, []byte(payload), cause)
	if !errors.Is(err, cause) || !strings.Contains(err.Error(), "Windows System Error 111") || strings.Contains(err.Error(), "CLIXML") || strings.Contains(err.Error(), "_x000") || strings.Contains(err.Error(), "ignored") {
		t.Fatalf("unreadable diagnostic: %v", err)
	}
	if powerShellDiagnostic([]byte("#< CLIXML\n<broken")) != "#< CLIXML\n<broken" {
		t.Fatal("malformed diagnostic lost")
	}
}
