package control

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const maxRestoreBackoff = 5 * time.Second

func restorePollDelay(attempt int, base time.Duration) time.Duration {
	if base <= 0 {
		base = 200 * time.Millisecond
	}
	if attempt <= 1 {
		return base
	}
	d := base
	for i := 1; i < attempt; i++ {
		d *= 2
		if d >= maxRestoreBackoff {
			return maxRestoreBackoff
		}
	}
	return d
}

func readRestoredPID(dir string) int {
	path := filepath.Join(dir, "restored.pid")
	b, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(b)))
	if err != nil || pid <= 0 {
		return 0
	}
	return pid
}

func writeRestoreHandoff(dir string, pid int) error {
	body := strconv.Itoa(pid) + "\n"
	if err := os.WriteFile(filepath.Join(dir, "restored.handoff"), []byte(body), 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "restored.pid"), []byte(body), 0o644)
}

func shimSocketPath(runDir string, pid int) string {
	if runDir == "" {
		runDir = "/run/cudackpt"
	}
	return filepath.Join(runDir, strconv.Itoa(pid)+".sock")
}

func shimSocketReady(runDir string, pid int) bool {
	if pid <= 0 {
		return false
	}
	st, err := os.Stat(shimSocketPath(runDir, pid))
	return err == nil && st.Mode()&os.ModeSocket != 0
}

func resolveRestorePID(imagePath string, criuPID int, shims []int) int {
	if criuPID > 0 {
		return criuPID
	}
	if pid := readRestoredPID(imagePath); pid > 0 {
		return pid
	}
	if len(shims) > 0 {
		return shims[0]
	}
	return 0
}

func restoreCandidates(primary int, shims []int) []int {
	var out []int
	seen := make(map[int]struct{})
	if primary > 0 {
		out = append(out, primary)
		seen[primary] = struct{}{}
	}
	for _, p := range shims {
		if _, ok := seen[p]; ok {
			continue
		}
		out = append(out, p)
		seen[p] = struct{}{}
	}
	return out
}
