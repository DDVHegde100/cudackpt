package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dhruvhegde/cudackpt/pkg/config"
	"github.com/dhruvhegde/cudackpt/pkg/control"
)

func runServe(cfg config.Config, orc *control.Orchestrator) error {
	in, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	cmd := strings.TrimSpace(string(in))
	if cmd == "" {
		cmd = "ps"
	}
	switch cmd {
	case "ps":
		pids, err := control.ListShims(cfg.RunDir)
		if err != nil {
			return err
		}
		for _, p := range pids {
			st, serr := orc.Status(p)
			if serr != nil {
				fmt.Printf("%d offline\n", p)
				continue
			}
			fmt.Printf("%d %s\n", p, control.StateName(st))
		}
		return nil
	default:
		return fmt.Errorf("unknown serve command %q", cmd)
	}
}

func emitCompletion(shell string, w io.Writer) error {
	switch shell {
	case "bash":
		return emitBashCompletion(w)
	case "zsh":
		return emitZshCompletion(w)
	default:
		return fmt.Errorf("unsupported shell %q (use bash or zsh)", shell)
	}
}

func completionLine(w *bufio.Writer, line string) error {
	_, err := fmt.Fprintln(w, line)
	return err
}

func emitBashCompletion(w io.Writer) error {
	sc := bufio.NewWriter(w)
	lines := []string{
		`# cudackpt bash completion`,
		`_cudackpt_completions() {`,
		`  local cur prev`,
		`  cur="${COMP_WORDS[COMP_CWORD]}"`,
		`  prev="${COMP_WORDS[COMP_CWORD-1]}"`,
		`  local commands="` + strings.Join(completionCommands, " ") + `"`,
		`  local flags="--until-running --timeout --listen --dry-run --stop --pin --root --older-than"`,
		`  if [[ ${COMP_CWORD} -eq 1 ]]; then`,
		`    COMPREPLY=( $(compgen -W "${commands}" -- "${cur}") )`,
		`    return`,
		`  fi`,
		`  case "${prev}" in`,
		`    completion)`,
		`      COMPREPLY=( $(compgen -W "bash zsh" -- "${cur}") )`,
		`      return`,
		`    *)`,
		`      COMPREPLY=( $(compgen -W "${flags}" -- "${cur}") )`,
		`      ;;`,
		`  esac`,
		`}`,
		`complete -F _cudackpt_completions cudackpt`,
	}
	for _, line := range lines {
		if err := completionLine(sc, line); err != nil {
			return err
		}
	}
	return sc.Flush()
}

func emitZshCompletion(w io.Writer) error {
	sc := bufio.NewWriter(w)
	lines := []string{
		`#compdef cudackpt`,
		`_cudackpt() {`,
		`  local -a commands flags`,
		`  commands=(` + strings.Join(completionCommands, " ") + `)`,
		`  flags=(--until-running --timeout --listen --dry-run --stop --pin --root --older-than)`,
		`  if (( CURRENT == 2 )); then`,
		`    _describe command commands`,
		`  elif [[ ${words[2]} == completion ]]; then`,
		`    _describe shell 'bash zsh'`,
		`  else`,
		`    _describe flag flags`,
		`  fi`,
		`}`,
		`_cudackpt`,
	}
	for _, line := range lines {
		if err := completionLine(sc, line); err != nil {
			return err
		}
	}
	return sc.Flush()
}

var completionCommands = []string{
	"checkpoint", "restore", "rollback", "promote", "gc", "metrics", "agent", "serve",
	"freeze", "ping", "resume", "status", "watch", "bench", "diff", "snapshot",
	"gpu-restore", "list", "ps", "inspect", "validate", "report", "stats",
	"health", "version", "completion",
}
