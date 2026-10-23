package platform

import (
	"fmt"
	"os"
	"os/exec"
)

func Asset(goos, goarch string) (string, error) {
	name := map[string]string{"windows/amd64": "stimulsoft-runner-win-x64.exe", "linux/amd64": "stimulsoft-runner-linux-x64", "darwin/arm64": "stimulsoft-runner-osx-arm64", "darwin/amd64": "stimulsoft-runner-osx-x64"}[goos+"/"+goarch]
	if name == "" {
		return "", fmt.Errorf("unsupported runner platform %s/%s", goos, goarch)
	}
	return name, nil
}

func MakeExecutable(path string, windows bool) error {
	if windows {
		return nil
	}
	return os.Chmod(path, 0755)
}

func RemoveQuarantine(path string, darwin bool) error {
	if !darwin {
		return nil
	}
	return exec.Command("xattr", "-d", "com.apple.quarantine", path).Run()
}
