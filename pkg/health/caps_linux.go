//go:build linux

package health

import (
	"os"
	"strconv"
	"strings"
)

const capSysAdmin = 1 << 21

func probeCaps() Check {
	b, err := os.ReadFile("/proc/self/status")
	if err != nil {
		return Check{Name: "caps", OK: false, Detail: err.Error()}
	}
	for _, line := range strings.Split(string(b), "\n") {
		if !strings.HasPrefix(line, "CapEff:") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			break
		}
		mask, err := parseCapMask(fields[1:])
		if err != nil {
			return Check{Name: "caps", OK: false, Detail: err.Error()}
		}
		if mask&capSysAdmin != 0 {
			return Check{Name: "caps", OK: true, Detail: "CAP_SYS_ADMIN"}
		}
		return Check{Name: "caps", OK: false, Detail: "missing CAP_SYS_ADMIN for criu"}
	}
	return Check{Name: "caps", OK: false, Detail: "CapEff missing"}
}

func parseCapMask(fields []string) (uint64, error) {
	var mask uint64
	for i, f := range fields {
		v, err := strconv.ParseUint(f, 16, 64)
		if err != nil {
			return 0, err
		}
		mask |= v << (32 * i)
	}
	return mask, nil
}
