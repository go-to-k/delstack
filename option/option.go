package option

import (
	"runtime"
	"time"
)

var AppName = "delstack"
var Version = ""
var ConcurrencyNum = runtime.NumCPU()
var MaxRetryCount = 5
var CloudFormationWaitNanoSecTime = time.Duration(5400000000000)
