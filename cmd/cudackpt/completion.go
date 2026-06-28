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

func emitBashCompletion(w io.Writer) error {
	sc := bufio.NewWriter(w)
	fmt.Fprintln(sc, `# cudackpt bash completion`)
	fmt.Fprintln(sc, `_cudackpt_completions() {`)
	fmt.Fprintln(sc, `  local cur prev`)
	fmt.Fprintln(sc, `  cur="${COMP_WORDS[COMP_CWORD]}"`)
	fmt.Fprintln(sc, `  prev="${COMP_WORDS[COMP_CWORD-1]}"`)
	fmt.Fprintln(sc, `  local commands="`+strings.Join(completionCommands, " ")+`"`)
	fmt.Fprintln(sc, `  local flags="--until-running --timeout --listen --dry-run --stop --pin --root --older-than"`)
	fmt.Fprintln(sc, `  if [[ ${COMP_CWORD} -eq 1 ]]; then`)
	fmt.Fprintln(sc, `    COMPREPLY=( $(compgen -W "${commands}" -- "${cur}") )`)
	fmt.Fprintln(sc, `    return`)
	fmt.Fprintln(sc, `  fi`)
	fmt.Fprintln(sc, `  case "${prev}" in`)
	fmt.Fprintln(sc, `    completion)`)
	fmt.Fprintln(sc, `      COMPREPLY=( $(compgen -W "bash zsh" -- "${cur}") )`)
	fmt.Fprintln(sc, `      return`)
	fmt.Fprintln(sc, `    *)`)
	fmt.Fprintln(sc, `      COMPREPLY=( $(compgen -W "${flags}" -- "${cur}") )`)
	fmt.Fprintln(sc, `      ;;`)
	fmt.Fprintln(sc, `  esac`)
	fmt.Fprintln(sc, `}`)
	fmt.Fprintln(sc, `complete -F _cudackpt_completions cudackpt`)
	return sc.Flush()
}

func emitZshCompletion(w io.Writer) error {
	sc := bufio.NewWriter(w)
	fmt.Fprintln(sc, `#compdef cudackpt`)
	fmt.Fprintln(sc, `_cudackpt() {`)
	fmt.Fprintln(sc, `  local -a commands flags`)
	fmt.Fprintln(sc, `  commands=(`+strings.Join(completionCommands, " ")+`)`)
	fmt.Fprintln(sc, `  flags=(--until-running --timeout --listen --dry-run --stop --pin --root --older-than)`)
	fmt.Fprintln(sc, `  if (( CURRENT == 2 )); then`)
	fmt.Fprintln(sc, `    _describe command commands`)
	fmt.Fprintln(sc, `  elif [[ ${words[2]} == completion ]]; then`)
	fmt.Fprintln(sc, `    _describe shell 'bash zsh'`)
	fmt.Fprintln(sc, `  else`)
	fmt.Fprintln(sc, `    _describe flag flags`)
	fmt.Fprintln(sc, `  fi`)
	fmt.Fprintln(sc, `}`)
	fmt.Fprintln(sc, `_cudackpt`)
	return sc.Flush()
}

var completionCommands = []string{
	"checkpoint", "restore", "rollback", "promote", "gc", "metrics", "agent", "serve",
	"freeze", "ping", "resume", "status", "watch", "bench", "diff", "snapshot",
	"gpu-restore", "list", "ps", "inspect", "validate", "report", "stats",
	"health", "version", "completion",
}
