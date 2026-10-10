package ports

import (
	"firewall-manager/internal/admin"
	"fmt"
	"os"
	"path/filepath"
)

type commandRunner func(string, ...string) ([]byte, error)

// ChangeServicePort modifies supported service configuration and restarts the service.
func ChangeServicePort(serviceName string, newPort int) error {
	return changeServicePortWith(serviceName, newPort, admin.IsAdmin, func(name string, args ...string) ([]byte, error) {
		return newHiddenCmd(name, args...).CombinedOutput()
	})
}

func changeServicePortWith(serviceName string, newPort int, isAdmin func() bool, run commandRunner) error {
	if serviceName == "LanmanServer" {
		return fmt.Errorf("Windows SMB TCP 监听端口固定为 445，不支持直接修改；未写入配置或重启服务")
	}
	if serviceName != "TermService" && serviceName != "WinRM" {
		return fmt.Errorf("暂不支持修改此服务的端口")
	}
	if newPort < 1 || newPort > 65535 {
		return fmt.Errorf("端口必须是 1–65535 的整数")
	}
	if !isAdmin() {
		return fmt.Errorf("修改服务端口需要管理员权限")
	}
	var output []byte
	var err error
	if serviceName == "TermService" {
		output, err = run("reg", "add", `HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Control\Terminal Server\WinStations\RDP-Tcp`, "/v", "PortNumber", "/t", "REG_DWORD", "/d", fmt.Sprintf("%d", newPort), "/f")
	} else {
		output, err = run("cscript.exe", "//nologo", filepath.Join(os.Getenv("SystemRoot"), "System32", "winrm.vbs"), "set", "winrm/config/Listener?Address=*+Transport=HTTP", `@{Port="`+fmt.Sprintf("%d", newPort)+`"}`)
	}
	if err != nil {
		return fmt.Errorf("修改 %s 端口配置失败: %w\n%s", serviceName, err, string(output))
	}
	if err := restartServiceWith(serviceName, run); err != nil {
		return fmt.Errorf("端口配置已写入，但服务重启失败: %w；请检查服务和连接状态，恢复原端口或手动启动服务后刷新，不要重复提交", err)
	}
	return nil
}

func restartServiceWith(serviceName string, run commandRunner) error {
	output, err := run("net", "stop", serviceName, "/y")
	if err != nil {
		return fmt.Errorf("停止服务 %s 失败: %w\n%s", serviceName, err, string(output))
	}
	output, err = run("net", "start", serviceName)
	if err != nil {
		return fmt.Errorf("启动服务 %s 失败: %w\n%s", serviceName, err, string(output))
	}
	return nil
}
