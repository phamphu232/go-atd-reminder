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
	isWorking := isActive && !IsScreenLocked(user)

	// log.Printf("UserIsWorking: %v, IsLocked: %v", isWorking, IsScreenLocked(user))

	return isWorking
}
