//go:build linux

package health

import "syscall"

func probeCaps() Check {
	hdr := syscall.CapUserHeader{Version: syscall.LINUX_CAPABILITY_VERSION_3}
	var data [2]syscall.CapUserData
	if err := syscall.Capget(&hdr, &data[0]); err != nil {
		return Check{Name: "caps", OK: false, Detail: err.Error()}
	}
	eff := uint64(data[0].Effective) | uint64(data[1].Effective)<<32
	if eff&(1<<syscall.CAP_SYS_ADMIN) != 0 {
		return Check{Name: "caps", OK: true, Detail: "CAP_SYS_ADMIN"}
	}
	return Check{Name: "caps", OK: false, Detail: "missing CAP_SYS_ADMIN for criu"}
}
