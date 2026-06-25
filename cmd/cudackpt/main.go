package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/dhruvhegde/cudackpt/pkg/config"
	"github.com/dhruvhegde/cudackpt/pkg/control"
	"github.com/dhruvhegde/cudackpt/pkg/health"
	"github.com/dhruvhegde/cudackpt/pkg/report"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	cfg := config.Default()
	orc := control.New(cfg)
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
		if err := orc.Checkpoint(pid, out); err != nil {
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
	case "health":
		st := health.Probe()
		fmt.Print(health.Format(st))
		if !st.OK {
			os.Exit(1)
		}
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: cudackpt checkpoint <pid> [dir]\n")
	fmt.Fprintf(os.Stderr, "       cudackpt restore <image>\n")
	fmt.Fprintf(os.Stderr, "       cudackpt freeze|ping|resume|status <pid>\n")
	fmt.Fprintf(os.Stderr, "       cudackpt snapshot <pid> <dir>\n")
	fmt.Fprintf(os.Stderr, "       cudackpt gpu-restore <pid> <dir>\n")
	fmt.Fprintf(os.Stderr, "       cudackpt list [root]\n")
	fmt.Fprintf(os.Stderr, "       cudackpt ps [-v]\n")
	fmt.Fprintf(os.Stderr, "       cudackpt inspect <image>\n")
	fmt.Fprintf(os.Stderr, "       cudackpt validate <image>\n")
	fmt.Fprintf(os.Stderr, "       cudackpt report <image>\n")
	fmt.Fprintf(os.Stderr, "       cudackpt health\n")
}

func die(err error) {
	fmt.Fprintf(os.Stderr, "cudackpt: %v\n", err)
	os.Exit(1)
}
