// Package fakerapi is the library behind the fakerapi command line:
// the HTTP client, request shaping, and the typed data models for the
// fakerapi.it API (fake test data generator — persons, addresses, products,
// companies, texts, and images — no key required).
//
// The Client is the spine every command shares. It sets a real User-Agent,
// paces requests so a busy session stays polite, and retries the transient
// failures (429 and 5xx) that any public API throws under load.
package fakerapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// Host is the API hostname this package talks to.
const Host = "fakerapi.it"

// Config holds all tuneable parameters for a Client.
type Config struct {
	BaseURL   string
	UserAgent string
	Rate      time.Duration
	Timeout   time.Duration
	Retries   int
}

// DefaultConfig returns the production configuration for fakerapi.it.
func DefaultConfig() Config {
	return Config{
		BaseURL:   "https://fakerapi.it/api/v2",
		UserAgent: "fakerapi-cli/0.1.0 (github.com/tamnd/fakerapi-cli)",
		Rate:      200 * time.Millisecond,
		Timeout:   30 * time.Second,
		Retries:   3,
	}
}

// Person is a generated fake person.
type Person struct {
	ID        int    `json:"id"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Birthday  string `json:"birthday"`
	Gender    string `json:"gender"`
	Website   string `json:"website"`
}

// Address is a generated fake address.
type Address struct {
	ID          int     `json:"id"`
	Street      string  `json:"street"`
	City        string  `json:"city"`
	Zipcode     string  `json:"zipcode"`
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
}

// Product is a generated fake product.
type Product struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	EAN         string  `json:"ean"`
	Price       float64 `json:"price"`
}

// Company is a generated fake company.
type Company struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Country     string `json:"country"`
	Website     string `json:"website"`
	FoundedYear int    `json:"founded_year"`
}

// Text is a generated fake text article.
type Text struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Author  string `json:"author"`
	Genre   string `json:"genre"`
	Content string `json:"content"`
}

// FakeImage is generated fake image metadata.
type FakeImage struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
}

// Client talks to fakerapi.it over HTTP.
type Client struct {
	cfg  Config
	http *http.Client
	mu   sync.Mutex
	last time.Time
}

// NewClient returns a Client configured from cfg.
func NewClient(cfg Config) *Client {
	return &Client{cfg: cfg, http: &http.Client{Timeout: cfg.Timeout}}
}

// apiResponse is the generic envelope all fakerapi.it responses share.
type apiResponse[T any] struct {
	Status string `json:"status"`
	Code   int    `json:"code"`
	Total  int    `json:"total"`
	Data   []T    `json:"data"`
}

// Persons returns a list of fake persons.
func (c *Client) Persons(ctx context.Context, count int, locale string) ([]Person, error) {
	params := url.Values{}
	params.Set("_quantity", fmt.Sprintf("%d", count))
	if locale != "" {
		params.Set("_locale", locale)
	}
	body, err := c.get(ctx, c.cfg.BaseURL+"/persons?"+params.Encode())
	if err != nil {
		return nil, err
	}
	var r apiResponse[Person]
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("fakerapi: decode persons: %w", err)
	}
	return r.Data, nil
}

// Addresses returns a list of fake addresses.
func (c *Client) Addresses(ctx context.Context, count int, countryCode string) ([]Address, error) {
	params := url.Values{}
	params.Set("_quantity", fmt.Sprintf("%d", count))
	if countryCode != "" {
		params.Set("_country_code", countryCode)
	}
	body, err := c.get(ctx, c.cfg.BaseURL+"/addresses?"+params.Encode())
	if err != nil {
		return nil, err
	}
	var r apiResponse[Address]
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("fakerapi: decode addresses: %w", err)
	}
	return r.Data, nil
}

// Products returns a list of fake products.
func (c *Client) Products(ctx context.Context, count int) ([]Product, error) {
	params := url.Values{}
	params.Set("_quantity", fmt.Sprintf("%d", count))
	body, err := c.get(ctx, c.cfg.BaseURL+"/products?"+params.Encode())
	if err != nil {
		return nil, err
	}
	var r apiResponse[Product]
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("fakerapi: decode products: %w", err)
	}
	return r.Data, nil
}

// Companies returns a list of fake companies.
func (c *Client) Companies(ctx context.Context, count int) ([]Company, error) {
	params := url.Values{}
	params.Set("_quantity", fmt.Sprintf("%d", count))
	body, err := c.get(ctx, c.cfg.BaseURL+"/companies?"+params.Encode())
	if err != nil {
		return nil, err
	}
	var r apiResponse[Company]
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("fakerapi: decode companies: %w", err)
	}
	return r.Data, nil
}

// Texts returns a list of fake text articles.
func (c *Client) Texts(ctx context.Context, count, chars int) ([]Text, error) {
	params := url.Values{}
	params.Set("_quantity", fmt.Sprintf("%d", count))
	if chars > 0 {
		params.Set("_characters", fmt.Sprintf("%d", chars))
	}
	body, err := c.get(ctx, c.cfg.BaseURL+"/texts?"+params.Encode())
	if err != nil {
		return nil, err
	}
	var r apiResponse[Text]
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("fakerapi: decode texts: %w", err)
	}
	return r.Data, nil
}

// Images returns a list of fake image records.
func (c *Client) Images(ctx context.Context, count int, imgType string, width, height int) ([]FakeImage, error) {
	params := url.Values{}
	params.Set("_quantity", fmt.Sprintf("%d", count))
	if imgType != "" {
		params.Set("_type", imgType)
	}
	if width > 0 {
		params.Set("_width", fmt.Sprintf("%d", width))
	}
	if height > 0 {
		params.Set("_height", fmt.Sprintf("%d", height))
	}
	body, err := c.get(ctx, c.cfg.BaseURL+"/images?"+params.Encode())
	if err != nil {
		return nil, err
	}
	var r apiResponse[FakeImage]
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("fakerapi: decode images: %w", err)
	}
	return r.Data, nil
}

func (c *Client) get(ctx context.Context, u string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		body, retry, err := c.do(ctx, u)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("fakerapi: get %s: %w", u, lastErr)
}

func (c *Client) do(ctx context.Context, u string) (body []byte, retry bool, err error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, true, err
	}
	return b, false, nil
}

// pace blocks until at least Rate has passed since the previous request.
func (c *Client) pace() {
	if c.cfg.Rate <= 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if wait := c.cfg.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		d = 5 * time.Second
	}
	return d
}
