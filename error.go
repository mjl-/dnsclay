package main

import (
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"

	"github.com/mjl-/bstore"
	"github.com/mjl-/sherpa"
)

var errUser = errors.New("user error")

func _checkf(err error, format string, args ...any) {
	if err == nil {
		return
	}

	msg := fmt.Sprintf(format, args...)

	if err == bstore.ErrAbsent {
		err = fmt.Errorf("%s: Not found", msg)
		slog.Debug("sherpa user error", "err", err)
		panic(&sherpa.Error{Code: "user:notFound", Message: err.Error()})
	}
	switch {
	case errors.Is(err, errUser),
		errors.Is(err, bstore.ErrUnique),
		errors.Is(err, bstore.ErrReference),
		errors.Is(err, bstore.ErrZero):

		err = fmt.Errorf("%s: %v", msg, err)
		slog.Debug("sherpa user error", "err", err)
		panic(&sherpa.Error{Code: "user:error", Message: err.Error()})
	}

	m := msg
	if m != "" {
		m += ": "
	}
	m += err.Error()
	slog.Error("sherpa server error", "err", m)
	debug.PrintStack()
	m = msg + ": " + err.Error()
	_serverError(m)
}

func _serverError(m string) {
	slog.Error("sherpa server error", "err", m)
	panic(&sherpa.Error{Code: "server:error", Message: m})
}

func _checkuserf(err error, format string, args ...any) {
	if err == nil {
		return
	}

	msg := fmt.Sprintf(format, args...)
	m := msg + ": " + err.Error()
	_userError(m)
}

func _userError(m string) {
	slog.Debug("sherpa user error", "err", m)
	panic(&sherpa.Error{Code: "user:error", Message: m})
}
