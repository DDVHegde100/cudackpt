package agent

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/dhruvhegde/cudackpt/pkg/config"
	"github.com/dhruvhegde/cudackpt/pkg/control"
	"github.com/dhruvhegde/cudackpt/pkg/health"
	jlog "github.com/dhruvhegde/cudackpt/pkg/log"
	"github.com/dhruvhegde/cudackpt/pkg/metrics"
)

type Options struct {
	Config          config.Config
	Listen          string
	RefreshInterval time.Duration
	GCInterval      time.Duration
	GCMaxAge        time.Duration
	PinFile         string
}

func OptionsFromConfig(cfg config.Config) Options {
	opts := Options{
		Config:          cfg,
		Listen:          "127.0.0.1:9090",
		RefreshInterval: 15 * time.Second,
	}
	if v := os.Getenv("CUDACKPT_METRICS_ADDR"); v != "" {
		opts.Listen = v
	}
	if v := os.Getenv("CUDACKPT_AGENT_GC_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			opts.GCInterval = d
		}
	}
	if v := os.Getenv("CUDACKPT_PIN_FILE"); v != "" {
		opts.PinFile = v
	}
	if v := os.Getenv("CUDACKPT_AGENT_GC_MAX_AGE"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			opts.GCMaxAge = d
		}
	}
	if opts.GCMaxAge == 0 {
		opts.GCMaxAge = 14 * 24 * time.Hour
	}
	return opts
}

func RefreshGauges(cfg config.Config) {
	imgs, err := control.ListImages(cfg.ImageRoot)
	if err == nil {
		metrics.Default.Set(metrics.ImagesGauge, float64(len(imgs)))
	}
	shims, err := control.ListShims(cfg.RunDir)
	if err == nil {
		metrics.Default.Set(metrics.ShimsGauge, float64(len(shims)))
	}
}

func Run(ctx context.Context, opts Options) error {
	if opts.Listen == "" {
		opts.Listen = "127.0.0.1:9090"
	}
	if opts.RefreshInterval <= 0 {
		opts.RefreshInterval = 15 * time.Second
	}
	mux := http.NewServeMux()
	mux.Handle("/metrics", metrics.Default.Handler())
	mux.HandleFunc("/health", healthHandler(opts.Config))
	srv := &http.Server{Addr: opts.Listen, Handler: mux}
	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutCtx)
	}()
	go runLoop(ctx, opts)
	fmt.Fprintf(os.Stderr, "cudackpt agent listening on http://%s/metrics (health: /health)\n", opts.Listen)
	err := srv.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func runLoop(ctx context.Context, opts Options) {
	refresh := time.NewTicker(opts.RefreshInterval)
	defer refresh.Stop()
	var gc <-chan time.Time
	if opts.GCInterval > 0 {
		t := time.NewTicker(opts.GCInterval)
		defer t.Stop()
		gc = t.C
	}
	RefreshGauges(opts.Config)
	for {
		select {
		case <-ctx.Done():
			return
		case <-refresh.C:
			RefreshGauges(opts.Config)
		case <-gc:
			runGC(opts)
		}
	}
}

func runGC(opts Options) {
	_, removed, err := control.RunImageGC(control.GCOptions{
		Root:    opts.Config.ImageRoot,
		MaxAge:  opts.GCMaxAge,
		PinFile: opts.PinFile,
	}, false)
	if err != nil {
		jlog.Error("agent_gc", err, map[string]any{"root": opts.Config.ImageRoot})
		metrics.Default.Inc(metrics.GCErrorsTotal)
		return
	}
	if len(removed) > 0 {
		metrics.Default.Add(metrics.GCRemovedTotal, uint64(len(removed)))
		RefreshGauges(opts.Config)
	}
}

func healthHandler(cfg config.Config) http.HandlerFunc {
	deepDefault := os.Getenv("CUDACKPT_AGENT_DEEP_HEALTH") == "1"
	return func(w http.ResponseWriter, r *http.Request) {
		deep := deepDefault || r.URL.Query().Get("deep") == "1"
		var st health.Status
		if deep {
			st = health.DeepProbeWith(cfg.RunDir)
		} else {
			st = health.ProbeWith(cfg.RunDir)
		}
		body := health.Format(st)
		if st.OK {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		_, _ = w.Write([]byte(body))
	}
}
