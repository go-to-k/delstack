package option

import (
	"runtime"
	"runtime/debug"
	"time"
)

const AppName = "delstack"
const MaxRetryCount = 10
const CloudFormationWaitNanoSecTime = time.Duration(4500000000000)

var Version = ""
var Revision = ""
var ConcurrencyNum = runtime.NumCPU()

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
