package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestExampleItemDataSourceMetadata(t *testing.T) {
	ds := NewExampleItemDataSource()

	var resp datasource.MetadataResponse
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "template"}, &resp)

	if resp.TypeName != "template_example_item" {
		t.Fatalf("expected template_example_item, got %q", resp.TypeName)
	}
}
