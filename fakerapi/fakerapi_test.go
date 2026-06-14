package fakerapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tamnd/fakerapi-cli/fakerapi"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *fakerapi.Client {
	t.Helper()
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	cfg := fakerapi.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	return fakerapi.NewClient(cfg)
}

const personsResp = `{"status":"OK","code":200,"total":1,"data":[{
	"id":1,"firstname":"Alice","lastname":"Smith",
	"email":"alice@example.com","phone":"555-1234",
	"birthday":"1990-03-15","gender":"female","website":"alice.example.com"
}]}`

const productsResp = `{"status":"OK","code":200,"total":1,"data":[{
	"id":1,"name":"Widget","description":"A test widget","ean":"1234567890123","price":9.99
}]}`

const addressesResp = `{"status":"OK","code":200,"total":1,"data":[{
	"id":1,"street":"123 Main St","city":"Springfield","zipcode":"12345",
	"country":"United States","country_code":"US","latitude":39.78,"longitude":-89.65
}]}`

func TestPersons_userAgent(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		ua := r.Header.Get("User-Agent")
		if ua == "" {
			t.Error("request carried no User-Agent header")
		}
		if !strings.Contains(ua, "fakerapi-cli") {
			t.Errorf("User-Agent %q does not contain fakerapi-cli", ua)
		}
		_, _ = w.Write([]byte(personsResp))
	})
	_, err := c.Persons(context.Background(), 1, "")
	if err != nil {
		t.Fatal(err)
	}
}

func TestPersons_parseFirstnameAndEmail(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(personsResp))
	})
	persons, err := c.Persons(context.Background(), 1, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(persons) != 1 {
		t.Fatalf("got %d persons, want 1", len(persons))
	}
	if persons[0].Firstname != "Alice" {
		t.Errorf("Firstname = %q, want Alice", persons[0].Firstname)
	}
	if persons[0].Email != "alice@example.com" {
		t.Errorf("Email = %q, want alice@example.com", persons[0].Email)
	}
}

func TestProducts_parseNameAndPrice(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(productsResp))
	})
	products, err := c.Products(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(products) == 0 {
		t.Fatal("expected at least 1 product")
	}
	if products[0].Name != "Widget" {
		t.Errorf("Name = %q, want Widget", products[0].Name)
	}
	if products[0].Price != 9.99 {
		t.Errorf("Price = %f, want 9.99", products[0].Price)
	}
	if products[0].EAN != "1234567890123" {
		t.Errorf("EAN = %q, want 1234567890123", products[0].EAN)
	}
}

func TestAddresses_parseCityAndCountry(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(addressesResp))
	})
	addrs, err := c.Addresses(context.Background(), 1, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(addrs) == 0 {
		t.Fatal("expected at least 1 address")
	}
	if addrs[0].City != "Springfield" {
		t.Errorf("City = %q, want Springfield", addrs[0].City)
	}
	if addrs[0].Country != "United States" {
		t.Errorf("Country = %q, want United States", addrs[0].Country)
	}
}

func TestPersons_retry503(t *testing.T) {
	var hits int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte(personsResp))
	}))
	defer ts.Close()

	cfg := fakerapi.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	cfg.Retries = 5
	c := fakerapi.NewClient(cfg)

	persons, err := c.Persons(context.Background(), 1, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(persons) == 0 || persons[0].Firstname != "Alice" {
		t.Errorf("unexpected result after retry: %+v", persons)
	}
	if hits != 3 {
		t.Errorf("server hits = %d, want 3", hits)
	}
}

func TestPersons_countParamInURL(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		qty := r.URL.Query().Get("_quantity")
		if qty != "5" {
			t.Errorf("_quantity = %q, want 5", qty)
		}
		_, _ = w.Write([]byte(`{"status":"OK","code":200,"total":5,"data":[]}`))
	})
	_, err := c.Persons(context.Background(), 5, "")
	if err != nil {
		t.Fatal(err)
	}
}
