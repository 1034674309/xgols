// Command xgols-zed starts the bundled xgols with its matching XGo runtime.
// It is intended for editor integrations that extract a release archive into
// a private cache instead of installing XGo globally.
package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	directory, err := executableDirectory()
	if err != nil {
		fatal(err)
	}

	suffix := ""
	if runtime.GOOS == "windows" {
		suffix = ".exe"
	}
	xgolsPath := filepath.Join(directory, "xgols"+suffix)
	xgoPath := filepath.Join(directory, "xgo"+suffix)
	xgoRoot := filepath.Join(directory, "xgo-runtime")
	for label, path := range map[string]string{
		"xgols executable": xgolsPath,
		"xgo executable":   xgoPath,
		"XGo runtime":      xgoRoot,
	} {
		if info, err := os.Stat(path); err != nil || (label == "XGo runtime" && !info.IsDir()) {
			fatal(fmt.Errorf("bundled %s is missing at %s", label, path))
		}
	}

	command := exec.Command(xgolsPath, os.Args[1:]...)
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Env = bundledEnvironment(directory, xgoRoot)
	if err := command.Run(); err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			os.Exit(exitError.ExitCode())
		}
		fatal(fmt.Errorf("start bundled xgols: %w", err))
	}
}

func executableDirectory() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("locate xgols-zed: %w", err)
	}
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		path = resolved
	}
	return filepath.Dir(path), nil
}

func bundledEnvironment(directory, xgoRoot string) []string {
	values := make([]string, 0, len(os.Environ())+2)
	for _, value := range os.Environ() {
		if strings.HasPrefix(value, "PATH=") || strings.HasPrefix(value, "XGOROOT=") {
			continue
		}
		values = append(values, value)
	}
	path := directory
	if inherited := os.Getenv("PATH"); inherited != "" {
		path += string(os.PathListSeparator) + inherited
	}
	return append(values, "PATH="+path, "XGOROOT="+xgoRoot)
}

func fatal(err error) {
	_, _ = fmt.Fprintln(os.Stderr, "xgols-zed:", err)
	os.Exit(1)
}
