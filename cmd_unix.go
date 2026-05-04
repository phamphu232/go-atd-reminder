//go:build !windows

package main

import (
	"fmt"
	"os/exec"
	"runtime"
)

func runAsAdmin(exePath string, args string) error {
	var cmd string

	switch args {
	case "reinstall":
		cmd = fmt.Sprintf("%q stop && %q uninstall && %q install", exePath, exePath, exePath)
	default:
		cmd = fmt.Sprintf("%q %s", exePath, args)
	}

	switch runtime.GOOS {
	case "darwin":
		appleScript := fmt.Sprintf("do shell script \"sh -c %q\" with administrator privileges", cmd)
		return exec.Command("osascript", "-e", appleScript).Run()

	case "linux":
		return exec.Command("pkexec", "sh", "-c", cmd).Run()

	default:
		return fmt.Errorf("error: OS %s not supported", runtime.GOOS)
	}
}
