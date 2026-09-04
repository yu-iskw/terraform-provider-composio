package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestExampleItemResourceMetadata(t *testing.T) {
	res := NewExampleItemResource()

	var resp resource.MetadataResponse
	res.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "template"}, &resp)

	if resp.TypeName != "template_example_item" {
		t.Fatalf("expected template_example_item, got %q", resp.TypeName)
	}
}

func TestExampleItemID(t *testing.T) {
	got := exampleItemID("My Example Item")
	want := "example-item-my-example-item"

	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
