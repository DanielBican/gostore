package tests

import (
	"testing"

	"github.com/DanielBican/gostore/data"
	"github.com/DanielBican/gostore/database"
	"github.com/hashicorp/go-memdb"
)

// NewUnit initializes a test database
func NewUnit(t *testing.T) *memdb.MemDB {

	db, err := database.Open()
	if err != nil {
		t.Fatalf("opening database connection: %v", err)
	}

	t.Log("planting the seed")

	database.Seed(db, TestUsers, TestProducts)

	t.Log("checking seeds")

	users, err := database.ListAllUsers(db)
	if err != nil {
		t.Fatalf("checking user seeds error %v", err)
	}
	products, err := database.ListAllProducts(db)
	if err != nil {
		t.Fatalf("checking product seeds error %v", err)
	}
	if len(users) != len(TestUsers) || len(products) != len(TestProducts) {
		t.Fatal("seed not planted correctly")
	}

	return db
}

// TestUsers contains user test data
var TestUsers = []*data.User{
	&data.User{ID: 1, Name: "Test User 1", Password: `$2a$10$9Sx.OOygxxjH/7g43IyKBuA4yHblzbAnx14861MUM.XB7o/qYkpPm`, Email: "test.user1@company.com"},
	&data.User{ID: 2, Name: "Test User 2", Password: `$2a$10$9Sx.OOygxxjH/7g43IyKBuA4yHblzbAnx14861MUM.XB7o/qYkpPm`, Email: "test.user2@company.com"},
}

// TestProducts contains product test data
var TestProducts = []*data.Product{
	&data.Product{ID: 1, Name: "Product 1", Price: 100, Stock: 2},
	&data.Product{ID: 2, Name: "Product 2", Price: 200, Stock: 3},
}
