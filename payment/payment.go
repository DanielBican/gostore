package payment

import (
	"sync/atomic"
	"time"
)

var dummyPaymentCounter uint64

// Pay simulates a 2 seconds call to a payment system with 50%/50% success result
func Pay() bool {
	c := atomic.AddUint64(&dummyPaymentCounter, 1)
	time.Sleep(2000 * time.Millisecond)
	if c%2 == 0 {
		return false
	}
	return true
}
