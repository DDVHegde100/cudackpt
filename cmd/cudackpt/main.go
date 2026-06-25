package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/dhruvhegde/cudackpt/internal/version"
	"github.com/dhruvhegde/cudackpt/pkg/agent"
	"github.com/dhruvhegde/cudackpt/pkg/bench"
	"github.com/dhruvhegde/cudackpt/pkg/config"
	"github.com/dhruvhegde/cudackpt/pkg/control"
	"github.com/dhruvhegde/cudackpt/pkg/health"
	"github.com/dhruvhegde/cudackpt/pkg/metrics"
	"github.com/dhruvhegde/cudackpt/pkg/report"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	cfg := config.Load()
	orc := control.New(cfg)
	var err error
	switch os.Args[1] {
	case "checkpoint":
		if len(os.Args) < 3 {
			usage()
			os.Exit(2)
		}
		pid, err := strconv.Atoi(os.Args[2])
		if err != nil {
			die(err)
		}
		out := ""
		if len(os.Args) > 3 {
			out = os.Args[3]
		}
		if err := orc.EnqueueCheckpoint(pid, out); err != nil {
			die(err)
		}
		fmt.Println("checkpoint ok")
	case "restore":
		if len(os.Args) < 3 {
			usage()
			os.Exit(2)
		}
		pid, err := orc.Restore(os.Args[2])
		if err != nil {
			die(err)
		}
		fmt.Printf("restore ok pid=%d\n", pid)
	case "rollback":
		if len(os.Args) < 3 {
			usage()
			os.Exit(2)
		}
		image := os.Args[2]
		stopPID := 0
		for i := 3; i < len(os.Args); i++ {
			if os.Args[i] == "--stop" && i+1 < len(os.Args) {
				stopPID, err = strconv.Atoi(os.Args[i+1])
				if err != nil {
					die(err)
				}
				i++
			}
		}
		pid, err := orc.Rollback(image, stopPID)
		if err != nil {
			die(err)
		}
		fmt.Printf("rollback ok pid=%d\n", pid)
	case "promote":
		if len(os.Args) < 4 {
			usage()
			os.Exit(2)
		}
		pin := control.ParsePromotePin("")
		for i := 4; i < len(os.Args); i++ {
			if os.Args[i] == "--pin" && i+1 < len(os.Args) {
				pin = os.Args[i+1]
				i++
			}
		}
		if err := orc.Promote(control.PromoteOptions{
			Src: os.Args[2], Dest: os.Args[3], PinFile: pin,
		}); err != nil {
			die(err)
		}
		fmt.Println("promote ok")
	case "gc":
		root := cfg.ImageRoot
		maxAge := 14 * 24 * time.Hour
		pin := os.Getenv("CUDACKPT_PIN_FILE")
		dryRun := false
		for i := 2; i < len(os.Args); i++ {
			switch os.Args[i] {
			case "--root":
				if i+1 >= len(os.Args) {
					die(fmt.Errorf("gc: missing value for --root"))
				}
				root = os.Args[i+1]
				i++
			case "--older-than":
				if i+1 >= len(os.Args) {
					die(fmt.Errorf("gc: missing value for --older-than"))
				}
				maxAge, err = parseAge(os.Args[i+1])
				if err != nil {
					die(err)
				}
				i++
			case "--pin":
				if i+1 >= len(os.Args) {
					die(fmt.Errorf("gc: missing value for --pin"))
				}
				pin = os.Args[i+1]
				i++
			case "--dry-run":
				dryRun = true
			default:
				die(fmt.Errorf("gc: unknown flag %q", os.Args[i]))
			}
		}
		_, removed, err := control.RunImageGC(control.GCOptions{
			Root: root, MaxAge: maxAge, PinFile: pin,
		}, dryRun)
		if err != nil {
			die(err)
		}
		for _, p := range removed {
			fmt.Println(p)
		}
		metrics.Default.Add(metrics.GCRemovedTotal, uint64(len(removed)))
		agent.RefreshGauges(cfg)
		if dryRun {
			fmt.Fprintf(os.Stderr, "gc dry-run: would remove %d paths\n", len(removed))
		} else {
			fmt.Fprintf(os.Stderr, "gc removed %d paths\n", len(removed))
		}
	case "freeze", "ping", "resume", "status":
		if len(os.Args) < 3 {
			usage()
			os.Exit(2)
		}
		pid, err := strconv.Atoi(os.Args[2])
		if err != nil {
			die(err)
		}
		switch os.Args[1] {
		case "ping":
			err = orc.Ping(pid)
		case "freeze":
			err = orc.Freeze(pid)
		case "resume":
			err = orc.Resume(pid)
		case "status":
			var st uint32
			st, err = orc.Status(pid)
			if err == nil {
				fmt.Printf("%s (%d)\n", control.StateName(st), st)
			}
		}
		if err != nil {
			die(err)
		}
		if os.Args[1] != "status" {
			fmt.Printf("%s ok\n", os.Args[1])
		}
	case "watch":
		if len(os.Args) < 3 {
			usage()
			os.Exit(2)
		}
		pid, err := strconv.Atoi(os.Args[2])
		if err != nil {
			die(err)
		}
		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer cancel()
		if err := control.WatchShim(orc, pid, cfg.ShimPoll, ctx.Done()); err != nil {
			die(err)
		}
	case "bench":
		if len(os.Args) < 3 {
			usage()
			os.Exit(2)
		}
		pid, err := strconv.Atoi(os.Args[2])
		if err != nil {
			die(err)
		}
		n := 100
		if len(os.Args) > 3 {
			n, err = strconv.Atoi(os.Args[3])
			if err != nil {
				die(err)
			}
		}
		ping := bench.Ping(pid, n)
		st := bench.Status(pid, n)
		fmt.Print(bench.FormatTable(ping, st))
	case "diff":
		if len(os.Args) < 4 {
			usage()
			os.Exit(2)
		}
		rep, err := control.CompareImages(os.Args[2], os.Args[3])
		if err != nil {
			die(err)
		}
		fmt.Print(control.FormatDrift(rep))
	case "snapshot":
		if len(os.Args) < 4 {
			usage()
			os.Exit(2)
		}
		pid, err := strconv.Atoi(os.Args[2])
		if err != nil {
			die(err)
		}
		if err := orc.Snapshot(pid, os.Args[3]); err != nil {
			die(err)
		}
		fmt.Println("snapshot ok")
	case "gpu-restore":
		if len(os.Args) < 4 {
			usage()
			os.Exit(2)
		}
		pid, err := strconv.Atoi(os.Args[2])
		if err != nil {
			die(err)
		}
		if err := orc.GpuRestore(pid, os.Args[3]); err != nil {
			die(err)
		}
		fmt.Println("gpu-restore ok")
	case "list":
		root := cfg.ImageRoot
		if len(os.Args) > 2 {
			root = os.Args[2]
		}
		imgs, err := control.ListImages(root)
		if err != nil {
			die(err)
		}
		for _, p := range imgs {
			fmt.Println(p)
		}
	case "ps":
		verbose := len(os.Args) > 2 && os.Args[2] == "-v"
		pids, err := control.ListShims(cfg.RunDir)
		if err != nil {
			die(err)
		}
		for _, p := range pids {
			if !verbose {
				fmt.Println(p)
				continue
			}
			st, serr := orc.Status(p)
			if serr != nil {
				fmt.Printf("%d offline\n", p)
				continue
			}
			fmt.Printf("%d %s\n", p, control.StateName(st))
		}
	case "inspect":
		if len(os.Args) < 3 {
			usage()
			os.Exit(2)
		}
		if err := control.InspectImage(os.Args[2]); err != nil {
			die(err)
		}
	case "validate":
		if len(os.Args) < 3 {
			usage()
			os.Exit(2)
		}
		if err := orc.ValidateImage(os.Args[2]); err != nil {
			die(err)
		}
		fmt.Println("validate ok")
	case "report":
		if len(os.Args) < 3 {
			usage()
			os.Exit(2)
		}
		out, err := report.RenderImage(os.Args[2])
		if err != nil {
			die(err)
		}
		fmt.Print(out)
	case "stats":
		if len(os.Args) < 3 {
			usage()
			os.Exit(2)
		}
		pid, err := strconv.Atoi(os.Args[2])
		if err != nil {
			die(err)
		}
		st, err := orc.Stats(pid)
		if err != nil {
			die(err)
		}
		fmt.Printf("state=%d allocs=%d bytes=%d streams=%d modules=%d symbols=%d events=%d ctxs=%d unsupported=%d\n",
			st.State, st.AllocCount, st.TotalBytes, st.StreamCount, st.ModuleCount,
			st.SymbolCount, st.EventCount, st.CtxCount, st.UnsupportedCode)
	case "health":
		deep := len(os.Args) > 2 && os.Args[2] == "-d"
		var st health.Status
		if deep {
			st = health.DeepProbe()
		} else {
			st = health.Probe()
		}
		fmt.Print(health.Format(st))
		if !st.OK {
			os.Exit(1)
		}
	case "version":
		fmt.Println(version.String())
	case "metrics":
		addr := ":9090"
		for i := 2; i < len(os.Args); i++ {
			if os.Args[i] == "--listen" && i+1 < len(os.Args) {
				addr = os.Args[i+1]
				i++
			} else {
				die(fmt.Errorf("metrics: unknown flag %q", os.Args[i]))
			}
		}
		agent.RefreshGauges(cfg)
		fmt.Fprintf(os.Stderr, "metrics listening on %s/metrics\n", addr)
		if err := metrics.Serve(addr, metrics.Default); err != nil {
			die(err)
		}
	case "agent":
		opts := agent.OptionsFromConfig(cfg)
		for i := 2; i < len(os.Args); i++ {
			switch os.Args[i] {
			case "--listen":
				if i+1 >= len(os.Args) {
					die(fmt.Errorf("agent: missing value for --listen"))
				}
				opts.Listen = os.Args[i+1]
				i++
			default:
				die(fmt.Errorf("agent: unknown flag %q", os.Args[i]))
			}
		}
		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer cancel()
		if err := agent.Run(ctx, opts); err != nil {
			die(err)
		}
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: cudackpt checkpoint <pid> [dir]\n")
	fmt.Fprintf(os.Stderr, "       cudackpt restore <image>\n")
	fmt.Fprintf(os.Stderr, "       cudackpt rollback <image> [--stop <pid>]\n")
	fmt.Fprintf(os.Stderr, "       cudackpt promote <src> <dest> [--pin file]\n")
	fmt.Fprintf(os.Stderr, "       cudackpt gc [--root dir] [--older-than 14d] [--pin file] [--dry-run]\n")
	fmt.Fprintf(os.Stderr, "       cudackpt metrics [--listen addr]\n")
	fmt.Fprintf(os.Stderr, "       cudackpt agent [--listen addr]\n")
	fmt.Fprintf(os.Stderr, "       cudackpt freeze|ping|resume|status <pid>\n")
	fmt.Fprintf(os.Stderr, "       cudackpt watch <pid>\n")
	fmt.Fprintf(os.Stderr, "       cudackpt bench <pid> [count]\n")
	fmt.Fprintf(os.Stderr, "       cudackpt diff <image-a> <image-b>\n")
	fmt.Fprintf(os.Stderr, "       cudackpt snapshot <pid> <dir>\n")
	fmt.Fprintf(os.Stderr, "       cudackpt gpu-restore <pid> <dir>\n")
	fmt.Fprintf(os.Stderr, "       cudackpt list [root]\n")
	fmt.Fprintf(os.Stderr, "       cudackpt ps [-v]\n")
	fmt.Fprintf(os.Stderr, "       cudackpt inspect <image>\n")
	fmt.Fprintf(os.Stderr, "       cudackpt validate <image>\n")
	fmt.Fprintf(os.Stderr, "       cudackpt report <image>\n")
	fmt.Fprintf(os.Stderr, "       cudackpt stats <pid>\n")
	fmt.Fprintf(os.Stderr, "       cudackpt health [-d]\n")
	fmt.Fprintf(os.Stderr, "       cudackpt version\n")
}

func die(err error) {
	fmt.Fprintf(os.Stderr, "cudackpt: %v\n", err)
	os.Exit(1)
}

func parseAge(s string) (time.Duration, error) {
	if strings.HasSuffix(s, "d") {
		n, err := strconv.Atoi(strings.TrimSuffix(s, "d"))
		if err != nil {
			return 0, err
		}
		return time.Duration(n) * 24 * time.Hour, nil
	}
	return time.ParseDuration(s)
}
