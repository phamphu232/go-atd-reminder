//go:build linux

package main

import (
	"bytes"
	"os/exec"
	"strings"
)

func getLoginctlProperty(user, property string) string {
	cmd := exec.Command("loginctl", "show-user", user, "--property="+property, "--value")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()

	if err != nil {
		return ""
	}

	return strings.TrimSpace(out.String())
}

// func ChangeSinceTime(user string) time.Time {
// 	idleSinceStr := getLoginctlProperty(user, "IdleSinceHint")
// 	if idleSinceStr == "" {
// 		return time.Now()
// 	}

// 	microtime, err := strconv.ParseInt(idleSinceStr, 10, 64)
// 	if err != nil {
// 		fmt.Println("Error parsing time:", err)
// 		return time.Now()
// 	}

// 	t := time.UnixMicro(microtime)

// 	return t
// }

func IsWorking(user string) bool {
	isActive := getLoginctlProperty(user, "State") == "active"
	isIdle := getLoginctlProperty(user, "IdleHint") != "no"

	isWorking := isActive && !isIdle && !IsScreenLocked()

	// log.Printf("UserIsWorking: %v, IsIdle: %v, IsLocked: %v", isWorking, isIdle, IsScreenLocked())

	return isWorking
}
