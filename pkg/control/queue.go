package control

import (
	"time"

	"github.com/dhruvhegde/cudackpt/internal/ckpterr"
	jlog "github.com/dhruvhegde/cudackpt/pkg/log"
	"github.com/dhruvhegde/cudackpt/pkg/metrics"
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

func applyRetry(policy RetryPolicy, attemptFn func(int) error, stopRetry func(error) bool) error {
	if policy.MaxAttempts <= 0 {
		policy.MaxAttempts = 1
	}
	if policy.Backoff <= 0 {
		policy.Backoff = 500 * time.Millisecond
	}
	var last error
	for attempt := 1; attempt <= policy.MaxAttempts; attempt++ {
		err := attemptFn(attempt)
		if err == nil {
			return nil
		}
		last = err
		if attempt == policy.MaxAttempts || stopRetry(err) {
			break
		}
		time.Sleep(policy.Backoff)
	}
	return last
}

func (o *Orchestrator) CheckpointWithRetry(pid int, out string, policy RetryPolicy) error {
	if policy.MaxAttempts <= 0 {
		policy = o.retryPolicy()
	}
	err := applyRetry(policy, func(attempt int) error {
		jlog.Info("checkpoint_attempt", map[string]any{"pid": pid, "attempt": attempt})
		return o.doCheckpoint(pid, out)
	}, func(err error) bool {
		if ce, ok := err.(*ckpterr.Error); ok && ce.Code == ckpterr.Unsupported {
			return true
		}
		return false
	})
	if err != nil {
		jlog.Error("checkpoint_fail", err, map[string]any{"pid": pid})
		metrics.Default.Inc(metrics.CheckpointFailures)
		return err
	}
	metrics.Default.Inc(metrics.CheckpointsTotal)
	jlog.Info("checkpoint_ok", map[string]any{"pid": pid, "dir": out})
	return nil
}

func (o *Orchestrator) EnqueueCheckpoint(pid int, out string) error {
	return o.CheckpointWithRetry(pid, out, o.retryPolicy())
}
