//go:build darwin

package process

import (
	"encoding/binary"

	"golang.org/x/sys/unix"
)

// List returns visible processes on macOS using sysctl so linactl does not
// depend on platform shell commands such as ps or lsof.
func List() ([]Info, error) {
	procs, err := unix.SysctlKinfoProcSlice("kern.proc.all")
	if err != nil {
		return nil, err
	}
	infos := make([]Info, 0, len(procs))
	for _, proc := range procs {
		pid := int(proc.Proc.P_pid)
		if pid <= 1 {
			continue
		}
		args, err := darwinArgs(pid)
		if err != nil || len(args) == 0 {
			continue
		}
		infos = append(infos, Info{PID: pid, Args: args})
	}
	return infos, nil
}

func darwinArgs(pid int) ([]string, error) {
	raw, err := unix.SysctlRaw("kern.procargs2", pid)
	if err != nil {
		return nil, err
	}
	if len(raw) < 4 {
		return nil, nil
	}
	argc := int(binary.LittleEndian.Uint32(raw[:4]))
	if argc <= 0 {
		return nil, nil
	}
	cursor := raw[4:]
	cursor = trimUntilNUL(cursor)
	cursor = trimLeadingNUL(cursor)

	args := make([]string, 0, argc)
	for len(cursor) > 0 && len(args) < argc {
		item, rest := cutNUL(cursor)
		if len(item) > 0 {
			args = append(args, string(item))
		}
		cursor = trimLeadingNUL(rest)
	}
	return args, nil
}

func trimUntilNUL(data []byte) []byte {
	for len(data) > 0 && data[0] != 0 {
		data = data[1:]
	}
	return data
}

func trimLeadingNUL(data []byte) []byte {
	for len(data) > 0 && data[0] == 0 {
		data = data[1:]
	}
	return data
}

func cutNUL(data []byte) ([]byte, []byte) {
	for i, value := range data {
		if value == 0 {
			return data[:i], data[i+1:]
		}
	}
	return data, nil
}
