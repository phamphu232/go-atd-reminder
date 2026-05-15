//go:build linux

package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func IsScreenLocked(username string) bool {
	uidByte, err := exec.Command("id", "-u", username).Output()
	if err != nil {
		return false
	}
	uidStr := strings.TrimSpace(string(uidByte))
	uid, _ := strconv.Atoi(uidStr)

	de := detectDE(username)
	dbusAddr := fmt.Sprintf("unix:path=/run/user/%d/bus", uid)

	var cmdStr string
	var contains string

	switch de {
	case "KDE":
		cmdStr = "qdbus org.freedesktop.ScreenSaver /ScreenSaver org.freedesktop.ScreenSaver.GetActive"
		contains = "true"
	case "XFCE":
		cmdStr = "xfconf-query -c xfce4-session -p /general/LockDialogIsVisible"
		contains = "true"
	default: // GNOME
		cmdStr = "gdbus call --session --dest org.gnome.ScreenSaver --object-path /org/gnome/ScreenSaver --method org.gnome.ScreenSaver.GetActive"
		contains = "true"
	}

	return checkAsUser(uid, dbusAddr, cmdStr, contains)
}

func checkAsUser(uid int, dbusAddr, cmdStr, contains string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	parts := strings.Split(cmdStr, " ")
	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{Uid: uint32(uid), Gid: uint32(uid)},
	}

	cmd.Env = append(os.Environ(), "DBUS_SESSION_BUS_ADDRESS="+dbusAddr)

	output, _ := cmd.Output()
	return strings.Contains(string(output), contains)
}

func detectDE(username string) string {
	out, _ := exec.Command("ps", "-u", username).Output()
	psOut := string(out)
	if strings.Contains(psOut, "gnome-shell") {
		return "GNOME"
	}
	if strings.Contains(psOut, "ksmserver") {
		return "KDE"
	}
	return "GNOME"
}
