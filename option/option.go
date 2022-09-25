package option

import (
	"runtime"
	"time"
)

const AppName = "delstack"
const MaxRetryCount = 10
const CloudFormationWaitNanoSecTime = time.Duration(4500000000000)

var Version = ""
var Revision = ""
var ConcurrencyNum = runtime.NumCPU()
