package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ImageRoot      string
	RunDir         string
	RestoreTimeout time.Duration
	ShimPoll       time.Duration
	MaxRetries     int
	RetryBackoff   time.Duration
}

type LoadResult struct {
	Config   Config
	Warnings []string
}

func Default() Config {
	return mergeEnv(Config{
		ImageRoot:      "/var/lib/cudackpt",
		RunDir:         "/run/cudackpt",
		RestoreTimeout: 60 * time.Second,
		ShimPoll:       200 * time.Millisecond,
		MaxRetries:     3,
		RetryBackoff:   500 * time.Millisecond,
	})
}

func Load() Config {
	r := LoadWithWarnings()
	for _, w := range r.Warnings {
		fmt.Fprintf(os.Stderr, "cudackpt config warning: %s\n", w)
	}
	return r.Config
}

func LoadWithWarnings() LoadResult {
	cfg := Default()
	var warnings []string
	path := os.Getenv("CUDACKPT_CONFIG")
	if path == "" {
		path = "/etc/cudackpt.conf"
	}
	f, err := os.Open(path)
	if err != nil {
		return LoadResult{Config: cfg}
	}
	defer func() { _ = f.Close() }()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			warnings = append(warnings, "invalid line (expected key=value): "+line)
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		switch k {
		case "image_root":
			cfg.ImageRoot = v
		case "run_dir":
			cfg.RunDir = v
		case "restore_timeout":
			if d, err := time.ParseDuration(v); err == nil {
				cfg.RestoreTimeout = d
			} else {
				warnings = append(warnings, "invalid restore_timeout: "+v)
			}
		case "shim_poll":
			if d, err := time.ParseDuration(v); err == nil {
				cfg.ShimPoll = d
			} else {
				warnings = append(warnings, "invalid shim_poll: "+v)
			}
		case "max_retries":
			if n, err := strconv.Atoi(v); err == nil {
				cfg.MaxRetries = n
			} else {
				warnings = append(warnings, "invalid max_retries: "+v)
			}
		case "retry_backoff":
			if d, err := time.ParseDuration(v); err == nil {
				cfg.RetryBackoff = d
			} else {
				warnings = append(warnings, "invalid retry_backoff: "+v)
			}
		default:
			warnings = append(warnings, "unknown config key: "+k)
		}
	}
	cfg = mergeEnv(cfg)
	return LoadResult{Config: cfg, Warnings: warnings}
}

func mergeEnv(cfg Config) Config {
	if v := os.Getenv("CUDACKPT_IMAGE_ROOT"); v != "" {
		cfg.ImageRoot = v
	}
	if v := os.Getenv("CUDACKPT_RUN_DIR"); v != "" {
		cfg.RunDir = v
	}
	if v := os.Getenv("CUDACKPT_RESTORE_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.RestoreTimeout = d
		}
	}
	if v := os.Getenv("CUDACKPT_SHIM_POLL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.ShimPoll = d
		}
	}
	if v := os.Getenv("CUDACKPT_MAX_RETRIES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.MaxRetries = n
		}
	}
	if v := os.Getenv("CUDACKPT_RETRY_BACKOFF"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.RetryBackoff = d
		}
	}
	return cfg
}
