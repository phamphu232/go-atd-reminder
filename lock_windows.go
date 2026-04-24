//go:build windows

package main

import (
	"context"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

func IsScreenLocked() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// tasklist /FI "IMAGENAME eq logonui.exe" /FO CSV /NH
	cmd := exec.CommandContext(ctx, "tasklist", "/FI", "IMAGENAME eq logonui.exe", "/FO", "CSV", "/NH")

	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000,
	}

	output, _ := cmd.Output()
	// output, err := cmd.Output()
	// log.Printf("Locked state: %s, Error: %v", string(output), err)

	return strings.Contains(strings.ToLower(string(output)), "logonui.exe")
}
