package smb

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateCreateRequest(t *testing.T) {
	directory := t.TempDir()
	tests := []struct {
		name    string
		request CreateShareRequest
		wantErr string
	}{
		{
			name: "valid read share",
			request: CreateShareRequest{
				Name:       "Documents",
				Path:       directory,
				AccessMode: "read",
			},
		},
		{
			name: "invalid name",
			request: CreateShareRequest{
				Name:       "Bad/Name",
				Path:       directory,
				AccessMode: "read",
			},
			wantErr: "不允许的字符",
		},
		{
			name: "protected name",
			request: CreateShareRequest{
				Name:       "ADMIN$",
				Path:       directory,
				AccessMode: "read",
			},
			wantErr: "系统保留",
		},
		{
			name: "relative path",
			request: CreateShareRequest{
				Name:       "Documents",
				Path:       "relative",
				AccessMode: "read",
			},
			wantErr: "绝对路径",
		},
		{
			name: "missing path",
			request: CreateShareRequest{
				Name:       "Documents",
				Path:       filepath.Join(directory, "missing"),
				AccessMode: "read",
			},
			wantErr: "不可用",
		},
		{
			name: "invalid access",
			request: CreateShareRequest{
				Name:       "Documents",
				Path:       directory,
				AccessMode: "owner",
			},
			wantErr: "共享权限",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateCreateRequest(test.request)
			if test.wantErr == "" && err != nil {
				t.Fatalf("validateCreateRequest() error = %v", err)
			}
			if test.wantErr != "" && (err == nil || !strings.Contains(err.Error(), test.wantErr)) {
				t.Fatalf("validateCreateRequest() error = %v, want substring %q", err, test.wantErr)
			}
		})
	}
}

func TestNormalizeRequest(t *testing.T) {
	request := normalizeRequest(CreateShareRequest{
		Name:        "  Documents  ",
		Path:        ` C:\Data\. `,
		Description: "  Team files  ",
		AccessMode:  " READ ",
	})

	if request.Name != "Documents" {
		t.Fatalf("Name = %q", request.Name)
	}
	if request.Path != `C:\Data` {
		t.Fatalf("Path = %q", request.Path)
	}
	if request.Description != "Team files" || request.AccessMode != "read" {
		t.Fatalf("normalized request = %#v", request)
	}
}

func TestSharesJSONAlwaysDecodesAsArray(t *testing.T) {
	payload := `[{"name":"Data","path":"C:\\Data","description":"","special":false,"state":"Online","concurrentUserLimit":0,"encryptData":false}]`
	var shares []Share
	if err := json.Unmarshal([]byte(payload), &shares); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if len(shares) != 1 || shares[0].Name != "Data" {
		t.Fatalf("shares = %#v", shares)
	}
}

func TestReservedShareNames(t *testing.T) {
	for _, name := range []string{"ADMIN$", "ipc$", "C$"} {
		if !isReservedShareName(name) {
			t.Errorf("isReservedShareName(%q) = false", name)
		}
	}
	for _, name := range []string{"Data", "Hidden$"} {
		if isReservedShareName(name) {
			t.Errorf("isReservedShareName(%q) = true", name)
		}
	}
}

func TestCommandResultKeepsDiagnosticsOutOfSuccessfulJSON(t *testing.T) {
	stdout := []byte(`{"serviceRunning":true}`)
	stderr := []byte("#< CLIXML\r\n<Objs>diagnostic</Objs>")

	result, err := commandResult(stdout, stderr, nil)
	if err != nil {
		t.Fatalf("commandResult() error = %v", err)
	}
	if string(result) != string(stdout) {
		t.Fatalf("commandResult() = %q, want %q", result, stdout)
	}
}

func TestCommandResultIncludesPowerShellError(t *testing.T) {
	runErr := errors.New("exit status 1")
	_, err := commandResult(nil, []byte("access denied"), runErr)
	if err == nil || !strings.Contains(err.Error(), "access denied") {
		t.Fatalf("commandResult() error = %v", err)
	}
}

func TestReadOnlyIntegration(t *testing.T) {
	if os.Getenv("FWM_SMB_INTEGRATION") != "1" {
		t.Skip("set FWM_SMB_INTEGRATION=1 in an elevated session")
	}

	overview, err := GetOverview(context.Background())
	if err != nil {
		t.Fatalf("GetOverview() error = %v", err)
	}
	if overview.ServiceStatus == "" {
		t.Fatal("GetOverview() returned an empty service status")
	}

	shares, err := GetShares(context.Background())
	if err != nil {
		t.Fatalf("GetShares() error = %v", err)
	}
	if shares == nil {
		t.Fatal("GetShares() returned nil")
	}

	sessions, err := GetSessions(context.Background())
	if err != nil {
		t.Fatalf("GetSessions() error = %v", err)
	}
	if sessions == nil {
		t.Fatal("GetSessions() returned nil")
	}

	firewallStatus, err := GetFirewallStatus(context.Background())
	if err != nil {
		t.Fatalf("GetFirewallStatus() error = %v", err)
	}
	if !firewallStatus.Available || len(firewallStatus.Profiles) != 3 {
		t.Fatalf("GetFirewallStatus() = %#v", firewallStatus)
	}
}
