package system

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func IsAdmin() bool {
	if runtime.GOOS == "windows" {
		return isWindowsAdmin()
	}
	return os.Geteuid() == 0
}

// isWindowsAdmin checks for admin privileges on Windows using PowerShell.
// PowerShell's IsInRole check is the most reliable method across VMs, cloud, and all Windows editions.
func isWindowsAdmin() bool {
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command",
		"([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)")

	output, err := cmd.CombinedOutput()
	if err != nil {
		// If PowerShell fails, try fallback methods
		return isWindowsAdminFallback()
	}

	// PowerShell returns "True" or "False"
	result := strings.TrimSpace(string(output))
	return strings.ToLower(result) == "true"
}

// isWindowsAdminFallback provides fallback methods for Windows admin detection
func isWindowsAdminFallback() bool {
	systemDrive := os.Getenv("SystemDrive")
	if systemDrive == "" {
		systemDrive = "C:"
	}

	f, err := os.Open(systemDrive + "\\Windows\\System32\\config") // requires admin
	if err == nil {
		f.Close()
		return true
	}

	cmd := exec.Command("net", "session")
	err = cmd.Run()
	if err == nil {
		return true
	}

	cmd = exec.Command("cmd", "/c", "fsutil", "dirty", "query", systemDrive)
	_, err = cmd.CombinedOutput()
	if err == nil {
		return true
	}

	return false
}
