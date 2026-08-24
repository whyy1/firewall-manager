package smb

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	resourceConnected  = 1
	resourceTypeDisk   = 1
	errorMoreData      = 234
	errorNoMoreItems   = 259
	networkScanTimeout = 8 * time.Second
	networkBufferSize  = 64 * 1024
)

var (
	mprDLL             = syscall.NewLazyDLL("mpr.dll")
	procWNetOpenEnum   = mprDLL.NewProc("WNetOpenEnumW")
	procWNetEnum       = mprDLL.NewProc("WNetEnumResourceW")
	procWNetCloseEnum  = mprDLL.NewProc("WNetCloseEnum")
	networkScanMu      sync.Mutex
	networkScanRunning bool
)

type netResource struct {
	Scope       uint32
	Type        uint32
	DisplayType uint32
	Usage       uint32
	LocalName   *uint16
	RemoteName  *uint16
	Comment     *uint16
	Provider    *uint16
}

// NetworkShare describes a visible SMB-style disk share on the local network.
type NetworkShare struct {
	Server   string `json:"server"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	Comment  string `json:"comment"`
	Provider string `json:"provider"`
}

type networkScanResult struct {
	shares []NetworkShare
	err    error
}

// DiscoverNetworkShares enumerates visible disk shares without reading files.
func DiscoverNetworkShares(ctx context.Context) ([]NetworkShare, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	scanCtx, cancel := context.WithTimeout(ctx, networkScanTimeout)
	defer cancel()

	networkScanMu.Lock()
	if networkScanRunning {
		networkScanMu.Unlock()
		return nil, errors.New("已连接共享检测仍在进行，请稍候")
	}
	networkScanRunning = true
	networkScanMu.Unlock()

	resultChannel := make(chan networkScanResult, 1)
	go func() {
		defer func() {
			networkScanMu.Lock()
			networkScanRunning = false
			networkScanMu.Unlock()
		}()
		shares, err := discoverNetworkShares()
		resultChannel <- networkScanResult{shares: shares, err: err}
	}()

	select {
	case <-scanCtx.Done():
		return nil, errors.New("检测已连接共享超时，请稍后重试")
	case result := <-resultChannel:
		return result.shares, result.err
	}
}

func discoverNetworkShares() ([]NetworkShare, error) {
	found := make(map[string]NetworkShare)
	if err := enumerateConnectedNetworkResources(found); err != nil {
		return nil, fmt.Errorf("检测已连接共享失败: %w", err)
	}
	result := make([]NetworkShare, 0, len(found))
	for _, share := range found {
		result = append(result, share)
	}
	sort.Slice(result, func(i, j int) bool {
		if strings.EqualFold(result[i].Server, result[j].Server) {
			return strings.ToLower(result[i].Name) < strings.ToLower(result[j].Name)
		}
		return strings.ToLower(result[i].Server) < strings.ToLower(result[j].Server)
	})
	return result, nil
}

func enumerateConnectedNetworkResources(found map[string]NetworkShare) error {
	var handle uintptr
	status, _, _ := procWNetOpenEnum.Call(
		resourceConnected,
		resourceTypeDisk,
		0,
		0,
		uintptr(unsafe.Pointer(&handle)),
	)
	if status != 0 {
		return syscall.Errno(status)
	}
	defer procWNetCloseEnum.Call(handle)

	buffer := make([]byte, networkBufferSize)
	for {
		count := uint32(0xffffffff)
		size := uint32(len(buffer))
		status, _, _ = procWNetEnum.Call(
			handle,
			uintptr(unsafe.Pointer(&count)),
			uintptr(unsafe.Pointer(&buffer[0])),
			uintptr(unsafe.Pointer(&size)),
		)
		if status == errorNoMoreItems {
			return nil
		}
		if status == errorMoreData {
			buffer = make([]byte, size)
			continue
		}
		if status != 0 {
			return syscall.Errno(status)
		}

		entrySize := unsafe.Sizeof(netResource{})
		for index := uint32(0); index < count; index++ {
			entry := (*netResource)(unsafe.Pointer(uintptr(unsafe.Pointer(&buffer[0])) + uintptr(index)*entrySize))
			remoteName := utf16PointerToString(entry.RemoteName)
			if share, ok := networkShareFromPath(remoteName, utf16PointerToString(entry.Comment), utf16PointerToString(entry.Provider)); ok {
				found[strings.ToLower(share.Path)] = share
			}
		}
	}
}

func networkShareFromPath(path, comment, provider string) (NetworkShare, bool) {
	cleaned := strings.TrimSpace(path)
	if !strings.HasPrefix(cleaned, `\\`) {
		return NetworkShare{}, false
	}
	parts := strings.Split(strings.TrimPrefix(cleaned, `\\`), `\`)
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return NetworkShare{}, false
	}
	uncPath := `\\` + parts[0] + `\` + parts[1]
	return NetworkShare{
		Server:   parts[0],
		Name:     parts[1],
		Path:     uncPath,
		Comment:  comment,
		Provider: provider,
	}, true
}

func utf16PointerToString(pointer *uint16) string {
	if pointer == nil {
		return ""
	}
	return windows.UTF16PtrToString(pointer)
}

// OpenLocation opens a validated local directory or UNC share in Explorer.
func OpenLocation(location string) error {
	location = strings.TrimSpace(location)
	if location == "" {
		return errors.New("打开位置不能为空")
	}
	if strings.HasPrefix(location, `\\`) {
		if _, ok := networkShareFromPath(location, "", ""); !ok {
			return errors.New("网络共享地址格式无效")
		}
	} else {
		if !filepath.IsAbs(location) {
			return errors.New("本机目录必须是绝对路径")
		}
		info, err := os.Stat(location)
		if err != nil {
			return fmt.Errorf("本机目录不可用: %w", err)
		}
		if !info.IsDir() {
			return errors.New("打开位置必须是目录")
		}
	}
	cmd := exec.Command("explorer.exe", filepath.Clean(location))
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("打开资源管理器失败: %w", err)
	}
	if err := cmd.Process.Release(); err != nil {
		return fmt.Errorf("释放资源管理器进程句柄失败: %w", err)
	}
	return nil
}
