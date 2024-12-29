package main

import (
	"runtime"
	"runtime/debug"
)

var version = "(devel)"

func init() {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	version = buildInfo.Main.Version
	if version == "(devel)" {
		for _, setting := range buildInfo.Settings {
			if setting.Key == "vcs.revision" {
				version = setting.Value
				break
			}
		}
	}
	version += "-" + runtime.Version()
}
