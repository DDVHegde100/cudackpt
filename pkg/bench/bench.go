package bench

import (
	"fmt"
	"strings"
	"time"

	"github.com/dhruvhegde/cudackpt/pkg/rpc"
)

type Result struct {
	Op       string
	Count    int
	Elapsed  time.Duration
	PerOpUs  float64
	Errors   int
}

func Ping(pid, count int) Result {
	start := time.Now()
	errs := 0
	for i := 0; i < count; i++ {
		cli, err := rpc.Dial(pid)
		if err != nil {
			errs++
			continue
		}
		if err := cli.Ping(); err != nil {
			errs++
		}
		_ = cli.Close()
	}
	elapsed := time.Since(start)
	return Result{Op: "ping", Count: count, Elapsed: elapsed, PerOpUs: usPerOp(count, elapsed), Errors: errs}
}

func Status(pid, count int) Result {
	start := time.Now()
	errs := 0
	cli, err := rpc.Dial(pid)
	if err != nil {
		return Result{Op: "status", Count: count, Errors: count}
	}
	defer cli.Close()
	for i := 0; i < count; i++ {
		if _, err := cli.Status(); err != nil {
			errs++
		}
	}
	elapsed := time.Since(start)
	return Result{Op: "status", Count: count, Elapsed: elapsed, PerOpUs: usPerOp(count, elapsed), Errors: errs}
}

func usPerOp(count int, d time.Duration) float64 {
	if count == 0 {
		return 0
	}
	return float64(d.Microseconds()) / float64(count)
}

func FormatTable(rows ...Result) string {
	var b strings.Builder
	b.WriteString("+--------+-------+-----------+----------+--------+\n")
	b.WriteString("| op     | count | elapsed   | us/op    | errors |\n")
	b.WriteString("+--------+-------+-----------+----------+--------+\n")
	for _, r := range rows {
		b.WriteString(fmt.Sprintf("| %-6s | %5d | %9s | %8.1f | %6d |\n",
			r.Op, r.Count, r.Elapsed.Round(time.Microsecond), r.PerOpUs, r.Errors))
	}
	b.WriteString("+--------+-------+-----------+----------+--------+\n")
	return b.String()
}
