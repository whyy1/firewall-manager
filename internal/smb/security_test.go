package smb

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func mockBackend(admin bool, run func(string, map[string]string) ([]byte, error)) backend {
	return backend{
		isAdmin: func() bool { return admin },
		run: func(_ context.Context, _ time.Duration, script string, env map[string]string) ([]byte, error) {
			return run(script, env)
		},
	}
}

func TestWriteOperationsRequireAdminBeforeCommands(t *testing.T) {
	b := mockBackend(false, func(string, map[string]string) ([]byte, error) {
		t.Fatal("non-admin request executed a command")
		return nil, nil
	})
	operations := map[string]func() error{
		"start":         func() error { return b.SetServiceRunning(nil, true) },
		"stop":          func() error { return b.SetServiceRunning(nil, false) },
		"disable SMB1":  func() error { return b.DisableSMB1(nil) },
		"create":        func() error { return b.CreateShare(nil, CreateShareRequest{}) },
		"remove":        func() error { return b.RemoveShare(nil, "Data") },
		"close session": func() error { return b.CloseSession(nil, "1") },
		"firewall":      func() error { return b.SetFirewallEnabled(nil, true) },
	}
	for name, operation := range operations {
		t.Run(name, func(t *testing.T) {
			if err := operation(); err == nil || !strings.Contains(err.Error(), "管理员权限") {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestCreateRejectsUnsafePathsBeforeFilesystemAccess(t *testing.T) {
	for _, path := range []string{
		`\\missing-host\share`, `\\?\C:\Data`, `\\.\C:\Data`, `\\?\UNC\host\share`,
		`\??\C:\Data`, `C:relative`, `\rooted`, `relative`, `C:\Data:stream`,
	} {
		t.Run(path, func(t *testing.T) {
			err := validateCreateRequest(normalizeRequest(CreateShareRequest{Name: "Data", Path: path, AccessMode: "read"}))
			if err == nil || strings.Contains(err.Error(), "不可用") {
				t.Fatalf("path should be rejected before os.Stat: %v", err)
			}
		})
	}
}

func TestOpenLocationValidation(t *testing.T) {
	for _, path := range []string{`\\NAS\Public`, `\\NAS\Public\Folder`, `\\NAS\C$`, t.TempDir()} {
		if err := validateOpenLocation(path); err != nil {
			t.Errorf("validateOpenLocation(%q) = %v", path, err)
		}
	}
	for _, path := range []string{
		"", `relative`, `\\NAS`, `\\NAS\`, `\\?\C:\Data`, `\\.\pipe\name`,
		`\\NAS\Public\..\Other`, `\\NAS\Public\\Other`, `\\NAS\Public:stream`,
		`\\NAS\Public\bad?`, `\\NAS\Public\bad.`, "\\\\NAS\\Public\\bad\x00",
	} {
		if err := validateOpenLocation(path); err == nil {
			t.Errorf("validateOpenLocation(%q) accepted invalid path", path)
		}
	}
	file := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(file, nil, 0600); err != nil {
		t.Fatal(err)
	}
	if err := validateOpenLocation(file); err == nil {
		t.Fatal("accepted a file as a directory")
	}
}

func TestProtectedSharesRejectedBeforeCommands(t *testing.T) {
	b := mockBackend(true, func(string, map[string]string) ([]byte, error) {
		t.Fatal("protected share request executed a command")
		return nil, nil
	})
	for _, name := range []string{"ADMIN$", "IPC$", "print$", "C$"} {
		if err := b.RemoveShare(nil, name); err == nil {
			t.Errorf("RemoveShare(%q) accepted protected name", name)
		}
	}
	if !strings.Contains(removeShareScript, "if ($share.Special)") || !strings.Contains(removeShareScript, "$share | Remove-SmbShare") {
		t.Fatal("removal script must check actual Special property before object removal")
	}
}

func TestShareInputOnlyPassedThroughEnvironment(t *testing.T) {
	request := CreateShareRequest{Name: "Team's files", Path: t.TempDir(), Description: "'; Stop-Service LanmanServer; ' 中文", AccessMode: "read"}
	b := mockBackend(true, func(script string, env map[string]string) ([]byte, error) {
		if script != createShareScript || strings.Contains(script, request.Description) || strings.Contains(script, request.Name) {
			t.Fatal("request was interpolated into script")
		}
		if env["FWM_SMB_SHARE_DESCRIPTION"] != request.Description || env["FWM_SMB_SHARE_NAME"] != request.Name || env["FWM_SMB_SHARE_PATH"] != request.Path || env["FWM_SMB_ACCESS_MODE"] != "read" {
			t.Fatalf("unexpected environment: %#v", env)
		}
		return nil, nil
	})
	if err := b.CreateShare(nil, request); err != nil {
		t.Fatal(err)
	}
}

func TestSessionIDValidationWithoutPrivilegeDependency(t *testing.T) {
	for _, id := range []string{"", "0", "-1", "+1", "1.0", "1;Stop-Service", "18446744073709551616", " 1 "} {
		if err := validateSessionID(id); err == nil {
			t.Errorf("accepted invalid session ID %q", id)
		}
	}
	for _, id := range []string{"1", "18446744073709551615"} {
		if err := validateSessionID(id); err != nil {
			t.Errorf("rejected valid session ID %q: %v", id, err)
		}
	}
	b := mockBackend(true, func(string, map[string]string) ([]byte, error) {
		t.Fatal("invalid session ID executed a command")
		return nil, nil
	})
	if err := b.CloseSession(nil, "not-a-number"); err == nil || !strings.Contains(err.Error(), "会话 ID 无效") {
		t.Fatalf("CloseSession error = %v", err)
	}
}

func TestArrayResponsesThroughBackend(t *testing.T) {
	for _, payload := range []string{"[]", "null", `[{"name":"Data","sessionId":"123"}]`} {
		t.Run(payload, func(t *testing.T) {
			b := mockBackend(false, func(string, map[string]string) ([]byte, error) { return []byte(payload), nil })
			shares, err := b.GetShares(nil)
			if err != nil || shares == nil {
				t.Fatalf("GetShares = %#v, %v", shares, err)
			}
			sessions, err := b.GetSessions(nil)
			if err != nil || sessions == nil {
				t.Fatalf("GetSessions = %#v, %v", sessions, err)
			}
			if payload[0] == '[' && len(payload) > 2 && (len(shares) != 1 || len(sessions) != 1) {
				t.Fatal("single item response did not decode as array")
			}
		})
	}
}

func TestQueryFailuresRemainIndependent(t *testing.T) {
	failure := errors.New("access denied")
	b := mockBackend(false, func(script string, _ map[string]string) ([]byte, error) {
		if script == sharesScript {
			return nil, failure
		}
		return []byte(`{"serviceStatus":"Running","serviceRunning":true}`), nil
	})
	if _, err := b.GetShares(nil); !errors.Is(err, failure) {
		t.Fatalf("GetShares error = %v", err)
	}
	overview, err := b.GetOverview(nil)
	if err != nil || !overview.ServiceRunning {
		t.Fatalf("GetOverview = %#v, %v", overview, err)
	}
}

func TestWriteFailuresPreserveOperationAndCause(t *testing.T) {
	failure := errors.New("simulated failure")
	b := mockBackend(true, func(string, map[string]string) ([]byte, error) { return nil, failure })
	operations := []struct {
		name string
		run  func() error
	}{
		{"启动 SMB", func() error { return b.SetServiceRunning(nil, true) }},
		{"关闭 SMB1", func() error { return b.DisableSMB1(nil) }},
		{"创建 SMB", func() error {
			return b.CreateShare(nil, CreateShareRequest{Name: "Data", Path: t.TempDir(), AccessMode: "read"})
		}},
		{"删除 SMB", func() error { return b.RemoveShare(nil, "Data") }},
		{"断开 SMB", func() error { return b.CloseSession(nil, "1") }},
		{"设置 SMB", func() error { return b.SetFirewallEnabled(nil, true) }},
	}
	for _, operation := range operations {
		err := operation.run()
		if !errors.Is(err, failure) || !strings.Contains(err.Error(), operation.name) {
			t.Errorf("%s error = %v", operation.name, err)
		}
	}
}
