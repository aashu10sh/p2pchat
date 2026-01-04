package utils

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

func GetCurrentSSID() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return getDarwinSSID()
	case "linux":
		return getLinuxSSID()
	case "windows":
		return getWindowsSSID()

	default:
		return "", fmt.Errorf("unsupported platform %s ", runtime.GOOS)
	}
}

func getDarwinSSID() (string, error) {
	return "", errors.New("not implemented")
}

func getLinuxSSID() (string, error) {
	cmd := exec.Command("iwgetid", "-r")
	output, err := cmd.Output()

	if err == nil {
		ssid := strings.TrimSpace(string(output))

		if ssid != "" {
			return ssid, nil
		}

	}

	cmd = exec.Command("nmcli", "-t", "-f", "active,ssid", "dev", "wifi")
	output, err = cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "yes:") {
				return strings.TrimPrefix(line, "yes:"), nil
			}
		}
	}

	cmd = exec.Command("sh", "-c", "iw dev | grep ssid | awk '{print $2}'")
	output, err = cmd.Output()
	if err == nil {
		ssid := strings.TrimSpace(string(output))
		if ssid != "" {
			return ssid, nil
		}
	}

	return "", fmt.Errorf("unable to get SSID on Linux")
}

// todo(aashutosh): impl
func getWindowsSSID() (string, error) {
	return "", errors.New("not implemented")
}
