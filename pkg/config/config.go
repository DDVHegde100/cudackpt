package config

import (
	"bufio"
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
	cfg := Default()
	path := os.Getenv("CUDACKPT_CONFIG")
	if path == "" {
		path = "/etc/cudackpt.conf"
	}
	f, err := os.Open(path)
	if err != nil {
		return cfg
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
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
			}
		case "shim_poll":
			if d, err := time.ParseDuration(v); err == nil {
				cfg.ShimPoll = d
			}
		case "max_retries":
			if n, err := strconv.Atoi(v); err == nil {
				cfg.MaxRetries = n
			}
		case "retry_backoff":
			if d, err := time.ParseDuration(v); err == nil {
				cfg.RetryBackoff = d
			}
		}
	}
	return mergeEnv(cfg)
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
