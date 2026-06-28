package main

import (
	"fmt"
	"io"
	"strings"
)

var completionCommands = []string{
	"checkpoint", "restore", "rollback", "promote", "gc", "metrics", "agent",
	"freeze", "ping", "resume", "status", "watch", "bench", "diff", "snapshot",
	"gpu-restore", "list", "ps", "inspect", "validate", "report", "stats",
	"health", "version", "completion",
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
	fmt.Fprintln(w, `# cudackpt bash completion`)
	fmt.Fprintln(w, `_cudackpt_completions() {`)
	fmt.Fprintln(w, `  local cur prev`)
	fmt.Fprintln(w, `  cur="${COMP_WORDS[COMP_CWORD]}"`)
	fmt.Fprintln(w, `  prev="${COMP_WORDS[COMP_CWORD-1]}"`)
	fmt.Fprintln(w, `  local commands="`+strings.Join(completionCommands, " ")+`"`)
	fmt.Fprintln(w, `  local flags="--until-running --timeout --listen --dry-run --stop --pin --root --older-than"`)
	fmt.Fprintln(w, `  if [[ ${COMP_CWORD} -eq 1 ]]; then`)
	fmt.Fprintln(w, `    COMPREPLY=( $(compgen -W "${commands}" -- "${cur}") )`)
	fmt.Fprintln(w, `    return`)
	fmt.Fprintln(w, `  fi`)
	fmt.Fprintln(w, `  case "${prev}" in`)
	fmt.Fprintln(w, `    watch|bench|checkpoint|freeze|ping|resume|status|snapshot|gpu-restore|stats|rollback)`)
	fmt.Fprintln(w, `      ;;`)
	fmt.Fprintln(w, `    completion)`)
	fmt.Fprintln(w, `      COMPREPLY=( $(compgen -W "bash zsh" -- "${cur}") )`)
	fmt.Fprintln(w, `      return`)
	fmt.Fprintln(w, `    *)`)
	fmt.Fprintln(w, `      COMPREPLY=( $(compgen -W "${flags}" -- "${cur}") )`)
	fmt.Fprintln(w, `      ;;`)
	fmt.Fprintln(w, `  esac`)
	fmt.Fprintln(w, `}`)
	fmt.Fprintln(w, `complete -F _cudackpt_completions cudackpt`)
	return nil
}

func emitZshCompletion(w io.Writer) error {
	fmt.Fprintln(w, `#compdef cudackpt`)
	fmt.Fprintln(w, `_cudackpt() {`)
	fmt.Fprintln(w, `  local -a commands flags`)
	fmt.Fprintln(w, `  commands=(`+strings.Join(completionCommands, " ")+`)`)
	fmt.Fprintln(w, `  flags=(--until-running --timeout --listen --dry-run --stop --pin --root --older-than)`)
	fmt.Fprintln(w, `  if (( CURRENT == 2 )); then`)
	fmt.Fprintln(w, `    _describe command commands`)
	fmt.Fprintln(w, `  elif [[ ${words[2]} == completion ]]; then`)
	fmt.Fprintln(w, `    _describe shell 'bash zsh'`)
	fmt.Fprintln(w, `  else`)
	fmt.Fprintln(w, `    _describe flag flags`)
	fmt.Fprintln(w, `  fi`)
	fmt.Fprintln(w, `}`)
	fmt.Fprintln(w, `_cudackpt`)
	return nil
}
