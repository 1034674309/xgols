// Command prepare-release-metadata records the XGo module version selected by
// xgols. Release consumers can then install the matching official XGo archive
// instead of copying the complete runtime into every xgols archive.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	xgoModule   = "github.com/goplus/xgo"
	metadataDir = ".release"
	versionFile = metadataDir + "/xgo-version"
)

func main() {
	version, err := selectedModuleVersion(xgoModule)
	if err != nil {
		fatal(err)
	}
	if err := os.MkdirAll(metadataDir, 0o755); err != nil {
		fatal(fmt.Errorf("create release metadata directory: %w", err))
	}
	if err := os.WriteFile(filepath.Clean(versionFile), []byte(version+"\n"), 0o644); err != nil {
		fatal(fmt.Errorf("write XGo version metadata: %w", err))
	}
}

func selectedModuleVersion(module string) (string, error) {
	command := exec.Command("go", "list", "-m", "-f", "{{.Version}}", module)
	output, err := command.Output()
	if err != nil {
		return "", fmt.Errorf("resolve %s version: %w", module, err)
	}
	version := strings.TrimSpace(string(output))
	if !strings.HasPrefix(version, "v") || len(version) == 1 {
		return "", fmt.Errorf("resolve %s version: got %q", module, version)
	}
	return version, nil
}

func fatal(err error) {
	_, _ = fmt.Fprintln(os.Stderr, "prepare-release-metadata:", err)
	os.Exit(1)
}
