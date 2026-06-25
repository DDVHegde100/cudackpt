package health

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

type Status struct {
	OK     bool
	Checks []Check
}

type Check struct {
	Name    string
	OK      bool
	Detail  string
}

func Probe() Status {
	var checks []Check
	checks = append(checks, probeOS())
	checks = append(checks, probeGPU())
	checks = append(checks, probeCRIU())
	checks = append(checks, probeRunDir())
	ok := true
	for _, c := range checks {
		if !c.OK {
			ok = false
		}
	}
	return Status{OK: ok, Checks: checks}
}

func probeOS() Check {
	if runtime.GOOS != "linux" {
		return Check{Name: "os", OK: false, Detail: runtime.GOOS}
	}
	return Check{Name: "os", OK: true, Detail: "linux"}
}

func probeGPU() Check {
	for _, d := range []string{"/dev/nvidiactl", "/dev/nvidia0"} {
		if _, err := os.Stat(d); err == nil {
			return Check{Name: "gpu", OK: true, Detail: d}
		}
	}
	return Check{Name: "gpu", OK: false, Detail: "no nvidia device nodes"}
}

func probeCRIU() Check {
	if _, err := exec.LookPath("criu"); err != nil {
		return Check{Name: "criu", OK: false, Detail: "not in PATH"}
	}
	out, err := exec.Command("criu", "check").CombinedOutput()
	if err != nil {
		return Check{Name: "criu", OK: false, Detail: string(out)}
	}
	return Check{Name: "criu", OK: true, Detail: "check passed"}
}

func probeRunDir() Check {
	dir := "/run/cudackpt"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return Check{Name: "run_dir", OK: false, Detail: err.Error()}
	}
	return Check{Name: "run_dir", OK: true, Detail: dir}
}

func Format(s Status) string {
	var b string
	if s.OK {
		b = "health ok\n"
	} else {
		b = "health degraded\n"
	}
	for _, c := range s.Checks {
		flag := "fail"
		if c.OK {
			flag = "ok"
		}
		b += fmt.Sprintf("  %-10s %-4s %s\n", c.Name, flag, c.Detail)
	}
	return b
}
