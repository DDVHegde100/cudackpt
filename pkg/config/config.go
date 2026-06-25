package config

import (
	"os"
	"time"
)

type Config struct {
	ImageRoot      string
	RunDir         string
	RestoreTimeout time.Duration
	ShimPoll       time.Duration
}

func Default() Config {
	root := "/var/lib/cudackpt"
	if v := os.Getenv("CUDACKPT_IMAGE_ROOT"); v != "" {
		root = v
	}
	run := "/run/cudackpt"
	if v := os.Getenv("CUDACKPT_RUN_DIR"); v != "" {
		run = v
	}
	return Config{
		ImageRoot:      root,
		RunDir:         run,
		RestoreTimeout: 60 * time.Second,
		ShimPoll:       200 * time.Millisecond,
	}
}
