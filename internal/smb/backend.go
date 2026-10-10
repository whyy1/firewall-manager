package smb

import (
	"context"
	"time"

	"firewall-manager/internal/admin"
)

// backend keeps command and privilege dependencies explicit for isolated tests.
type backend struct {
	isAdmin func() bool
	run     func(context.Context, time.Duration, string, map[string]string) ([]byte, error)
}

func localBackend() backend {
	return backend{isAdmin: admin.IsAdmin, run: runPowerShell}
}

// GetOverview returns the local SMB service and protocol state.
func GetOverview(ctx context.Context) (Overview, error) {
	return localBackend().GetOverview(ctx)
}

// GetShares returns all local shares, including protected system shares.
func GetShares(ctx context.Context) ([]Share, error) {
	return localBackend().GetShares(ctx)
}

// SetServiceRunning starts or stops the local SMB server service.
func SetServiceRunning(ctx context.Context, running bool) error {
	return localBackend().SetServiceRunning(ctx, running)
}

// DisableSMB1 disables the deprecated SMB1 server protocol.
func DisableSMB1(ctx context.Context) error {
	return localBackend().DisableSMB1(ctx)
}

// CreateShare creates a local share without changing filesystem permissions.
func CreateShare(ctx context.Context, request CreateShareRequest) error {
	return localBackend().CreateShare(ctx, request)
}

// RemoveShare removes a non-special local SMB share.
func RemoveShare(ctx context.Context, name string) error {
	return localBackend().RemoveShare(ctx, name)
}

// GetSessions returns inbound SMB sessions to this computer.
func GetSessions(ctx context.Context) ([]Session, error) {
	return localBackend().GetSessions(ctx)
}

// CloseSession forcibly closes one inbound SMB session.
func CloseSession(ctx context.Context, sessionID string) error {
	return localBackend().CloseSession(ctx, sessionID)
}

// GetFirewallStatus returns the built-in sharing firewall group state.
func GetFirewallStatus(ctx context.Context) (FirewallStatus, error) {
	return localBackend().GetFirewallStatus(ctx)
}

// SetFirewallEnabled toggles the sharing rule group for all profiles.
func SetFirewallEnabled(ctx context.Context, enabled bool) error {
	return localBackend().SetFirewallEnabled(ctx, enabled)
}
