package control

import (
	"fmt"
	"time"

	"github.com/dhruvhegde/cudackpt/internal/ckpterr"
)

func WatchShim(o *Orchestrator, pid int, interval time.Duration, stop <-chan struct{}) error {
	if interval <= 0 {
		interval = 500 * time.Millisecond
	}
	last := uint32(999)
	for {
		select {
		case <-stop:
			return nil
		default:
		}
		st, err := o.Status(pid)
		if err != nil {
			return ckpterr.Wrap(ckpterr.RPC, "watch", err)
		}
		if st != last {
			fmt.Printf("%s ts=%s\n", StateName(st), time.Now().UTC().Format(time.RFC3339))
			last = st
		}
		time.Sleep(interval)
	}
}
