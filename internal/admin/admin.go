package admin

import (
	"os"
	"syscall"
	"unsafe"
)

// IsAdmin 检查当前进程是否以管理员权限运行
func IsAdmin() bool {
	// 通过尝试创建文件到系统目录来简单判断
	// 更可靠的方式：使用 Windows API
	return checkAdminViaAPI()
}

func checkAdminViaAPI() bool {
	shell32 := syscall.NewLazyDLL("shell32.dll")
	procIsUserAnAdmin := shell32.NewProc("IsUserAnAdmin")

	ret, _, _ := procIsUserAnAdmin.Call()
	return ret != 0
}

// GetCurrentExePath 获取当前可执行文件路径
func GetCurrentExePath() (string, error) {
	return os.Executable()
}

// StringToUTF16Ptr 字符串转 UTF16 指针（用于 Windows API）
func StringToUTF16Ptr(s string) *uint16 {
	p, _ := syscall.UTF16PtrFromString(s)
	return p
}

// RunAsAdmin 以管理员权限重启程序
func RunAsAdmin() error {
	shell32 := syscall.NewLazyDLL("shell32.dll")
	procShellExecute := shell32.NewProc("ShellExecuteW")

	exePath, err := GetCurrentExePath()
	if err != nil {
		return err
	}

	verb, _ := syscall.UTF16PtrFromString("runas")
	file, _ := syscall.UTF16PtrFromString(exePath)

	ret, _, _ := procShellExecute.Call(
		0,
		uintptr(unsafe.Pointer(verb)),
		uintptr(unsafe.Pointer(file)),
		0, 0,
		1, // SW_SHOWNORMAL
	)

	if ret <= 32 {
		return syscall.Errno(ret)
	}

	// 提权成功后退出当前进程
	os.Exit(0)
	return nil
}
