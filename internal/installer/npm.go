package installer

import (
	"os/exec"
	"strings"

	ylog "github.com/yangduck/yduck/internal/log"
)

type NpmInstaller struct{}

func NewNpmInstaller() *NpmInstaller {
	return &NpmInstaller{}
}

func (n *NpmInstaller) IsAvailable() bool {
	_, err := exec.LookPath("npm")
	return err == nil
}

func (n *NpmInstaller) IsInstalled(pkg string) (bool, string) {
	bin := pkg
	if i := strings.LastIndex(pkg, "/"); i >= 0 {
		bin = pkg[i+1:]
	}
	path, err := exec.LookPath(bin)
	if err != nil {
		return false, ""
	}
	ylog.S.Debugw("npm package found", "package", pkg, "path", path)
	out, err := exec.Command(bin, "--version").Output()
	if err != nil {
		return true, ""
	}
	return true, strings.TrimSpace(string(out))
}

func (n *NpmInstaller) Install(pkg string) error {
	ylog.S.Debugw("npm install -g", "package", pkg)
	cmd := exec.Command("npm", "install", "-g", pkg)
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		ylog.S.Errorw("npm install failed", "package", pkg, "error", err)
		return err
	}
	ylog.S.Infow("npm install succeeded", "package", pkg)
	return nil
}
