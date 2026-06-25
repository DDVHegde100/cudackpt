package ckpterr

import "fmt"

type Code int

const (
	Ok Code = iota
	NotFound
	IO
	RPC
	CUDA
	CRIU
	Unsupported
	Invalid
)

type Error struct {
	Code Code
	Msg  string
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("%s (%d)", e.Msg, e.Code)
}

func E(code Code, msg string) *Error {
	return &Error{Code: code, Msg: msg}
}

func Wrap(code Code, msg string, err error) *Error {
	if err == nil {
		return E(code, msg)
	}
	return E(code, fmt.Sprintf("%s: %v", msg, err))
}
