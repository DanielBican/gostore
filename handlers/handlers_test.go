package handlers_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DanielBican/gostore/handlers"
	"github.com/DanielBican/gostore/tests"
)

const (
	username  = "test.user1@company.com"
	password  = "1234"
	loginBody = `{"Email": "test.user1@company.com", "Password": "1234"}`
	cartBody  = `{"ProductName": "Product 1","Quantity": 1}`
)

func TestHandlers(t *testing.T) {
	db := tests.NewUnit(t)
	h := handlers.HandleGroup{db}

	// Test login
	sc := post(h, t, "/v1/login", loginBody, h.Login, nil)

	// Test add to cart
	sc = post(h, t, "/v1/cart", cartBody, h.AddToCart, sc)

	// Test checkout
	sc = post(h, t, "/v1/checkout", "", h.Checkout, sc)

	// Test logout
	sc = post(h, t, "/v1/logout", "", h.Logout, sc)
}

func post(h handlers.HandleGroup, t *testing.T, url string, body string, handler func(w http.ResponseWriter, r *http.Request), sc *http.Cookie) *http.Cookie {

	t.Logf("received post request to %s with body %s", url, body)

	r, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(body)))
	if err != nil {
		t.Fatal(err)
	}

	if sc != nil {
		r.AddCookie(sc)
	}
	r.SetBasicAuth(username, password)

	w := httptest.NewRecorder()
	handler(w, r)

	resp := w.Result()
	if status := resp.StatusCode; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	cookies := resp.Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == handlers.SessionName {
			sessionCookie = cookie
		}
	}
	if sessionCookie == nil {
		t.Fatal("handler returned no session cookie")
	}

	return sessionCookie
}
