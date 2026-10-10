package smb

import (
	"context"
	"os"
	"testing"
)

func TestNetworkShareFromPath(t *testing.T) {
	share, ok := networkShareFromPath(`\\NAS\Public\Subfolder`, "Team files", "Microsoft Windows Network")
	if !ok {
		t.Fatal("networkShareFromPath() rejected a valid UNC path")
	}
	if share.Server != "NAS" || share.Name != "Public" || share.Path != `\\NAS\Public` {
		t.Fatalf("networkShareFromPath() = %#v", share)
	}
}

func TestNetworkShareFromPathRejectsServerOnly(t *testing.T) {
	if _, ok := networkShareFromPath(`\\NAS`, "", ""); ok {
		t.Fatal("networkShareFromPath() accepted a server without a share")
	}
}

func TestCloseSessionRejectsInvalidID(t *testing.T) {
	for _, sessionID := range []string{"", "0", "-1", "not-a-number"} {
		if err := validateSessionID(sessionID); err == nil {
			t.Fatalf("validateSessionID(%q) returned nil", sessionID)
		}
	}
}

func TestNetworkDiscoveryIntegration(t *testing.T) {
	if os.Getenv("FWM_SMB_NETWORK_INTEGRATION") != "1" {
		t.Skip("set FWM_SMB_NETWORK_INTEGRATION=1 to enumerate visible network shares")
	}

	shares, err := DiscoverNetworkShares(context.Background())
	if err != nil {
		t.Fatalf("DiscoverNetworkShares() error = %v", err)
	}
	if shares == nil {
		t.Fatal("DiscoverNetworkShares() returned nil")
	}
	for _, share := range shares {
		if _, ok := networkShareFromPath(share.Path, share.Comment, share.Provider); !ok {
			t.Fatalf("DiscoverNetworkShares() returned invalid share %#v", share)
		}
	}
}
