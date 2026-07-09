package ports

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

// newHiddenCmd 创建隐藏控制台窗口的命令
func newHiddenCmd(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd
}

// ServicePort 常用服务端口信息
type ServicePort struct {
	Name        string `json:"name"`        // 服务名称
	ServiceName string `json:"serviceName"` // Windows 服务名
	DefaultPort int    `json:"defaultPort"` // 默认端口
	Protocol    string `json:"protocol"`    // 协议
	Description string `json:"description"` // 描述
	Running     bool   `json:"running"`     // 服务是否运行中
	ListenPort  int    `json:"listenPort"`  // 实际监听端口（0=未监听）
}

// 预定义的常用 Windows 服务端口
var commonServices = []ServicePort{
	{Name: "RDP 远程桌面", ServiceName: "TermService", DefaultPort: 3389, Protocol: "tcp", Description: "Windows 远程桌面连接"},
	{Name: "SMB 文件共享", ServiceName: "LanmanServer", DefaultPort: 445, Protocol: "tcp", Description: "SMB 网络文件共享"},
	{Name: "WinRM 远程管理", ServiceName: "WinRM", DefaultPort: 5985, Protocol: "tcp", Description: "Windows 远程管理服务"},
	{Name: "DNS 域名解析", ServiceName: "Dnscache", DefaultPort: 53, Protocol: "udp", Description: "域名系统解析服务"},
	{Name: "DHCP 客户端", ServiceName: "Dhcp", DefaultPort: 67, Protocol: "udp", Description: "动态主机配置协议"},
	{Name: "打印服务", ServiceName: "Spooler", DefaultPort: 515, Protocol: "tcp", Description: "打印后台处理程序"},
	{Name: "远程注册表", ServiceName: "RemoteRegistry", DefaultPort: 445, Protocol: "tcp", Description: "远程注册表服务"},
}

// GetServicePorts 获取常用端口列表，检测服务状态和实际监听端口
func GetServicePorts() []ServicePort {
	result := make([]ServicePort, len(commonServices))
	copy(result, commonServices)

	svcPids := getServicePids()
	pidPorts := getPidPorts()
	portPids := getPortPids()

	for i := range result {
		svc := &result[i]
		pid, ok := svcPids[svc.ServiceName]
		svc.Running = ok && pid > 0

		if svc.Running {
			if ports, found := pidPorts[pid]; found && len(ports) > 0 {
				svc.ListenPort = ports[0]
			} else if _, found := portPids[svc.DefaultPort]; found {
				svc.ListenPort = svc.DefaultPort
			}
		} else {
			if _, found := portPids[svc.DefaultPort]; found {
				svc.ListenPort = svc.DefaultPort
			}
		}
	}

	return result
}

// ChangeServicePort 修改服务端口
func ChangeServicePort(serviceName string, newPort int) error {
	switch serviceName {
	case "TermService":
		return changeRDPPort(newPort)
	case "WinRM":
		return changeWinRMPort(newPort)
	case "LanmanServer":
		return changeSMBPort(newPort)
	default:
		return fmt.Errorf("暂不支持修改此服务的端口")
	}
}

func changeRDPPort(newPort int) error {
	cmd := newHiddenCmd("reg", "add",
		`HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Control\Terminal Server\WinStations\RDP-Tcp`,
		"/v", "PortNumber", "/t", "REG_DWORD", "/d", fmt.Sprintf("%d", newPort), "/f")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("修改 RDP 端口失败: %w\n%s", err, string(output))
	}
	return restartService("TermService")
}

func changeWinRMPort(newPort int) error {
	cmd := newHiddenCmd("winrm", "set", "winrm/config/Listener?Address=*+Transport=HTTP",
		`@{Port="`+fmt.Sprintf("%d", newPort)+`"}`)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("修改 WinRM 端口失败: %w\n%s", err, string(output))
	}
	return restartService("WinRM")
}

func changeSMBPort(newPort int) error {
	cmd := newHiddenCmd("reg", "add",
		`HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Services\LanmanServer\Parameters`,
		"/v", "SmbPort", "/t", "REG_DWORD", "/d", fmt.Sprintf("%d", newPort), "/f")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("修改 SMB 端口失败: %w\n%s", err, string(output))
	}
	return restartService("LanmanServer")
}

func restartService(serviceName string) error {
	stopCmd := newHiddenCmd("net", "stop", serviceName, "/y")
	stopCmd.CombinedOutput()

	startCmd := newHiddenCmd("net", "start", serviceName)
	output, err := startCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("重启服务 %s 失败: %w\n%s", serviceName, err, string(output))
	}
	return nil
}

func getServicePids() map[string]int {
	result := make(map[string]int)
	cmd := newHiddenCmd("tasklist", "/svc", "/FO", "CSV", "/NH")
	output, err := cmd.Output()
	if err != nil {
		return result
	}

	re := regexp.MustCompile(`"([^"]+)","(\d+)","([^"]*)"`)
	for _, line := range strings.Split(string(output), "\n") {
		matches := re.FindStringSubmatch(strings.TrimSpace(line))
		if len(matches) >= 4 {
			pid, _ := strconv.Atoi(matches[2])
			svcs := strings.Split(matches[3], ",")
			for _, svc := range svcs {
				result[strings.TrimSpace(svc)] = pid
			}
		}
	}
	return result
}

func getPidPorts() map[int][]int {
	result := make(map[int][]int)
	cmd := newHiddenCmd("netstat", "-ano")
	output, err := cmd.Output()
	if err != nil {
		return result
	}

	re := regexp.MustCompile(`\s+TCP\s+\S+?:(\d+)\s+\S+\s+LISTENING\s+(\d+)`)
	for _, line := range strings.Split(string(output), "\n") {
		matches := re.FindStringSubmatch(line)
		if len(matches) >= 3 {
			port, _ := strconv.Atoi(matches[1])
			pid, _ := strconv.Atoi(matches[2])
			result[pid] = append(result[pid], port)
		}
	}
	return result
}

func getPortPids() map[int]int {
	result := make(map[int]int)
	cmd := newHiddenCmd("netstat", "-ano")
	output, err := cmd.Output()
	if err != nil {
		return result
	}

	re := regexp.MustCompile(`\s+TCP\s+\S+?:(\d+)\s+\S+\s+LISTENING\s+(\d+)`)
	for _, line := range strings.Split(string(output), "\n") {
		matches := re.FindStringSubmatch(line)
		if len(matches) >= 3 {
			port, _ := strconv.Atoi(matches[1])
			pid, _ := strconv.Atoi(matches[2])
			if _, exists := result[port]; !exists {
				result[port] = pid
			}
		}
	}
	return result
}
