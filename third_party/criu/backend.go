package criu

type Backend interface {
	Dump(pid int, dir string) error
	Restore(dir string, logFile string, env []string) (int, error)
}

var _ Backend = (*Wrapper)(nil)
