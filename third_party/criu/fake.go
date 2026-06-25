package criu

import (
	"os"
	"path/filepath"
	"strconv"
)

type Fake struct {
	RestorePID int
	DumpErr    error
	RestoreErr error
}

func (f *Fake) Dump(pid int, dir string) error {
	if f.DumpErr != nil {
		return f.DumpErr
	}
	return os.MkdirAll(filepath.Join(dir, "criu"), 0o755)
}

func (f *Fake) Restore(dir string, logFile string, env []string) (int, error) {
	_ = env
	if f.RestoreErr != nil {
		return 0, f.RestoreErr
	}
	pid := f.RestorePID
	if pid <= 0 {
		pid = 4242
	}
	body := []byte(strconv.Itoa(pid) + "\n")
	_ = os.WriteFile(filepath.Join(dir, "restored.pid"), body, 0o644)
	if logFile != "" {
		_ = os.WriteFile(logFile, []byte("fake criu restore pid="+strconv.Itoa(pid)+"\n"), 0o644)
	}
	return pid, nil
}
