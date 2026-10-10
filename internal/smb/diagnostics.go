package smb

import "firewall-manager/internal/diagnostics"

func powerShellDiagnostic(output []byte) string {
	return diagnostics.PowerShell(output)
}
