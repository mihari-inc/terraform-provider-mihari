package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func TestProviderSchema(t *testing.T) {
	// Verify the provider schema compiles and is valid
	providerFactory := providerserver.NewProtocol6WithError(New("test")())
	_, err := providerFactory()
	if err != nil {
		t.Fatalf("provider factory error: %v", err)
	}
}

// testAccProtoV6ProviderFactories are used for acceptance tests.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"mihari": providerserver.NewProtocol6WithError(New("test")()),
}
