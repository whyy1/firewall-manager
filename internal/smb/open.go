package smb

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// OpenLocation requests a visible Explorer window for a validated directory.
func OpenLocation(location string) error {
	return openLocationWith(location, validateOpenLocation, startExplorer)
}

func openLocationWith(location string, validate func(string) error, start func(*exec.Cmd) error) error {
	location = strings.TrimSpace(location)
	if err := validate(location); err != nil {
		return err
	}
	cmd := exec.Command("explorer.exe", filepath.Clean(location))
	// Explorer is the requested user interface, so its window must remain visible.
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: false}
	if err := start(cmd); err != nil {
		return fmt.Errorf("请求资源管理器打开「%s」失败: %w", location, err)
	}
	return nil
}

func startExplorer(cmd *exec.Cmd) error {
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Process.Release()
}
