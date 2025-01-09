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

	msg := fmt.Errorf(format, args...)

	if err == bstore.ErrAbsent {
		err = fmt.Errorf("%s: Not found", msg)
		slog.Debug("sherpa user error", "err", err)
		panic(&sherpa.Error{Code: "user:notFound", Message: err.Error()})
	}
	xerr := fmt.Errorf("%w: %w", msg, err)
	switch {
	case errors.Is(xerr, errUser),
		errors.Is(xerr, bstore.ErrUnique),
		errors.Is(xerr, bstore.ErrReference),
		errors.Is(xerr, bstore.ErrZero):

		slog.Debug("sherpa user error", "err", xerr)
		panic(&sherpa.Error{Code: "user:error", Message: xerr.Error()})
	}

	m := msg.Error()
	if m != "" {
		m += ": "
	}
	m += err.Error()
	slog.Error("sherpa server error", "err", m)
	debug.PrintStack()
	m = msg.Error() + ": " + err.Error()
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
