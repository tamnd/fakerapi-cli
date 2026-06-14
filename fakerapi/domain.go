package fakerapi

import (
	"context"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/any-cli/kit/errs"
)

func init() { kit.Register(Domain{}) }

// Domain is the fakerapi driver.
type Domain struct{}

// Info describes the scheme, the hostnames a pasted link is matched against, and
// the identity reused for the binary's help and version.
func (Domain) Info() kit.DomainInfo {
	return kit.DomainInfo{
		Scheme: "fakerapi",
		Hosts:  []string{Host},
		Identity: kit.Identity{
			Binary: "fakerapi",
			Short:  "Generate fake test data: persons, addresses, products, companies, texts, and images.",
			Long: `fakerapi calls the free fakerapi.it API to generate realistic fake data for
testing: persons, mailing addresses, products, companies, text articles, and
image metadata. No API key required. All commands emit clean JSON records.`,
			Site: Host,
			Repo: "https://github.com/tamnd/fakerapi-cli",
		},
	}
}

// Register installs the client factory and every operation onto app.
func (Domain) Register(app *kit.App) {
	app.SetClient(newClient)

	kit.Handle(app, kit.OpMeta{
		Name:    "persons",
		Group:   "read",
		List:    true,
		Summary: "Generate fake persons",
		URIType: "person",
	}, listPersons)

	kit.Handle(app, kit.OpMeta{
		Name:    "addresses",
		Group:   "read",
		List:    true,
		Summary: "Generate fake addresses",
		URIType: "address",
	}, listAddresses)

	kit.Handle(app, kit.OpMeta{
		Name:    "products",
		Group:   "read",
		List:    true,
		Summary: "Generate fake products",
		URIType: "product",
	}, listProducts)

	kit.Handle(app, kit.OpMeta{
		Name:    "companies",
		Group:   "read",
		List:    true,
		Summary: "Generate fake companies",
		URIType: "company",
	}, listCompanies)

	kit.Handle(app, kit.OpMeta{
		Name:    "texts",
		Group:   "read",
		List:    true,
		Summary: "Generate fake text articles",
		URIType: "text",
	}, listTexts)

	kit.Handle(app, kit.OpMeta{
		Name:    "images",
		Group:   "read",
		List:    true,
		Summary: "Generate fake image metadata",
		URIType: "image",
	}, listImages)
}

// newClient builds the client from the host-resolved config.
func newClient(_ context.Context, cfg kit.Config) (any, error) {
	c := DefaultConfig()
	if cfg.UserAgent != "" {
		c.UserAgent = cfg.UserAgent
	}
	if cfg.Rate > 0 {
		c.Rate = cfg.Rate
	}
	if cfg.Retries > 0 {
		c.Retries = cfg.Retries
	}
	if cfg.Timeout > 0 {
		c.Timeout = cfg.Timeout
	}
	return NewClient(c), nil
}

// --- inputs ---

type personsInput struct {
	Count  int     `kit:"flag" help:"number of persons to generate (default 10)"`
	Locale string  `kit:"flag" help:"locale code (e.g. en_US, fr_FR)"`
	Client *Client `kit:"inject"`
}

type addressesInput struct {
	Count   int     `kit:"flag" help:"number of addresses to generate (default 10)"`
	Country string  `kit:"flag" help:"ISO country code (e.g. US)"`
	Client  *Client `kit:"inject"`
}

type productsInput struct {
	Count  int     `kit:"flag" help:"number of products to generate (default 10)"`
	Client *Client `kit:"inject"`
}

type companiesInput struct {
	Count  int     `kit:"flag" help:"number of companies to generate (default 10)"`
	Client *Client `kit:"inject"`
}

type textsInput struct {
	Count  int     `kit:"flag" help:"number of texts to generate (default 10)"`
	Chars  int     `kit:"flag" help:"characters per text (default 500)"`
	Client *Client `kit:"inject"`
}

type imagesInput struct {
	Count  int     `kit:"flag" help:"number of images to generate (default 10)"`
	Type   string  `kit:"flag" help:"image type: animals, cats, dogs, city, nature"`
	Width  int     `kit:"flag" help:"image width in pixels"`
	Height int     `kit:"flag" help:"image height in pixels"`
	Client *Client `kit:"inject"`
}

// --- handlers ---

func listPersons(ctx context.Context, in personsInput, emit func(*Person) error) error {
	count := in.Count
	if count <= 0 {
		count = 10
	}
	items, err := in.Client.Persons(ctx, count, in.Locale)
	if err != nil {
		return mapErr(err)
	}
	for i := range items {
		if err := emit(&items[i]); err != nil {
			return err
		}
	}
	return nil
}

func listAddresses(ctx context.Context, in addressesInput, emit func(*Address) error) error {
	count := in.Count
	if count <= 0 {
		count = 10
	}
	items, err := in.Client.Addresses(ctx, count, in.Country)
	if err != nil {
		return mapErr(err)
	}
	for i := range items {
		if err := emit(&items[i]); err != nil {
			return err
		}
	}
	return nil
}

func listProducts(ctx context.Context, in productsInput, emit func(*Product) error) error {
	count := in.Count
	if count <= 0 {
		count = 10
	}
	items, err := in.Client.Products(ctx, count)
	if err != nil {
		return mapErr(err)
	}
	for i := range items {
		if err := emit(&items[i]); err != nil {
			return err
		}
	}
	return nil
}

func listCompanies(ctx context.Context, in companiesInput, emit func(*Company) error) error {
	count := in.Count
	if count <= 0 {
		count = 10
	}
	items, err := in.Client.Companies(ctx, count)
	if err != nil {
		return mapErr(err)
	}
	for i := range items {
		if err := emit(&items[i]); err != nil {
			return err
		}
	}
	return nil
}

func listTexts(ctx context.Context, in textsInput, emit func(*Text) error) error {
	count := in.Count
	if count <= 0 {
		count = 10
	}
	items, err := in.Client.Texts(ctx, count, in.Chars)
	if err != nil {
		return mapErr(err)
	}
	for i := range items {
		if err := emit(&items[i]); err != nil {
			return err
		}
	}
	return nil
}

func listImages(ctx context.Context, in imagesInput, emit func(*FakeImage) error) error {
	count := in.Count
	if count <= 0 {
		count = 10
	}
	items, err := in.Client.Images(ctx, count, in.Type, in.Width, in.Height)
	if err != nil {
		return mapErr(err)
	}
	for i := range items {
		if err := emit(&items[i]); err != nil {
			return err
		}
	}
	return nil
}

// --- Resolver ---

// Classify turns a resource type name into the canonical (type, id).
func (Domain) Classify(input string) (uriType, id string, err error) {
	if input == "" {
		return "", "", errs.Usage("unrecognized fakerapi reference: %q", input)
	}
	return "resource", input, nil
}

// Locate returns the live https URL for a (type, id).
func (Domain) Locate(uriType, id string) (string, error) {
	paths := map[string]string{
		"person":  "persons",
		"address": "addresses",
		"product": "products",
		"company": "companies",
		"text":    "texts",
		"image":   "images",
	}
	if p, ok := paths[uriType]; ok {
		return "https://" + Host + "/api/v2/" + p, nil
	}
	return "", errs.Usage("fakerapi has no resource type %q", uriType)
}

func mapErr(err error) error {
	return err
}
