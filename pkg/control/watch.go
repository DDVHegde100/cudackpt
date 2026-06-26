package control

import (
	"fmt"
	"time"

	"github.com/dhruvhegde/cudackpt/internal/ckpterr"
)

type WatchOptions struct {
	UntilRunning bool
	Timeout      time.Duration
}

func WatchShim(o *Orchestrator, pid int, interval time.Duration, stop <-chan struct{}) error {
	return WatchShimWith(o, pid, interval, WatchOptions{}, stop)
}

func WatchShimWith(o *Orchestrator, pid int, interval time.Duration, opts WatchOptions, stop <-chan struct{}) error {
	if interval <= 0 {
		interval = 500 * time.Millisecond
	}
	last := uint32(999)
	deadline := time.Time{}
	if opts.Timeout > 0 {
		deadline = time.Now().Add(opts.Timeout)
	}
	for {
		select {
		case <-stop:
			return nil
		default:
		}
		if !deadline.IsZero() && time.Now().After(deadline) {
			return ckpterr.E(ckpterr.RPC, "watch timeout")
		}
		st, err := o.Status(pid)
		if err != nil {
			return ckpterr.Wrap(ckpterr.RPC, "watch", err)
		}
		if st != last {
			fmt.Printf("%s ts=%s\n", StateName(st), time.Now().UTC().Format(time.RFC3339))
			last = st
		}
		if opts.UntilRunning && (st == 4 || st == 3) {
			return nil
		}
		time.Sleep(interval)
	}
}
