package database

import (
	"fmt"
	"log"

	"github.com/hashicorp/go-memdb"

	"github.com/DanielBican/gostore/data"
	"github.com/DanielBican/gostore/payment"
)

// CheckoutResult represents the data returned by the checkout operation
type CheckoutResult struct {
	Result bool
	Error  error
}

// Open creates a new MemDB with the given schema
func Open() (*memdb.MemDB, error) {
	schema := &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"user": &memdb.TableSchema{
				Name: "user",
				Indexes: map[string]*memdb.IndexSchema{
					"id": &memdb.IndexSchema{
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.IntFieldIndex{Field: "ID"},
					},
					"email": &memdb.IndexSchema{
						Name:    "email",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "Email"},
					},
					"password": &memdb.IndexSchema{
						Name:    "password",
						Indexer: &memdb.StringFieldIndex{Field: "Password"},
					},
				},
			},
			"product": &memdb.TableSchema{
				Name: "product",
				Indexes: map[string]*memdb.IndexSchema{
					"id": &memdb.IndexSchema{
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.IntFieldIndex{Field: "ID"},
					},
					"name": &memdb.IndexSchema{
						Name:    "name",
						Indexer: &memdb.StringFieldIndex{Field: "Name"},
					},
				},
			},
		},
	}

	return memdb.NewMemDB(schema)
}

// GetUserByEmail returns an user based on its email
func GetUserByEmail(db *memdb.MemDB, email string) (*data.User, error) {

	txn := db.Txn(false)
	defer txn.Abort()

	raw, err := txn.First("user", "email", email)
	if err != nil {
		return nil, err
	}

	if raw == nil {
		return nil, fmt.Errorf("user with email %s not found", email)
	}

	return raw.(*data.User), nil
}

// VerifyProductAndStock returns a product based on its name and checks if the stock is sufficient
func VerifyProductAndStock(db *memdb.MemDB, cartItem data.CartItem) error {

	txn := db.Txn(true)
	defer txn.Abort()

	raw, err := txn.First("product", "name", cartItem.ProductName)
	if err != nil {
		return err
	}

	if raw == nil {
		return fmt.Errorf("product %s not found", cartItem.ProductName)
	}

	p := raw.(*data.Product)
	if p.Stock < cartItem.Quantity {
		return fmt.Errorf("insufficient stock for %s", p.Name)
	}

	return nil
}

// Checkout verifies the stock and returns the total price for cart items
func Checkout(db *memdb.MemDB, cart map[string]int) *CheckoutResult {

	txn := db.Txn(true)

	for name, q := range cart {
		raw, err := txn.First("product", "name", name)
		if err != nil {
			txn.Abort()
			return &CheckoutResult{false, err}
		}

		p := raw.(*data.Product)
		if p.Stock < q {
			txn.Abort()
			return &CheckoutResult{false, fmt.Errorf("insufficient stock for %s", p.Name)}
		}

		p.Stock = p.Stock - q
	}

	if payment.Pay() {
		txn.Commit()
		return &CheckoutResult{true, nil}
	} else {
		txn.Abort()
		return &CheckoutResult{false, nil}
	}
}

// Seed inserts default users and products
func Seed(db *memdb.MemDB, users []*data.User, products []*data.Product) error {

	txn := db.Txn(true)

	// Insert users
	for _, u := range users {
		if err := txn.Insert("user", u); err != nil {
			panic(err)
		}
	}

	// Insert products
	for _, p := range products {
		if err := txn.Insert("product", p); err != nil {
			panic(err)
		}
	}

	txn.Commit()

	log.Println("seed is planted into the DB")

	return nil
}

// ListAllUsers retrieves all users
func ListAllUsers(db *memdb.MemDB) ([]*data.User, error) {

	txn := db.Txn(false)
	defer txn.Abort()

	it, err := txn.Get("user", "id")
	if err != nil {
		return nil, err
	}

	users := []*data.User{}
	for obj := it.Next(); obj != nil; obj = it.Next() {
		u := obj.(*data.User)
		users = append(users, u)
	}

	return users, nil
}

// ListAllProducts retrieves all products
func ListAllProducts(db *memdb.MemDB) ([]*data.Product, error) {

	txn := db.Txn(false)
	defer txn.Abort()

	it, err := txn.Get("product", "id")
	if err != nil {
		return nil, err
	}

	products := []*data.Product{}
	for obj := it.Next(); obj != nil; obj = it.Next() {
		p := obj.(*data.Product)
		products = append(products, p)
	}

	return products, nil
}

// GetProductByName returns a product using its name
func GetProductByName(db *memdb.MemDB, name string) (*data.Product, error) {

	txn := db.Txn(false)
	defer txn.Abort()

	raw, err := txn.First("product", "name", name)
	if err != nil {
		return nil, err
	}

	if raw == nil {
		return nil, fmt.Errorf("product %s not found", name)
	}

	return raw.(*data.Product), nil
}

// Users array contains user seeds
// Passwords are hashed with bcrypt
var Users = []*data.User{
	&data.User{ID: 1, Name: "Jon Doe", Password: `$2a$10$9Sx.OOygxxjH/7g43IyKBuA4yHblzbAnx14861MUM.XB7o/qYkpPm`, Email: "jon.doe@company.com"},
	&data.User{ID: 2, Name: "Jon Doe 2", Password: `$2a$10$9Sx.OOygxxjH/7g43IyKBuA4yHblzbAnx14861MUM.XB7o/qYkpPm`, Email: "jon.doe2@company.com"},
}

// Products array contains product seeds
var Products = []*data.Product{
	&data.Product{ID: 1, Name: "Product 1", Price: 100, Stock: 2},
	&data.Product{ID: 2, Name: "Product 2", Price: 200, Stock: 3},
}
