package criu

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type Wrapper struct {
	Bin string
}

func New() *Wrapper {
	b := "criu"
	if p := os.Getenv("CRIU_BIN"); p != "" {
		b = p
	}
	return &Wrapper{Bin: b}
}

func nvidiaExtMounts() []string {
	candidates := []string{
		"/dev/nvidiactl",
		"/dev/nvidia-uvm",
		"/dev/nvidia-uvm-tools",
		"/dev/nvidia0",
		"/dev/nvidia1",
		"/dev/nvidia2",
		"/dev/nvidia3",
	}
	var args []string
	for _, d := range candidates {
		if _, err := os.Stat(d); err == nil {
			args = append(args, "--ext-mount", d+":"+d)
		}
	}
	return args
}

func (w *Wrapper) Dump(pid int, dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	img := filepath.Join(dir, "criu")
	args := []string{"dump", "-t", fmt.Sprintf("%d", pid), "-D", img,
		"--shell-job", "--tcp-established", "--file-locks", "--link-remap"}
	args = append(args, nvidiaExtMounts()...)
	cmd := exec.Command(w.Bin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

var pidRe = regexp.MustCompile(`(?i)(?:Restored|restored).*pid[:\s]+(\d+)`)

func (w *Wrapper) Restore(dir string, logFile string, env []string) (int, error) {
	img := filepath.Join(dir, "criu")
	pidFile := filepath.Join(dir, "restored.pid")
	args := []string{"restore", "-D", img, "--shell-job", "--pidfile", pidFile,
		"--env", "CUDACKPT_NEED_GPU=1", "--link-remap"}
	args = append(args, nvidiaExtMounts()...)
	for _, e := range env {
		if e == "" {
			continue
		}
		args = append(args, "--env", e)
	}
	cmd := exec.Command(w.Bin, args...)
	var out []byte
	var err error
	if logFile != "" {
		out, err = cmd.CombinedOutput()
		_ = os.WriteFile(logFile, out, 0o644)
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
	}
	if err != nil {
		return 0, err
	}
	if b, rerr := os.ReadFile(pidFile); rerr == nil {
		s := strings.TrimSpace(string(b))
		if pid, perr := strconv.Atoi(s); perr == nil && pid > 0 {
			return pid, nil
		}
	}
	if len(out) == 0 && logFile != "" {
		out, _ = os.ReadFile(logFile)
	}
	if m := pidRe.FindSubmatch(out); len(m) == 2 {
		pid, _ := strconv.Atoi(string(m[1]))
		if pid > 0 {
			return pid, nil
		}
	}
	return 0, nil
}
