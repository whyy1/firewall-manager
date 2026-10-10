package smb

import (
	"errors"
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

func TestOpenLocationRequestsVisibleExplorerWithoutShell(t *testing.T) {
	for _, path := range []string{`C:\团队资料\带空格的目录`, `\\NAS\团队资料`, `C:\资料\逗号,与&符号`} {
		called := false
		err := openLocationWith("  "+path+"  ", func(location string) error {
			if location != path {
				t.Fatalf("validation location = %q", location)
			}
			return nil
		}, func(cmd *exec.Cmd) error {
			called = true
			if !reflect.DeepEqual(cmd.Args, []string{"explorer.exe", path}) {
				t.Fatalf("arguments = %#v", cmd.Args)
			}
			if cmd.SysProcAttr == nil || cmd.SysProcAttr.HideWindow {
				t.Fatal("Explorer window is hidden")
			}
			return nil
		})
		if err != nil || !called {
			t.Fatalf("open request = %v, called = %v", err, called)
		}
	}
}

func TestOpenLocationRejectsBeforeLaunchAndPreservesLaunchError(t *testing.T) {
	denied := errors.New("invalid directory")
	err := openLocationWith(`\\?\C:\invalid`, func(string) error { return denied }, func(*exec.Cmd) error {
		t.Fatal("invalid path launched Explorer")
		return nil
	})
	if !errors.Is(err, denied) {
		t.Fatalf("validation error = %v", err)
	}
	failed := errors.New("access denied")
	err = openLocationWith(`\\NAS\团队资料`, func(string) error { return nil }, func(*exec.Cmd) error { return failed })
	if !errors.Is(err, failed) || !strings.Contains(err.Error(), `\\NAS\团队资料`) {
		t.Fatalf("launch error = %v", err)
	}
}
