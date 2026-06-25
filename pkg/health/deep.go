package health

import (
	"os"
	"os/exec"
	"strings"
)

func DeepProbe() Status {
	var checks []Check
	checks = append(checks, probeOS())
	checks = append(checks, probeGPU())
	checks = append(checks, probeCRIU())
	checks = append(checks, probeRunDir())
	checks = append(checks, probeNvidiaDriver())
	checks = append(checks, probeCRIUFeatures())
	checks = append(checks, probeCaps())
	ok := true
	for _, c := range checks {
		if !c.OK {
			ok = false
		}
	}
	return Status{OK: ok, Checks: checks}
}

func probeNvidiaDriver() Check {
	if out, err := exec.Command("nvidia-smi", "--query-gpu=driver_version,name", "--format=csv,noheader").CombinedOutput(); err == nil {
		return Check{Name: "nvidia_smi", OK: true, Detail: strings.TrimSpace(string(out))}
	}
	if b, err := os.ReadFile("/proc/driver/nvidia/version"); err == nil {
		line := strings.SplitN(string(b), "\n", 2)[0]
		return Check{Name: "nvidia_drv", OK: true, Detail: strings.TrimSpace(line)}
	}
	return Check{Name: "nvidia_drv", OK: false, Detail: "driver version unavailable"}
}

func probeCRIUFeatures() Check {
	out, err := exec.Command("criu", "check", "--feature", "mem_track").CombinedOutput()
	if err != nil {
		return Check{Name: "criu_mem", OK: false, Detail: strings.TrimSpace(string(out))}
	}
	return Check{Name: "criu_mem", OK: true, Detail: "mem_track available"}
}
