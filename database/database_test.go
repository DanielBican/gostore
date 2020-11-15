package database_test

import (
	"log"
	"testing"

	"github.com/DanielBican/gostore/data"
	"github.com/DanielBican/gostore/database"
	"github.com/DanielBican/gostore/tests"
)

func TestDatabase(t *testing.T) {
	db := tests.NewUnit(t)

	// Test user retrieved by email
	testUser := tests.TestUsers[0]
	u, err := database.GetUserByEmail(db, testUser.Email)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("user from the db %+v", u)

	if u.Name != testUser.Name || u.Password != testUser.Password {
		log.Fatal("users don't match")
	}

	// Test sufficient stock
	testProduct := tests.TestProducts[0]
	cartItem := data.CartItem{ProductName: testProduct.Name, Quantity: 1}
	err = database.VerifyProductAndStock(db, cartItem)
	if err != nil {
		t.Fatal(err)
	}

	// Test exact stock
	cartItem = data.CartItem{ProductName: testProduct.Name, Quantity: testProduct.Stock}
	err = database.VerifyProductAndStock(db, cartItem)
	if err != nil {
		t.Fatal(err)
	}

	// Test insufficient stock
	cartItem = data.CartItem{ProductName: testProduct.Name, Quantity: 3}
	err = database.VerifyProductAndStock(db, cartItem)
	if err == nil {
		t.Fatal("expected insufficient stock error")
	}

	// Test product not found
	cartItem = data.CartItem{ProductName: "Inexistent Product", Quantity: 3}
	err = database.VerifyProductAndStock(db, cartItem)
	if err == nil {
		t.Fatal("expected product not found")
	}

	testProduct1 := tests.TestProducts[0]
	testProduct2 := tests.TestProducts[1]
	// Test checkout insufficient stock
	cart := map[string]int{
		testProduct1.Name: 10,
		testProduct2.Name: 2,
	}
	checkoutResult := database.Checkout(db, cart)
	if checkoutResult.Error == nil {
		t.Fatalf("checkout result %+v", checkoutResult)
	}

	product1, _ := database.GetProductByName(db, testProduct1.Name)
	product2, _ := database.GetProductByName(db, testProduct2.Name)
	if product1.Stock != testProduct1.Stock || product2.Stock != testProduct2.Stock {
		t.Fatal("checkout changed stock")
	}

	// Test checkout without error
	q1 := 2
	q2 := 1
	stockAfterCheckout1 := testProduct1.Stock - q1
	stockAfterCheckout2 := testProduct2.Stock - q2
	cart = map[string]int{
		testProduct1.Name: 2,
		testProduct2.Name: 1,
	}
	checkoutResult = database.Checkout(db, cart)
	if checkoutResult.Error != nil {
		t.Fatalf("checkout result %+v", checkoutResult)
	}

	product1, _ = database.GetProductByName(db, testProduct1.Name)
	product2, _ = database.GetProductByName(db, testProduct2.Name)
	if product1.Stock != stockAfterCheckout1 || product2.Stock != stockAfterCheckout2 {
		t.Fatalf("checkout did not change stock, product1 %v, product2 %v", product1, product2)
	}
}
