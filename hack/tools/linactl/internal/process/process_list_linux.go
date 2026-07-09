//go:build linux

package process

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// List returns visible processes on Linux using /proc.
func List() ([]Info, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}
	infos := make([]Info, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(entry.Name())
		if err != nil || pid <= 1 {
			continue
		}
		args := readLinuxArgs(pid)
		if len(args) == 0 {
			continue
		}
		cwd, _ := os.Readlink(filepath.Join("/proc", entry.Name(), "cwd"))
		infos = append(infos, Info{PID: pid, Args: args, CWD: cwd})
	}
	return infos, nil
}

func readLinuxArgs(pid int) []string {
	data, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "cmdline"))
	if err != nil || len(data) == 0 {
		return nil
	}
	parts := strings.Split(strings.TrimRight(string(data), "\x00"), "\x00")
	args := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			args = append(args, part)
		}
	}
	return args
}
