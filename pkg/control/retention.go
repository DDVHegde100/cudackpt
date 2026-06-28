package control

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dhruvhegde/cudackpt/pkg/image"
)

type GCOptions struct {
	Root     string
	MaxAge   time.Duration
	PinFile  string
	Now      time.Time
}

type GCPlan struct {
	Remove []string
	Keep   []string
}

func LoadPinnedPaths(path string) (map[string]struct{}, error) {
	out := make(map[string]struct{})
	if path == "" {
		return out, nil
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return out, nil
		}
		return nil, err
	}
	defer func() { _ = f.Close() }()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out[filepath.Clean(line)] = struct{}{}
	}
	return out, sc.Err()
}

func PlanImageGC(opts GCOptions) (GCPlan, error) {
	if opts.Root == "" {
		opts.Root = "/var/lib/cudackpt"
	}
	if opts.Now.IsZero() {
		opts.Now = time.Now()
	}
	pinned, err := LoadPinnedPaths(opts.PinFile)
	if err != nil {
		return GCPlan{}, err
	}
	ents, err := os.ReadDir(opts.Root)
	if err != nil {
		if os.IsNotExist(err) {
			return GCPlan{}, nil
		}
		return GCPlan{}, err
	}
	var plan GCPlan
	for _, e := range ents {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		path := filepath.Join(opts.Root, name)
		clean := filepath.Clean(path)
		if _, ok := pinned[clean]; ok {
			plan.Keep = append(plan.Keep, path)
			continue
		}
		if strings.HasSuffix(name, ".staging") {
			plan.Remove = append(plan.Remove, path)
			continue
		}
		if !image.IsComplete(path) {
			plan.Keep = append(plan.Keep, path)
			continue
		}
		st, err := os.Stat(path)
		if err != nil {
			continue
		}
		if opts.MaxAge > 0 && opts.Now.Sub(st.ModTime()) >= opts.MaxAge {
			plan.Remove = append(plan.Remove, path)
			continue
		}
		plan.Keep = append(plan.Keep, path)
	}
	return plan, nil
}

func ApplyGCPlan(plan GCPlan, dryRun bool) ([]string, error) {
	var removed []string
	for _, path := range plan.Remove {
		if dryRun {
			removed = append(removed, path)
			continue
		}
		if err := os.RemoveAll(path); err != nil {
			return removed, err
		}
		removed = append(removed, path)
	}
	return removed, nil
}

func RunImageGC(opts GCOptions, dryRun bool) (GCPlan, []string, error) {
	plan, err := PlanImageGC(opts)
	if err != nil {
		return GCPlan{}, nil, err
	}
	removed, err := ApplyGCPlan(plan, dryRun)
	return plan, removed, err
}
