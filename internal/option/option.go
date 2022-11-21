package option

import (
	"runtime/debug"
)

const AppName = "delstack"

var Version = ""
var Revision = ""

func IsDebug() bool {
	if Version == "" || Revision != "" {
		return true
	}
	return false
}

func GetVersion() string {
	if Version != "" && Revision != "" {
		return Version + "-" + Revision
	}
	if Version != "" {
		return Version
	}

	i, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}
	return i.Main.Version
}
