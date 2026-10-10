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
	Name          string `json:"name"`          // 服务名称
	ServiceName   string `json:"serviceName"`   // Windows 服务名
	DefaultPort   int    `json:"defaultPort"`   // 默认端口
	Protocol      string `json:"protocol"`      // 协议
	Description   string `json:"description"`   // 描述
	Running       bool   `json:"running"`       // 服务是否运行中
	ListenPort    int    `json:"listenPort"`    // 实际监听端口（0=未监听）
	CanChangePort bool   `json:"canChangePort"` // 是否支持修改端口
}

// 预定义的常用 Windows 服务端口
var commonServices = []ServicePort{
	{Name: "RDP 远程桌面", ServiceName: "TermService", DefaultPort: 3389, Protocol: "tcp", Description: "Windows 远程桌面连接"},
	{Name: "SMB 文件共享", ServiceName: "LanmanServer", DefaultPort: 445, Protocol: "tcp", Description: "Windows SMB TCP 端口固定为 445，不能直接修改"},
	{Name: "WinRM 远程管理", ServiceName: "WinRM", DefaultPort: 5985, Protocol: "tcp", Description: "Windows 远程管理服务"},
	{Name: "DNS 域名解析", ServiceName: "Dnscache", DefaultPort: 53, Protocol: "udp", Description: "域名系统解析服务"},
	{Name: "DHCP 客户端", ServiceName: "Dhcp", DefaultPort: 67, Protocol: "udp", Description: "动态主机配置协议"},
	{Name: "打印服务", ServiceName: "Spooler", DefaultPort: 515, Protocol: "tcp", Description: "打印后台处理程序"},
	{Name: "远程注册表", ServiceName: "RemoteRegistry", DefaultPort: 445, Protocol: "tcp", Description: "远程注册表服务"},
}

// GetServicePorts 获取常用端口列表，检测服务状态和实际监听端口
func GetServicePorts() ([]ServicePort, error) {
	return getServicePortsWith(func(name string, args ...string) ([]byte, error) {
		return newHiddenCmd(name, args...).Output()
	})
}

func getServicePortsWith(run func(string, ...string) ([]byte, error)) ([]ServicePort, error) {
	result := make([]ServicePort, len(commonServices))
	copy(result, commonServices)

	svcPids, err := getServicePids(run)
	if err != nil {
		return nil, err
	}
	pidPorts, err := getPidPorts(run)
	if err != nil {
		return nil, err
	}
	portPids, err := getPortPids(run)
	if err != nil {
		return nil, err
	}

	for i := range result {
		svc := &result[i]
		svc.CanChangePort = svc.ServiceName == "TermService" || svc.ServiceName == "WinRM"
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

	return result, nil
}

func getServicePids(run func(string, ...string) ([]byte, error)) (map[string]int, error) {
	result := make(map[string]int)
	output, err := run("tasklist", "/svc", "/FO", "CSV", "/NH")
	if err != nil {
		return nil, fmt.Errorf("检测服务进程失败: %w", err)
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
	return result, nil
}

func getPidPorts(run func(string, ...string) ([]byte, error)) (map[int][]int, error) {
	result := make(map[int][]int)
	output, err := run("netstat", "-ano")
	if err != nil {
		return nil, fmt.Errorf("检测监听端口失败: %w", err)
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
	return result, nil
}

func getPortPids(run func(string, ...string) ([]byte, error)) (map[int]int, error) {
	result := make(map[int]int)
	output, err := run("netstat", "-ano")
	if err != nil {
		return nil, fmt.Errorf("检测监听端口失败: %w", err)
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
	return result, nil
}
