package handlers

import (
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/DanielBican/gostore/data"
	"github.com/DanielBican/gostore/database"
	"github.com/hashicorp/go-memdb"

	"github.com/gorilla/sessions"

	"golang.org/x/crypto/bcrypt"
)

// HandleGroup keeps a pointer to the DB connexion
type HandleGroup struct {
	Db *memdb.MemDB
}

// A better way to store the key than the source code would be an environmental variable
var authKey = "a-very-secure-key"
var store *sessions.CookieStore

// SessionName represents the key for user session
var SessionName = "session"

func init() {

	store = sessions.NewCookieStore([]byte(authKey))

	store.Options = &sessions.Options{
		MaxAge:   3600,
		HttpOnly: true,
	}

	// Register data structure for storing cart elements
	gob.Register(map[string]int{})
}

// Login authenticates user based on email and password
func (h HandleGroup) Login(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		// Decode body
		var login data.Login
		err := decode(r, &login)
		if err != nil {
			handleError(w, err, http.StatusBadRequest)
			return
		}

		// Authenticate user
		u, err := auth(h.Db, login.Email, login.Password)
		if err != nil {
			handleError(w, err, http.StatusUnauthorized)
			return
		}

		// Create a new session
		session, err := store.Get(r, SessionName)
		if err != nil {
			handleError(w, err, http.StatusInternalServerError)
			return
		}

		// Set user authenticated and save session
		session.Values["username"] = u.Name
		session.Values["cart"] = map[string]int{}
		err = sessions.Save(r, w)
		if err != nil {
			handleError(w, err, http.StatusInternalServerError)
			return
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// Logout terminates user session
func (h HandleGroup) Logout(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		// Authenticate user
		_, err := basicAuth(h.Db, r)
		if err != nil {
			w.Header().Set("WWW-Authenticate", "Basic realm=\"Access to the gostore\", charset=\"UTF-8\"")
			handleError(w, err, http.StatusUnauthorized)
			return
		}

		// Get session
		session, err := store.Get(r, SessionName)
		if err != nil {
			handleError(w, err, http.StatusInternalServerError)
			return
		}

		// Check if previously authenticated
		if username, ok := session.Values["username"]; !ok || username == "" {
			w.Write([]byte("not logged in"))
		} else {
			deleteSession(w, r, session)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// AddToCart adds a product to the cart
func (h HandleGroup) AddToCart(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		// Authenticate user
		_, err := basicAuth(h.Db, r)
		if err != nil {
			w.Header().Set("WWW-Authenticate", "Basic realm=\"Access to the gostore\", charset=\"UTF-8\"")
			handleError(w, err, http.StatusUnauthorized)
			return
		}

		// Get user session
		session, err := store.Get(r, SessionName)
		if err != nil {
			handleError(w, err, http.StatusInternalServerError)
			return
		}

		// User authenticated?
		username, ok := session.Values["username"]
		if !ok || username == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			deleteSession(w, r, session)
			return
		}

		// Decode body
		var cartItem data.CartItem
		err = decode(r, &cartItem)
		if err != nil {
			handleError(w, err, http.StatusBadRequest)
			return
		}

		// Check that product is in stock
		err = database.VerifyProductAndStock(h.Db, cartItem)
		if err != nil {
			handleError(w, fmt.Errorf("%s %s", username, err.Error()), http.StatusNotFound)
			return
		}

		// Add product to cart
		log.Printf("%s add to cart %+v\n", username, cartItem)
		cart := session.Values["cart"].(map[string]int)
		cart[cartItem.ProductName] += cartItem.Quantity
		err = sessions.Save(r, w)
		if err != nil {
			handleError(w, err, http.StatusInternalServerError)
			return
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// Checkout finalizes cart products purchase
func (h HandleGroup) Checkout(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		// Authenticate user
		_, err := basicAuth(h.Db, r)
		if err != nil {
			w.Header().Set("WWW-Authenticate", "Basic realm=\"Access to the gostore\", charset=\"UTF-8\"")
			handleError(w, err, http.StatusUnauthorized)
			return
		}

		// Get user session
		session, err := store.Get(r, SessionName)
		if err != nil {
			handleError(w, err, http.StatusInternalServerError)
			return
		}

		// User authenticated?
		username, ok := session.Values["username"]
		if !ok || username == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			deleteSession(w, r, session)
			return
		}

		// Check if cart has items
		cart := session.Values["cart"].(map[string]int)
		if len(cart) == 0 {
			w.Write([]byte("cart is empty"))
			return
		}

		// Checkout products
		c := make(chan *database.CheckoutResult)
		go func(c chan *database.CheckoutResult) {
			c <- database.Checkout(h.Db, cart)
		}(c)

		checkoutResult := <-c
		if checkoutResult.Result {
			// Clear cart
			log.Printf("%s clear cart %v\n", username, cart)
			session.Values["cart"] = map[string]int{}
			err = sessions.Save(r, w)
			if err != nil {
				handleError(w, err, http.StatusInternalServerError)
				return
			}
		} else {
			log.Printf("%s checkout result %+v\n", username, checkoutResult)
			if checkoutResult.Error != nil {
				handleError(w, checkoutResult.Error, http.StatusNotFound)
				return
			}
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// Decode takes the request body of a HTTP call and decodes using the interface provided as argument
// TODO move this to utils package
func decode(r *http.Request, val interface{}) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(val); err != nil {
		log.Printf("decode error %v", err)
		return err
	}
	return nil
}

func auth(db *memdb.MemDB, email, pass string) (*data.User, error) {

	// Check if user exists
	u, err := database.GetUserByEmail(db, email)
	if err != nil {
		return nil, err
	}

	// Check if passwords match
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(pass))
	if err != nil {
		return nil, err
	}

	return u, nil
}

func basicAuth(db *memdb.MemDB, r *http.Request) (*data.User, error) {

	// Check Authorization header
	email, _, ok := r.BasicAuth()
	if !ok {
		return nil, errors.New("unauthorized")
	}

	// Get user
	return database.GetUserByEmail(db, email)
}

func deleteSession(w http.ResponseWriter, r *http.Request, session *sessions.Session) error {
	session.Options.MaxAge = -1
	err := sessions.Save(r, w)
	if err != nil {
		handleError(w, err, http.StatusUnauthorized)
		return err
	}
	return nil
}

func handleError(w http.ResponseWriter, err error, statusCode int) {
	log.Println(err)
	http.Error(w, err.Error(), statusCode)
}
