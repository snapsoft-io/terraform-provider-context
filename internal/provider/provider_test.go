// Copyright SnapSoft 2026
// SPDX-License-Identifier: MIT

package provider

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

// testAccProtoV6ProviderFactories is used to instantiate a provider during acceptance testing.
// The factory function is called for each Terraform CLI command to create a provider
// server that the CLI can connect to and interact with.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"context": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccPreCheck skips the test unless TF_ACC is set, mirroring the behaviour of resource.Test
// but with an explicit, descriptive skip message.
func testAccPreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless TF_ACC is set")
	}
}

// writeTestFile writes content into a temporary directory owned by the test and returns the absolute path.
func writeTestFile(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write %s: %v", name, err)
	}
	return path
}

// expectId is a shorthand state check for the `id` attribute of a context_label data source.
func expectId(address, want string) statecheck.StateCheck {
	return statecheck.ExpectKnownValue(address, tfjsonpath.New("id"), knownvalue.StringExact(want))
}

// expectStringMap is a shorthand state check asserting a map attribute holds exactly the given entries.
func expectStringMap(address, attribute string, want map[string]string) statecheck.StateCheck {
	checks := make(map[string]knownvalue.Check, len(want))
	for k, v := range want {
		checks[k] = knownvalue.StringExact(v)
	}
	return statecheck.ExpectKnownValue(address, tfjsonpath.New(attribute), knownvalue.MapExact(checks))
}

// expectTags is a shorthand state check for the full `tags` map of a context_label data source.
func expectTags(address string, want map[string]string) statecheck.StateCheck {
	return expectStringMap(address, "tags", want)
}

// expectVars is a shorthand state check for the full `vars` map of a context_variables data source.
func expectVars(address string, want map[string]string) statecheck.StateCheck {
	return expectStringMap(address, "vars", want)
}

func TestAcc_Provider_Basic(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `provider "context" {}`,
			},
		},
	})
}

func TestAcc_Provider_AllAttributes(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "context" {
					  id_casing                   = "snake_case"
					  id_prefix                   = "acme"
					  include_resource_type_in_id = true
					  vars = {
					    organization = "acme"
					    team         = "platform"
					  }
					  mappers = [
					    {
					      name     = "noop"
					      function = ".context"
					    },
					    {
					      name          = "conditional noop"
					      run_condition = ".vars.team == \"platform\""
					      function      = ".context"
					    }
					  ]
					}

					data "context_namespace" "env" {
					  name    = "production"
					  context = { stack = [] }
					}

					data "context_label" "bucket" {
					  name          = "assets"
					  resource_type = "s3-bucket"
					  context       = data.context_namespace.env.context
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					expectId("data.context_label.bucket", "acme_production_assets_s3-bucket"),
				},
			},
		},
	})
}

func TestAcc_Provider_InvalidIdCasing(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "context" {
					  id_casing = "SCREAMING_SNAKE"
					}

					data "context_namespace" "env" {
					  name    = "production"
					  context = { stack = [] }
					}
				`,
				ExpectError: regexp.MustCompile(`(?s)Invalid Attribute Value Match.*id_casing.*SCREAMING_SNAKE`),
			},
		},
	})
}

func TestAcc_Provider_InvalidMapperJqSyntax(t *testing.T) {
	testAccPreCheck(t)

	t.Run("function", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: `
						provider "context" {
						  mappers = [
						    {
						      name     = "broken"
						      function = ".context | |"
						    }
						  ]
						}

						data "context_namespace" "env" {
						  name    = "production"
						  context = { stack = [] }
						}
					`,
					ExpectError: regexp.MustCompile(`(?s)Invalid jq syntax in the mapper function.*"broken"`),
				},
			},
		})
	})

	t.Run("run_condition", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: `
						provider "context" {
						  mappers = [
						    {
						      name          = "broken-cond"
						      run_condition = ".vars.x =="
						      function      = ".context"
						    }
						  ]
						}

						data "context_namespace" "env" {
						  name    = "production"
						  context = { stack = [] }
						}
					`,
					ExpectError: regexp.MustCompile(`(?s)Invalid jq syntax in the mapper run condition.*"broken-cond"`),
				},
			},
		})
	})
}

func TestAcc_Provider_MappersFilePath(t *testing.T) {
	testAccPreCheck(t)

	mappersFile := writeTestFile(t, "mappers.json", `[
		{
			"name": "Enrich tags with organisational metadata",
			"function": "(.context.outputs.tags += {Team: .vars.team, Organization: .vars.organization}).context"
		}
	]`)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "context" {
					  mappers_file_path = "` + mappersFile + `"
					  vars = {
					    organization = "acme"
					    team         = "platform"
					  }
					}

					data "context_namespace" "env" {
					  name    = "production"
					  context = { stack = [] }
					}

					data "context_label" "bucket" {
					  name          = "assets"
					  resource_type = "s3-bucket"
					  context       = data.context_namespace.env.context
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					expectId("data.context_label.bucket", "production-assets"),
					expectTags("data.context_label.bucket", map[string]string{
						"Team":         "platform",
						"Organization": "acme",
					}),
				},
			},
		},
	})
}

func TestAcc_Provider_MappersFilePath_Errors(t *testing.T) {
	testAccPreCheck(t)

	t.Run("missing file", func(t *testing.T) {
		missing := filepath.Join(t.TempDir(), "missing.json")
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: `
						provider "context" {
						  mappers_file_path = "` + missing + `"
						}

						data "context_namespace" "env" {
						  name    = "production"
						  context = { stack = [] }
						}
					`,
					ExpectError: regexp.MustCompile(`Failed to read the '.*missing\.json' mappers JSON file`),
				},
			},
		})
	})

	t.Run("invalid jq in file", func(t *testing.T) {
		broken := writeTestFile(t, "broken.json", `[{"name": "broken-file", "function": ".context | |"}]`)
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: `
						provider "context" {
						  mappers_file_path = "` + broken + `"
						}

						data "context_namespace" "env" {
						  name    = "production"
						  context = { stack = [] }
						}
					`,
					ExpectError: regexp.MustCompile(`(?s)Failed to read the .* mappers JSON file.*"broken-file" has an invalid jq syntax`),
				},
			},
		})
	})

	t.Run("malformed json", func(t *testing.T) {
		malformed := writeTestFile(t, "malformed.json", `{not json`)
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: `
						provider "context" {
						  mappers_file_path = "` + malformed + `"
						}

						data "context_namespace" "env" {
						  name    = "production"
						  context = { stack = [] }
						}
					`,
					ExpectError: regexp.MustCompile(`Failed to read the '.*malformed\.json' mappers JSON file`),
				},
			},
		})
	})
}

// TestAcc_Provider_InlineMappersTakePrecedenceOverFile verifies that when both `mappers` and
// `mappers_file_path` are given, only the inline mappers are used.
func TestAcc_Provider_InlineMappersTakePrecedenceOverFile(t *testing.T) {
	testAccPreCheck(t)

	mappersFile := writeTestFile(t, "mappers.json", `[
		{"name": "from-file", "function": "(.context.outputs += {id: \"from-file\"}).context"}
	]`)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "context" {
					  mappers_file_path = "` + mappersFile + `"
					  mappers = [
					    {
					      name     = "inline"
					      function = "(.context.outputs += {id: \"inline\"}).context"
					    }
					  ]
					}

					data "context_namespace" "env" {
					  name    = "production"
					  context = { stack = [] }
					}

					data "context_label" "bucket" {
					  name          = "assets"
					  resource_type = "s3-bucket"
					  context       = data.context_namespace.env.context
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					expectId("data.context_label.bucket", "inline"),
				},
			},
		},
	})
}
