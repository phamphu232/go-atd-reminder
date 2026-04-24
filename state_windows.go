//go:build windows

package main

import (
	"context"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

func getUserState(user string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// tasklist /FI "USERNAME eq phamphu232" /FI "IMAGENAME eq explorer.exe" /FO CSV /NH
	cmd := exec.CommandContext(ctx, "tasklist", "/FI", "USERNAME eq "+user, "/FI", "IMAGENAME eq explorer.exe", "/FO", "CSV", "/NH")

	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000,
	}

	output, _ := cmd.Output()
	// output, err := cmd.Output()
	// log.Printf("User state: %s, Error: %v", string(output), err)

	return strings.Contains(strings.ToLower(string(output)), "explorer.exe")
}

func IsWorking(user string) bool {
	isActive := getUserState(user)
	isWorking := isActive && !IsScreenLocked()

	// log.Printf("UserIsWorking: %v, IsLocked: %v", isWorking, IsScreenLocked())

	return isWorking
}
