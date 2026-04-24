//go:build linux

package main

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"time"
)

func IsScreenLocked() bool {
	// WARNING: these are not tested
	switch os.Getenv("XDG_CURRENT_DESKTOP") {
	case "Unity":
		return check(
			"gdbus call -e -d com.canonical.Unity -o /com/canonical/Unity/Session -m com.canonical.Unity.Session.IsLocked",
			"true",
		)
	case "KDE":
		return check(
			"qdbus org.kde.screensaver /ScreenSaver org.freedesktop.ScreenSaver.GetActive",
			"true",
		)
	case "XFCE":
		return check(
			"xfconf-query -c xfce4-session -p /general/LockDialogIsVisible",
			"true",
		)
	case "LXQt":
	case "MATE":
		return check(
			"mate-screensaver-command -q",
			"is active",
		)
	case "Cinnamon":
		return check(
			"gdbus call --session --dest org.Cinnamon.ScreenSaver --object-path /org/Cinnamon/ScreenSaver --method org.Cinnamon.ScreenSaver.GetActive",
			"true",
		)
	case "LXDE":
	case "Deepin":
		return check(
			"dbus-send --session --dest=com.deepin.ScreenSaver --type=method_call --print-reply /com/deepin/ScreenSaver com.deepin.ScreenSaver.GetStatus",
			"true",
		)
	default: // default to gnome
		return check(
			"gdbus call --session --dest org.gnome.ScreenSaver --object-path /org/gnome/ScreenSaver --method org.gnome.ScreenSaver.GetActive",
			"true",
		)
	}

	return false
}

func check(cmdStr, contains string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	cmd := strings.Split(cmdStr, " ")

	// This will output a list of tasks currently running
	output, _ := exec.CommandContext(ctx, cmd[0], cmd[1:]...).Output()

	// This will check if LogonUI.exe is in the list
	return strings.Contains(string(output), contains)
}
