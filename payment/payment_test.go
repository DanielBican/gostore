package payment_test

import (
	"sync"
	"testing"

	"github.com/DanielBican/gostore/payment"
)

func TestPay(t *testing.T) {

	var result1 bool
	var result2 bool

	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		result1 = payment.Pay()
		wg.Done()
	}()
	go func() {
		result2 = payment.Pay()
		wg.Done()
	}()
	wg.Wait()

	t.Logf("result1: %t, result2: %t", result1, result2)

	if result1 == result2 {
		t.Fatal("payment is not dummy")
	}
}
