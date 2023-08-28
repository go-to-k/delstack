package operation

import (
	"fmt"
	"testing"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	fmt.Println()
	fmt.Println("==========================================")
	fmt.Println("========== Start Test: operation =========")
	fmt.Println("==========================================")
	goleak.VerifyTestMain(m)
}
