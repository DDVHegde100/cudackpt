//go:build !linux

package health

import "runtime"

func probeCaps() Check {
	return Check{Name: "caps", OK: true, Detail: "skipped on " + runtime.GOOS}
}
