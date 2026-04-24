//go:build darwin

package main

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

func IsScreenLocked() bool {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	cmd := strings.Split("ioreg -n Root -d1", " ")

	// This will output registry with session information
	output, _ := exec.CommandContext(ctx, cmd[0], cmd[1:]...).Output()

	// This will check if output["IOConsoleUsers"][0]["CGSSessionScreenIsLocked"] exists
	return strings.Contains(string(output), "CGSSessionScreenIsLocked")
}
