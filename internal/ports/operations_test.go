package ports

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestPortChangesRejectBeforeCommands(t *testing.T) {
	for _, item := range []struct {
		service string
		port    int
		admin   bool
	}{
		{"TermService", 0, true}, {"TermService", 3390, false}, {"LanmanServer", 1445, true}, {"Dnscache", 5353, true},
	} {
		err := changeServicePortWith(item.service, item.port, func() bool { return item.admin }, func(string, ...string) ([]byte, error) { t.Fatal("unexpected system command"); return nil, nil })
		if err == nil {
			t.Fatalf("accepted invalid request: %#v", item)
		}
	}
}

func TestPortChangeFailuresStopSequenceAndPreserveCause(t *testing.T) {
	for _, service := range []string{"TermService", "WinRM"} {
		for _, failAt := range []int{0, 1, 2, 3} {
			calls := [][]string{}
			cause := errors.New("mock command failure")
			err := changeServicePortWith(service, 3390, func() bool { return true }, func(name string, args ...string) ([]byte, error) {
				calls = append(calls, append([]string{name}, args...))
				if len(calls) == failAt {
					return []byte("mock diagnostic"), cause
				}
				return nil, nil
			})
			wantCalls := 3
			if failAt > 0 {
				wantCalls = failAt
			}
			if len(calls) != wantCalls {
				t.Fatalf("%s failure %d: calls %#v", service, failAt, calls)
			}
			if failAt == 0 && err != nil {
				t.Fatal(err)
			}
			if failAt > 0 && (!errors.Is(err, cause) || !strings.Contains(err.Error(), "mock diagnostic")) {
				t.Fatalf("missing failure: %v", err)
			}
			if failAt > 1 && !strings.Contains(err.Error(), "端口配置已写入") {
				t.Fatalf("partial write hidden: %v", err)
			}
			if len(calls) > 1 && !reflect.DeepEqual(calls[1], []string{"net", "stop", service, "/y"}) {
				t.Fatal(calls)
			}
			if len(calls) > 2 && !reflect.DeepEqual(calls[2], []string{"net", "start", service}) {
				t.Fatal(calls)
			}
			if service == "TermService" && calls[0][0] != "reg" {
				t.Fatal(calls)
			}
			if service == "WinRM" && !reflect.DeepEqual(calls[0], []string{"cscript.exe", "//nologo", filepath.Join(os.Getenv("SystemRoot"), "System32", "winrm.vbs"), "set", "winrm/config/Listener?Address=*+Transport=HTTP", `@{Port="3390"}`}) {
				t.Fatal(calls)
			}
		}
	}
}
