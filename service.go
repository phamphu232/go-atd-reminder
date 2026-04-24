package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
)

func controlService(action string) {
	exePath, _ := os.Executable()

	err := runAsAdmin(exePath, action)
	if err != nil {
		log.Printf("Error: %v", err)
	}
}

func runAsAdmin(exePath string, args string) error {
	var cmd string

	switch args {
	case "run":
		cmd = fmt.Sprintf("%q install && %q start", exePath, exePath)
	case "autostart":
		cmd = fmt.Sprintf("%q stop && %q uninstall && %q install && %q start", exePath, exePath, exePath, exePath)
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
