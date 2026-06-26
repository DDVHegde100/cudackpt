package control

import (
	"errors"
	"testing"
	"time"

	"github.com/dhruvhegde/cudackpt/internal/ckpterr"
)

func TestApplyRetrySucceedsAfterFailures(t *testing.T) {
	attempts := 0
	err := applyRetry(RetryPolicy{MaxAttempts: 3, Backoff: time.Millisecond}, func(int) error {
		attempts++
		if attempts < 3 {
			return errors.New("transient")
		}
		return nil
	}, func(error) bool { return false })
	if err != nil {
		t.Fatal(err)
	}
	if attempts != 3 {
		t.Fatalf("attempts=%d", attempts)
	}
}

func TestApplyRetryStopsOnUnsupported(t *testing.T) {
	attempts := 0
	err := applyRetry(RetryPolicy{MaxAttempts: 5, Backoff: time.Millisecond}, func(int) error {
		attempts++
		return ckpterr.E(ckpterr.Unsupported, "multi-gpu")
	}, func(err error) bool {
		if ce, ok := err.(*ckpterr.Error); ok && ce.Code == ckpterr.Unsupported {
			return true
		}
		return false
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if attempts != 1 {
		t.Fatalf("attempts=%d want 1", attempts)
	}
}

func TestApplyRetryExhaustsAttempts(t *testing.T) {
	attempts := 0
	err := applyRetry(RetryPolicy{MaxAttempts: 2, Backoff: time.Millisecond}, func(int) error {
		attempts++
		return errors.New("fail")
	}, func(error) bool { return false })
	if err == nil {
		t.Fatal("expected error")
	}
	if attempts != 2 {
		t.Fatalf("attempts=%d", attempts)
	}
}
