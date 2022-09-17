package option

import "runtime"

var ConcurrencyNum = runtime.NumCPU()
var MaxRetryCount = 5
