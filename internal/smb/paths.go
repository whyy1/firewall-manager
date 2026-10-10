package smb

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

// Only normal drive-qualified paths may reach filesystem inspection.
func validateLocalDirectoryPath(path string) error {
	if len(path) < 3 || !isDriveLetter(path[0]) || path[1] != ':' || (path[2] != '\\' && path[2] != '/') || !filepath.IsAbs(path) {
		return errors.New("目录必须是本机盘符绝对路径，不支持 UNC 或设备路径")
	}
	if strings.ContainsAny(path[2:], `:"<>|?*`) || hasControlCharacter(path) {
		return errors.New("本机目录包含无效字符")
	}
	return nil
}

func isDriveLetter(char byte) bool {
	return char >= 'A' && char <= 'Z' || char >= 'a' && char <= 'z'
}

func hasControlCharacter(value string) bool {
	return strings.IndexFunc(value, func(char rune) bool { return char < 32 }) >= 0
}

func localDirectoryInfo(path string) (os.FileInfo, error) {
	root, err := windows.UTF16PtrFromString(path[:2] + `\`)
	if err != nil {
		return nil, err
	}
	if windows.GetDriveType(root) == windows.DRIVE_REMOTE {
		return nil, errors.New("本机目录不能位于映射网络驱动器")
	}
	return os.Stat(path)
}

func ordinaryUNCParts(path string) ([]string, bool) {
	if !strings.HasPrefix(path, `\\`) || strings.ContainsAny(path, `/"<>|?*:`) || hasControlCharacter(path) {
		return nil, false
	}
	parts := strings.Split(strings.TrimSuffix(path[2:], `\`), `\`)
	if len(parts) < 2 {
		return nil, false
	}
	for _, part := range parts {
		if part == "" || part == "." || part == ".." || strings.TrimSpace(part) != part || strings.HasSuffix(part, ".") {
			return nil, false
		}
	}
	return parts, true
}

func validateOpenLocation(location string) error {
	if location == "" {
		return errors.New("打开位置不能为空")
	}
	if strings.HasPrefix(location, `\\`) {
		if _, ok := ordinaryUNCParts(location); !ok {
			return errors.New("网络共享地址格式无效")
		}
		return nil
	}
	if err := validateLocalDirectoryPath(location); err != nil {
		return err
	}
	info, err := localDirectoryInfo(location)
	if err != nil {
		return fmt.Errorf("本机目录不可用: %w", err)
	}
	if !info.IsDir() {
		return errors.New("打开位置必须是目录")
	}
	return nil
}
