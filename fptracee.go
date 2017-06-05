//go:generate make install clean
package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kardianos/osext"
)

func lookBesideExecutable(name string) (string, error) {
	if strings.Contains(name, "/") {
		return "", fmt.Errorf("path not relative to executable: %s", name)
	}
	dir, err := osext.ExecutableFolder()
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, name)
	return exec.LookPath(path)
}
