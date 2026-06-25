package control

import (
	"time"

	"github.com/dhruvhegde/cudackpt/internal/ckpterr"
	jlog "github.com/dhruvhegde/cudackpt/pkg/log"
)

type RetryPolicy struct {
	MaxAttempts int
	Backoff     time.Duration
}

func (o *Orchestrator) retryPolicy() RetryPolicy {
	p := RetryPolicy{MaxAttempts: o.cfg.MaxRetries, Backoff: o.cfg.RetryBackoff}
	if p.MaxAttempts <= 0 {
		p.MaxAttempts = 1
	}
	if p.Backoff <= 0 {
		p.Backoff = 500 * time.Millisecond
	}
	return p
}

func (o *Orchestrator) CheckpointWithRetry(pid int, out string, policy RetryPolicy) error {
	if policy.MaxAttempts <= 0 {
		policy = o.retryPolicy()
	}
	var last error
	for attempt := 1; attempt <= policy.MaxAttempts; attempt++ {
		jlog.Info("checkpoint_attempt", map[string]any{"pid": pid, "attempt": attempt})
		err := o.Checkpoint(pid, out)
		if err == nil {
			return nil
		}
		last = err
		if attempt == policy.MaxAttempts {
			break
		}
		if ce, ok := err.(*ckpterr.Error); ok && ce.Code == ckpterr.Unsupported {
			break
		}
		time.Sleep(policy.Backoff)
	}
	return last
}

func (o *Orchestrator) EnqueueCheckpoint(pid int, out string) error {
	return o.CheckpointWithRetry(pid, out, o.retryPolicy())
}
