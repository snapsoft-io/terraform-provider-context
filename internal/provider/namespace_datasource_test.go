// Copyright SnapSoft 2026
// SPDX-License-Identifier: MIT

package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

// stackPath returns the tfjsonpath to the context stack list of a data source.
func stackPath() tfjsonpath.Path {
	return tfjsonpath.New("context").AtMapKey("stack")
}

// stackElementPath returns the tfjsonpath to attribute `attr` of the i-th context stack element.
func stackElementPath(i int, attr string) tfjsonpath.Path {
	return stackPath().AtSliceIndex(i).AtMapKey(attr)
}

func TestAcc_NamespaceDataSource_Basic(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "context" {}

					data "context_namespace" "env" {
					  name    = "production"
					  context = { stack = [] }
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.context_namespace.env", tfjsonpath.New("name"), knownvalue.StringExact("production")),
					statecheck.ExpectKnownValue("data.context_namespace.env", stackPath(), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue("data.context_namespace.env", stackElementPath(0, "name"), knownvalue.StringExact("production")),
					statecheck.ExpectKnownValue("data.context_namespace.env", stackElementPath(0, "type"), knownvalue.StringExact("ns")),
					statecheck.ExpectKnownValue("data.context_namespace.env", stackElementPath(0, "vars"), knownvalue.Null()),
					statecheck.ExpectKnownValue("data.context_namespace.env", stackElementPath(0, "mappers"), knownvalue.Null()),
					statecheck.ExpectKnownValue("data.context_namespace.env", stackElementPath(0, "id_casing"), knownvalue.Null()),
					statecheck.ExpectKnownValue("data.context_namespace.env", stackElementPath(0, "id_prefix"), knownvalue.Null()),
					statecheck.ExpectKnownValue("data.context_namespace.env", stackElementPath(0, "include_resource_type_in_id"), knownvalue.Null()),
					// Optional inputs must stay null when not configured.
					statecheck.ExpectKnownValue("data.context_namespace.env", tfjsonpath.New("vars"), knownvalue.Null()),
					statecheck.ExpectKnownValue("data.context_namespace.env", tfjsonpath.New("mappers"), knownvalue.Null()),
					statecheck.ExpectKnownValue("data.context_namespace.env", tfjsonpath.New("id_casing"), knownvalue.Null()),
					statecheck.ExpectKnownValue("data.context_namespace.env", tfjsonpath.New("id_prefix"), knownvalue.Null()),
					statecheck.ExpectKnownValue("data.context_namespace.env", tfjsonpath.New("include_resource_type_in_id"), knownvalue.Null()),
				},
			},
		},
	})
}

func TestAcc_NamespaceDataSource_AllAttributesArePropagatedToStack(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "context" {}

					data "context_namespace" "env" {
					  name                        = "production"
					  context                     = { stack = [] }
					  id_casing                   = "snake_case"
					  id_prefix                   = "acme"
					  include_resource_type_in_id = true
					  vars = {
					    environment = "production"
					    owner       = "platform"
					  }
					  mappers = [
					    {
					      name     = "Tag with environment"
					      function = "(.context.outputs.tags += {Environment: .vars.environment}).context"
					    },
					    {
					      name          = "Conditional"
					      run_condition = ".vars.environment != \"production\""
					      function      = ".context"
					    }
					  ]
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.context_namespace.env", stackPath(), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue("data.context_namespace.env", stackElementPath(0, "name"), knownvalue.StringExact("production")),
					statecheck.ExpectKnownValue("data.context_namespace.env", stackElementPath(0, "type"), knownvalue.StringExact("ns")),
					statecheck.ExpectKnownValue("data.context_namespace.env", stackElementPath(0, "id_casing"), knownvalue.StringExact("snake_case")),
					statecheck.ExpectKnownValue("data.context_namespace.env", stackElementPath(0, "id_prefix"), knownvalue.StringExact("acme")),
					statecheck.ExpectKnownValue("data.context_namespace.env", stackElementPath(0, "include_resource_type_in_id"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("data.context_namespace.env", stackElementPath(0, "vars"), knownvalue.MapExact(map[string]knownvalue.Check{
						"environment": knownvalue.StringExact("production"),
						"owner":       knownvalue.StringExact("platform"),
					})),
					statecheck.ExpectKnownValue("data.context_namespace.env", stackElementPath(0, "mappers"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"name":          knownvalue.StringExact("Tag with environment"),
							"run_condition": knownvalue.Null(),
							"function":      knownvalue.StringExact("(.context.outputs.tags += {Environment: .vars.environment}).context"),
						}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"name":          knownvalue.StringExact("Conditional"),
							"run_condition": knownvalue.StringExact(`.vars.environment != "production"`),
							"function":      knownvalue.StringExact(".context"),
						}),
					})),
				},
			},
		},
	})
}

func TestAcc_NamespaceDataSource_Chaining(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "context" {}

					data "context_namespace" "env" {
					  name    = "production"
					  context = { stack = [] }
					  vars    = { environment = "production" }
					}

					data "context_namespace" "region" {
					  name    = "eu-west-1"
					  context = data.context_namespace.env.context
					  vars    = { region = "eu-west-1" }
					}

					data "context_namespace" "component" {
					  name      = "#aws-vpc"
					  context   = data.context_namespace.region.context
					  id_prefix = "acme"
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					// Each namespace appends exactly one element; parent elements are preserved verbatim.
					statecheck.ExpectKnownValue("data.context_namespace.env", stackPath(), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue("data.context_namespace.region", stackPath(), knownvalue.ListSizeExact(2)),
					statecheck.ExpectKnownValue("data.context_namespace.component", stackPath(), knownvalue.ListSizeExact(3)),

					statecheck.ExpectKnownValue("data.context_namespace.component", stackElementPath(0, "name"), knownvalue.StringExact("production")),
					statecheck.ExpectKnownValue("data.context_namespace.component", stackElementPath(0, "vars"), knownvalue.MapExact(map[string]knownvalue.Check{
						"environment": knownvalue.StringExact("production"),
					})),
					statecheck.ExpectKnownValue("data.context_namespace.component", stackElementPath(1, "name"), knownvalue.StringExact("eu-west-1")),
					statecheck.ExpectKnownValue("data.context_namespace.component", stackElementPath(1, "vars"), knownvalue.MapExact(map[string]knownvalue.Check{
						"region": knownvalue.StringExact("eu-west-1"),
					})),
					statecheck.ExpectKnownValue("data.context_namespace.component", stackElementPath(2, "name"), knownvalue.StringExact("#aws-vpc")),
					statecheck.ExpectKnownValue("data.context_namespace.component", stackElementPath(2, "id_prefix"), knownvalue.StringExact("acme")),
					statecheck.ExpectKnownValue("data.context_namespace.component", stackElementPath(2, "type"), knownvalue.StringExact("ns")),
				},
			},
		},
	})
}

func TestAcc_NamespaceDataSource_EmptyNameIsAllowed(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "context" {}

					data "context_namespace" "root" {
					  name    = "production"
					  context = { stack = [] }
					}

					data "context_namespace" "anonymous" {
					  name    = ""
					  context = data.context_namespace.root.context
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.context_namespace.anonymous", stackPath(), knownvalue.ListSizeExact(2)),
					statecheck.ExpectKnownValue("data.context_namespace.anonymous", stackElementPath(1, "name"), knownvalue.StringExact("")),
				},
			},
		},
	})
}

func TestAcc_NamespaceDataSource_InvalidIdCasing(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "context" {}

					data "context_namespace" "env" {
					  name      = "production"
					  context   = { stack = [] }
					  id_casing = "Title Case"
					}
				`,
				ExpectError: regexp.MustCompile(`(?s)Invalid Attribute Value Match.*id_casing`),
			},
		},
	})
}

func TestAcc_NamespaceDataSource_MissingRequiredAttributes(t *testing.T) {
	testAccPreCheck(t)

	t.Run("name", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: `
						provider "context" {}

						data "context_namespace" "env" {
						  context = { stack = [] }
						}
					`,
					ExpectError: regexp.MustCompile(`(?s)Missing required argument.*"name" is required`),
				},
			},
		})
	})

	t.Run("context", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: `
						provider "context" {}

						data "context_namespace" "env" {
						  name = "production"
						}
					`,
					ExpectError: regexp.MustCompile(`(?s)Missing required argument.*"context" is required`),
				},
			},
		})
	})
}

func TestAcc_NamespaceDataSource_InvalidStackOrder(t *testing.T) {
	testAccPreCheck(t)

	t.Run("namespace cannot follow a label", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: `
						provider "context" {}

						data "context_namespace" "env" {
						  name = "after-label"
						  context = {
						    stack = [
						      { name = "production", type = "ns" },
						      { name = "bucket", type = "lb" },
						    ]
						  }
						}
					`,
					ExpectError: regexp.MustCompile(`Invalid context label order`),
				},
			},
		})
	})

	// An unknown type also breaks the stack order check, which is the error surfaced by the CLI here.
	// The dedicated per-element type validator is covered by unit tests in internal/context/ctxvalidator.
	t.Run("invalid stack element type", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: `
						provider "context" {}

						data "context_namespace" "env" {
						  name = "x"
						  context = {
						    stack = [
						      { name = "production", type = "namespace" },
						    ]
						  }
						}
					`,
					ExpectError: regexp.MustCompile(`Invalid context label order`),
				},
			},
		})
	})
}

// TestAcc_NamespaceDataSource_ManuallyBuiltStack verifies that a stack passed as a literal object
// (as opposed to one produced by another data source) is accepted and extended.
func TestAcc_NamespaceDataSource_ManuallyBuiltStack(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "context" {}

					data "context_namespace" "child" {
					  name = "child"
					  context = {
					    stack = [
					      {
					        name      = "root"
					        type      = "ns"
					        vars      = { a = "1" }
					        id_casing = "PascalCase"
					        mappers = [
					          { name = "m", function = ".context" }
					        ]
					      },
					    ]
					  }
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.context_namespace.child", stackPath(), knownvalue.ListSizeExact(2)),
					statecheck.ExpectKnownValue("data.context_namespace.child", stackElementPath(0, "name"), knownvalue.StringExact("root")),
					statecheck.ExpectKnownValue("data.context_namespace.child", stackElementPath(0, "id_casing"), knownvalue.StringExact("PascalCase")),
					statecheck.ExpectKnownValue("data.context_namespace.child", stackElementPath(0, "mappers"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue("data.context_namespace.child", stackElementPath(1, "name"), knownvalue.StringExact("child")),
				},
			},
		},
	})
}
