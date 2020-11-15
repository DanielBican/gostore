# Go Store

REST API built with net/http Go package, github.com/gorilla/sessions and in-memory DB https://github.com/hashicorp/go-memdb

By default it will start on localhost:8080 and it will expose following endpoints
* POST http://localhost:8080/v1/login with body 
`{
  "Email": "jon.doe2@company.com",
  "Password": "1234"
}`
* POST http://localhost:8080/v1/cart with body
`{
  "ProductName": "Product 2",
  "Quantity": 1
}`
* POST http://localhost:8080/v1/checkout without a body
* POST http://localhost:8080/v1/logout without a body

The DB is already seeded with:

Products:
* `{id: 1, name: ‘Product 1’, price: 100, stock: 2}`
* `{id: 2, name: ‘Product 2’, price: 200, stock: 3}`

Users:
* `{id: 1, name: ‘Jon Doe’, password: ‘1234’, email: ‘jon.doe@company.com'}`
* `{id: 2, name: ‘Jon Doe 2’, password: ‘1234’, email: ‘jon.doe2@company.com'}`

Each request except login must send the cookie called `session` received in the previous response and the `Authorization` header
