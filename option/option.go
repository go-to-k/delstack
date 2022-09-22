package option

import (
	"runtime"
	"time"
)

const AppName = "delstack"
const MaxRetryCount = 5
const CloudFormationWaitNanoSecTime = time.Duration(4500000000000)

var Version = ""
var ConcurrencyNum = runtime.NumCPU()
