package ports

import (
	"errors"
	"strings"
	"testing"
)

func TestServicePortQueryErrors(t *testing.T) {
	for _, failAt := range []int{1, 2, 3} {
		calls := 0
		failure := errors.New("mock command failure")
		result, err := getServicePortsWith(func(string, ...string) ([]byte, error) {
			calls++
			if calls == failAt {
				return nil, failure
			}
			return []byte(""), nil
		})
		if !errors.Is(err, failure) || result != nil || calls != failAt {
			t.Fatalf("failure %d: result=%v error=%v calls=%d", failAt, result, err, calls)
		}
	}
}

func TestServicePortCapabilitiesAndDetection(t *testing.T) {
	result, err := getServicePortsWith(func(name string, args ...string) ([]byte, error) {
		if name == "tasklist" {
			return []byte(`"svchost.exe","123","TermService, WinRM"`), nil
		}
		return []byte("  TCP  0.0.0.0:3390  0.0.0.0:0  LISTENING  123\n"), nil
	})
	if err != nil || len(result) != len(commonServices) {
		t.Fatalf("%v %v", result, err)
	}
	for _, service := range result {
		want := service.ServiceName == "TermService" || service.ServiceName == "WinRM"
		if service.CanChangePort != want {
			t.Fatalf("wrong capability: %#v", service)
		}
		if service.ServiceName == "TermService" && (!service.Running || service.ListenPort != 3390) {
			t.Fatalf("wrong detection: %#v", service)
		}
		if service.ServiceName == "Dnscache" && service.ListenPort != 0 {
			t.Fatalf("default substituted: %#v", service)
		}
	}
}

func TestChangeServicePortRejectsInvalidPorts(t *testing.T) {
	for _, port := range []int{0, -1, 65536} {
		if err := ChangeServicePort("TermService", port); err == nil {
			t.Fatal("accepted invalid port")
		}
	}
}

func TestSMBPortChangeRejectedBeforeSystemOperations(t *testing.T) {
	for _, port := range []int{445, 1445, 0} {
		err := ChangeServicePort("LanmanServer", port)
		if err == nil || !strings.Contains(err.Error(), "不支持直接修改") {
			t.Fatalf("unsupported SMB modification accepted: %v", err)
		}
	}
}
