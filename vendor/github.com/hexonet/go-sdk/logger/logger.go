// Copyright (c) 2018 Kai Schwarz (HEXONET GmbH). All rights reserved.
//
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE.md file.

// Package logger provides functionality around debug outputs/logging of API communication
package logger

import (
	"fmt"

	R "github.com/hexonet/go-sdk/response"
)

// ilogger reflect basic interface for loggers
type ILogger interface {
	Log(post string, r *R.Response, errormsg ...string)
}

// Logger is a struct representing logger for API communication.
type Logger struct{}

// NewLogger represents the constructor for struct Logger.
func NewLogger() *Logger {
	return &Logger{}
}

// Log method to ouput/log api communication
func (c *Logger) Log(post string, r *R.Response, errormsg ...string) {
	fmt.Printf("%s\n", r.GetCommandPlain())
	fmt.Printf("POST: %s\n", post)
	if len(errormsg) > 0 && len(errormsg[0]) > 0 {
		fmt.Printf("HTTP communication failed: %s\n", errormsg)
	}
	fmt.Println(r.GetPlain())
}
