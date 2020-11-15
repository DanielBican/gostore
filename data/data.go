package data

// Login represents the request body for login endpoint
type Login struct {
	Email    string
	Password string
}

// CartItem represents the request body for add to cart action
type CartItem struct {
	ProductName string
	Quantity    int
}

// User represents the data for an API user
type User struct {
	ID       int64
	Name     string
	Password string
	Email    string
}

// Product represents the data stored in the DB for a product
type Product struct {
	ID    int64
	Name  string
	Price float64
	Stock int
}
