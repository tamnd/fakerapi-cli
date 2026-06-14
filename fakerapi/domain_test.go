package fakerapi

import (
	"testing"
)

// These tests are offline: they exercise the URI driver's pure string functions.
// The client's HTTP behaviour is covered in fakerapi_test.go.

func TestDomainInfo(t *testing.T) {
	info := Domain{}.Info()
	if info.Scheme != "fakerapi" {
		t.Errorf("Scheme = %q, want fakerapi", info.Scheme)
	}
	if len(info.Hosts) == 0 || info.Hosts[0] != Host {
		t.Errorf("Hosts = %v, want [%s]", info.Hosts, Host)
	}
	if info.Identity.Binary != "fakerapi" {
		t.Errorf("Identity.Binary = %q, want fakerapi", info.Identity.Binary)
	}
}

func TestClassify_resource(t *testing.T) {
	typ, id, err := Domain{}.Classify("persons")
	if err != nil {
		t.Fatalf("Classify: %v", err)
	}
	if typ != "resource" {
		t.Errorf("type = %q, want resource", typ)
	}
	if id != "persons" {
		t.Errorf("id = %q, want persons", id)
	}
}

func TestClassify_empty(t *testing.T) {
	_, _, err := Domain{}.Classify("")
	if err == nil {
		t.Error("expected error on empty input, got nil")
	}
}

func TestLocate_person(t *testing.T) {
	got, err := Domain{}.Locate("person", "1")
	if err != nil {
		t.Fatalf("Locate person: %v", err)
	}
	if got == "" {
		t.Error("Locate returned empty URL")
	}
}

func TestLocate_product(t *testing.T) {
	got, err := Domain{}.Locate("product", "1")
	if err != nil {
		t.Fatalf("Locate product: %v", err)
	}
	if got == "" {
		t.Error("Locate returned empty URL")
	}
}

func TestLocate_badType(t *testing.T) {
	_, err := Domain{}.Locate("page", "foo")
	if err == nil {
		t.Error("expected error for unknown type, got nil")
	}
}
