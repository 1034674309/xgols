// Command prepare-xgo-runtime copies the XGo version used by xgols into the
// release staging directory. GoReleaser then builds the matching xgo command
// and archives both it and this source tree with xgols.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	xgoModule  = "github.com/goplus/xgo"
	xgoVersion = "v1.6.2"
	runtimeDir = ".runtime/xgo"
)

type downloadResult struct {
	Dir   string
	Error string
}

func main() {
	source, err := moduleDirectory()
	if err != nil {
		fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(runtimeDir), 0o755); err != nil {
		fatal(fmt.Errorf("create runtime staging directory: %w", err))
	}
	if err := os.Chmod(filepath.Dir(runtimeDir), 0o755); err != nil {
		fatal(fmt.Errorf("make runtime staging directory writable: %w", err))
	}
	if err := os.RemoveAll(runtimeDir); err != nil {
		fatal(fmt.Errorf("remove previous XGo runtime: %w", err))
	}
	if err := copyTree(source, runtimeDir); err != nil {
		fatal(fmt.Errorf("copy XGo runtime: %w", err))
	}
	if err := os.WriteFile(filepath.Join(runtimeDir, "VERSION"), []byte(xgoVersion+"\n"), 0o644); err != nil {
		fatal(fmt.Errorf("write XGo runtime version: %w", err))
	}
}

func moduleDirectory() (string, error) {
	command := exec.Command("go", "mod", "download", "-json", xgoModule+"@"+xgoVersion)
	output, err := command.Output()
	if err != nil {
		return "", fmt.Errorf("download %s: %w", xgoModule+"@"+xgoVersion, err)
	}
	var result downloadResult
	if err := json.Unmarshal(output, &result); err != nil {
		return "", fmt.Errorf("decode Go module metadata: %w", err)
	}
	if result.Error != "" {
		return "", fmt.Errorf("download %s: %s", xgoModule+"@"+xgoVersion, result.Error)
	}
	if result.Dir == "" {
		return "", fmt.Errorf("download %s: module directory was empty", xgoModule+"@"+xgoVersion)
	}
	return result.Dir, nil
}

func copyTree(source, destination string) error {
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, relative)
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if entry.IsDir() {
			// Go's module cache is read-only. The release staging tree must be
			// writable while the rest of the copy is still being populated.
			return os.MkdirAll(target, 0o755)
		}
		if entry.Type()&os.ModeSymlink != 0 {
			link, err := os.Readlink(path)
			if err != nil {
				return err
			}
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			return os.Symlink(link, target)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		input, err := os.Open(path)
		if err != nil {
			return err
		}
		defer input.Close()
		output, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(output, input)
		closeErr := output.Close()
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	})
}

func fatal(err error) {
	_, _ = fmt.Fprintln(os.Stderr, "prepare-xgo-runtime:", err)
	os.Exit(1)
}
