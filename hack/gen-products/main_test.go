package main

import (
	"go/format"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

// TestGenerateEmpty verifies that generate(nil) produces a valid Go file
// containing the empty registeredProducts stub.
func TestGenerateEmpty(t *testing.T) {
	src, err := generate(nil)
	if err != nil {
		t.Fatalf("generate(nil) error: %v", err)
	}

	s := string(src)

	if !strings.Contains(s, "func registeredProducts() []cli.Product") {
		t.Errorf("missing registeredProducts signature; got:\n%s", s)
	}
	if !strings.Contains(s, "return nil") {
		t.Errorf("expected 'return nil'; got:\n%s", s)
	}

	// Must be valid Go (gofmt-clean).
	if _, err := format.Source(src); err != nil {
		t.Errorf("output is not valid gofmt source: %v", err)
	}
	// Must parse as valid Go.
	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, "products.gen.go", src, 0); err != nil {
		t.Errorf("output does not parse as Go: %v", err)
	}
}

// TestGenerateWithProduct verifies that generate with one enabled product
// emits the correct import path and constructor call.
func TestGenerateWithProduct(t *testing.T) {
	products := []Product{
		{Name: "udb", Dir: "products/udb", Enabled: true},
	}

	src, err := generate(products)
	if err != nil {
		t.Fatalf("generate error: %v", err)
	}

	s := string(src)

	if !strings.Contains(s, `"github.com/ucloud/ucloud-cli/products/udb"`) {
		t.Errorf("missing import path for products/udb; got:\n%s", s)
	}
	if !strings.Contains(s, "udb.New()") {
		t.Errorf("missing udb.New() in return slice; got:\n%s", s)
	}
	if !strings.Contains(s, "func registeredProducts() []cli.Product") {
		t.Errorf("missing registeredProducts signature; got:\n%s", s)
	}

	// Must be valid Go (gofmt-clean).
	if _, err := format.Source(src); err != nil {
		t.Errorf("output is not valid gofmt source: %v", err)
	}
	// Must parse as valid Go.
	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, "products.gen.go", src, 0); err != nil {
		t.Errorf("output does not parse as Go: %v", err)
	}
}

// TestGenerateEmptySlice verifies that generate([]Product{}) behaves identically
// to generate(nil).
func TestGenerateEmptySlice(t *testing.T) {
	fromNil, err := generate(nil)
	if err != nil {
		t.Fatalf("generate(nil) error: %v", err)
	}
	fromEmpty, err := generate([]Product{})
	if err != nil {
		t.Fatalf("generate([]Product{}) error: %v", err)
	}
	if string(fromNil) != string(fromEmpty) {
		t.Errorf("generate(nil) and generate([]Product{}) differ:\nnull:\n%s\nempty:\n%s",
			fromNil, fromEmpty)
	}
}
