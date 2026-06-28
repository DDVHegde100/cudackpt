package metrics

import (
	"fmt"
	"io"
	"net/http"
	"sort"
	"sync"
)

type Registry struct {
	mu       sync.Mutex
	counters map[string]uint64
	gauges   map[string]float64
}

func NewRegistry() *Registry {
	return &Registry{
		counters: make(map[string]uint64),
		gauges:   make(map[string]float64),
	}
}

var Default = NewRegistry()

func (r *Registry) Inc(name string) {
	r.Add(name, 1)
}

func (r *Registry) Add(name string, delta uint64) {
	r.mu.Lock()
	r.counters[name] += delta
	r.mu.Unlock()
}

func (r *Registry) Set(name string, v float64) {
	r.mu.Lock()
	r.gauges[name] = v
	r.mu.Unlock()
}

func (r *Registry) Snapshot() (map[string]uint64, map[string]float64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c := make(map[string]uint64, len(r.counters))
	for k, v := range r.counters {
		c[k] = v
	}
	g := make(map[string]float64, len(r.gauges))
	for k, v := range r.gauges {
		g[k] = v
	}
	return c, g
}

func (r *Registry) WritePrometheus(w io.Writer) {
	counters, gauges := r.Snapshot()
	names := make([]string, 0, len(counters)+len(gauges))
	seen := make(map[string]struct{})
	for k := range counters {
		names = append(names, k)
		seen[k] = struct{}{}
	}
	for k := range gauges {
		if _, ok := seen[k]; !ok {
			names = append(names, k)
		}
	}
	sort.Strings(names)
	for _, name := range names {
		if v, ok := counters[name]; ok {
			_, _ = fmt.Fprintf(w, "# TYPE %s counter\n", name)
			_, _ = fmt.Fprintf(w, "%s %d\n", name, v)
		}
		if v, ok := gauges[name]; ok {
			_, _ = fmt.Fprintf(w, "# TYPE %s gauge\n", name)
			_, _ = fmt.Fprintf(w, "%s %g\n", name, v)
		}
	}
}

func (r *Registry) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/metrics" {
			http.NotFound(w, req)
			return
		}
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		r.WritePrometheus(w)
	})
}

func Serve(addr string, reg *Registry) error {
	if reg == nil {
		reg = Default
	}
	srv := &http.Server{Addr: addr, Handler: reg.Handler()}
	return srv.ListenAndServe()
}

const (
	CheckpointsTotal      = "cudackpt_checkpoints_total"
	CheckpointFailures    = "cudackpt_checkpoint_failures_total"
	RestoresTotal         = "cudackpt_restores_total"
	RestoreFailuresTotal  = "cudackpt_restore_failures_total"
	RollbacksTotal        = "cudackpt_rollbacks_total"
	GCRemovedTotal        = "cudackpt_gc_removed_total"
	GCErrorsTotal         = "cudackpt_gc_errors_total"
	ImagesGauge           = "cudackpt_images"
	ShimsGauge            = "cudackpt_shims"
)
