// Copyright SnapSoft 2026
// SPDX-License-Identifier: MIT

package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/echoprovider"
)

// testAccProtoV6ProviderFactories is used to instantiate a provider during acceptance testing.
// The factory function is called for each Terraform CLI command to create a provider
// server that the CLI can connect to and interact with.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"context": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccProtoV6ProviderFactoriesWithEcho includes the echo provider alongside the scaffolding provider.
// It allows for testing assertions on data returned by an ephemeral resource during Open.
// The echoprovider is used to arrange tests by echoing ephemeral data into the Terraform state.
// This lets the data be referenced in test assertions with state checks.
var testAccProtoV6ProviderFactoriesWithEcho = map[string]func() (tfprotov6.ProviderServer, error){
	"context": providerserver.NewProtocol6WithError(New("test")()),
	"echo":    echoprovider.NewProviderServer(),
}

func testAccPreCheck(t *testing.T) {
	// You can add code here to run prior to any test case execution, for example assertions
	// about the appropriate environment variables being set are common to see in a pre-check
	// function.
}

func TestAcc_Provider_Basic(t *testing.T) {
	// You must skip the test if TF_ACC is not set
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless TF_ACC is set")
	}

	resource.Test(t, resource.TestCase{
		// This tells the test harness how to get your provider
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,

		// PreCheck is where you would check for API keys from env vars.
		// For a dummy test, it can be empty.
		PreCheck: func() {
			// e.g., if os.Getenv("MY_API_KEY") == "" { t.Fatal("...") }
		},

		// Steps define the sequence of 'terraform apply' and 'check'
		Steps: []resource.TestStep{
			// This single step just runs a 'plan' and checks for errors.
			{
				// This is the HCL config to run.
				// It's the minimum possible: just the provider block.
				Config: `
					provider "context" {
					  # No configuration arguments needed
					}
				`,

				// We don't need a 'Check' block for this test.
				// The test will pass as long as the 'Config'
				// can be planned without any errors.
			},
		},
	})
}
