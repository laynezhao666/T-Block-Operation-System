package os

import (
	"os/exec"

	"trpc.group/trpc-go/trpc-go/log"
)

const appRestartShell = "/opt/tbbox/shells/tbbox_restart.sh 2>&1 >/dev/null &"

// Reboot 重启
func Reboot() error {
	log.Warn("Call Reboot")
	cmd := exec.Command("reboot")
	return cmd.Run()
}

// ReStartApp 重启应用
func ReStartApp() {
	log.Warn("Call ReStart shell")
	cmd := exec.Command("sh", "-c", appRestartShell)
	_ = cmd.Run()
}
