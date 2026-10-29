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

func TestAcc_VariablesDataSource_Basic(t *testing.T) {
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
					  vars = {
					    environment = "production"
					    region      = "eu-west-1"
					  }
					}

					data "context_variables" "this" {
					  context = data.context_namespace.env.context
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					expectVars("data.context_variables.this", map[string]string{
						"environment": "production",
						"region":      "eu-west-1",
					}),
					// The input context is echoed back unchanged (one namespace element, no label appended).
					statecheck.ExpectKnownValue("data.context_variables.this", stackPath(), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue("data.context_variables.this", stackElementPath(0, "name"), knownvalue.StringExact("production")),
				},
			},
		},
	})
}

func TestAcc_VariablesDataSource_NoVarsAnywhere(t *testing.T) {
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

					data "context_variables" "this" {
					  context = data.context_namespace.env.context
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.context_variables.this", tfjsonpath.New("vars"), knownvalue.MapSizeExact(0)),
				},
			},
		},
	})
}

func TestAcc_VariablesDataSource_ProviderVarsOnly(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "context" {
					  vars = {
					    organization = "acme"
					    team         = "platform"
					  }
					}

					data "context_namespace" "env" {
					  name    = "production"
					  context = { stack = [] }
					}

					data "context_variables" "this" {
					  context = data.context_namespace.env.context
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					expectVars("data.context_variables.this", map[string]string{
						"organization": "acme",
						"team":         "platform",
					}),
				},
			},
		},
	})
}

// TestAcc_VariablesDataSource_Precedence mirrors examples/data-sources/context_variables: provider vars are
// the lowest priority and deeper namespaces override shallower ones on key conflicts.
func TestAcc_VariablesDataSource_Precedence(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "context" {
					  vars = {
					    organization = "acme"
					    team         = "platform"
					    region       = "provider-region"
					  }
					}

					data "context_namespace" "env" {
					  name    = "production"
					  context = { stack = [] }
					  vars = {
					    environment = "production"
					    region      = "eu-west-1"
					  }
					}

					data "context_namespace" "region" {
					  name    = "eu-west-1"
					  context = data.context_namespace.env.context
					  vars = {
					    region = "eu-west-1-overridden"
					  }
					}

					data "context_variables" "env" {
					  context = data.context_namespace.env.context
					}

					data "context_variables" "region" {
					  context = data.context_namespace.region.context
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					expectVars("data.context_variables.env", map[string]string{
						"organization": "acme",
						"team":         "platform",
						"environment":  "production",
						"region":       "eu-west-1",
					}),
					expectVars("data.context_variables.region", map[string]string{
						"organization": "acme",
						"team":         "platform",
						"environment":  "production",
						"region":       "eu-west-1-overridden",
					}),
				},
			},
		},
	})
}

func TestAcc_VariablesDataSource_EmptyStringValuesOverride(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "context" {
					  vars = { prefix = "acme" }
					}

					data "context_namespace" "env" {
					  name    = "production"
					  context = { stack = [] }
					  vars    = { prefix = "" }
					}

					data "context_variables" "this" {
					  context = data.context_namespace.env.context
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					expectVars("data.context_variables.this", map[string]string{"prefix": ""}),
				},
			},
		},
	})
}

// TestAcc_VariablesDataSource_ManuallyBuiltStack verifies vars are read from a literal stack, including
// elements that declare no vars at all.
func TestAcc_VariablesDataSource_ManuallyBuiltStack(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "context" {}

					data "context_variables" "this" {
					  context = {
					    stack = [
					      { name = "root", type = "ns", vars = { a = "1", b = "1" } },
					      { name = "middle", type = "ns" },
					      { name = "leaf", type = "ns", vars = { b = "2", c = "2" } },
					    ]
					  }
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					expectVars("data.context_variables.this", map[string]string{
						"a": "1",
						"b": "2",
						"c": "2",
					}),
				},
			},
		},
	})
}

// TestAcc_VariablesDataSource_LandingZoneStyle mirrors how component modules consume placement metadata
// (account_id, region, region_partition) injected as provider vars.
func TestAcc_VariablesDataSource_LandingZoneStyle(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "context" {
					  vars = {
					    account_id       = "123456789012"
					    region           = "eu-central-1"
					    region_partition = "aws"
					    environment      = "dev"
					  }
					}

					data "context_namespace" "component" {
					  name    = "#aws-vpc"
					  context = { stack = [] }
					}

					data "context_variables" "vpc" {
					  context = data.context_namespace.component.context
					}

					output "kms_arn" {
					  value = "arn:${data.context_variables.vpc.vars.region_partition}:kms:${data.context_variables.vpc.vars.region}:${data.context_variables.vpc.vars.account_id}:key/abc"
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("kms_arn", knownvalue.StringExact("arn:aws:kms:eu-central-1:123456789012:key/abc")),
				},
			},
		},
	})
}

func TestAcc_VariablesDataSource_ValidationErrors(t *testing.T) {
	testAccPreCheck(t)

	tests := []struct {
		name      string
		config    string
		wantError *regexp.Regexp
	}{
		{
			name: "context is required",
			config: `
				provider "context" {}

				data "context_variables" "this" {}
			`,
			wantError: regexp.MustCompile(`(?s)Missing required argument.*"context" is required`),
		},
		{
			// context_variables is validated like a label: it must sit below at least one namespace.
			name: "empty stack is rejected",
			config: `
				provider "context" {}

				data "context_variables" "this" {
				  context = { stack = [] }
				}
			`,
			wantError: regexp.MustCompile(`Invalid context label order`),
		},
		{
			name: "stack ending with a label is rejected",
			config: `
				provider "context" {}

				data "context_variables" "this" {
				  context = {
				    stack = [
				      { name = "production", type = "ns" },
				      { name = "bucket", type = "lb" },
				    ]
				  }
				}
			`,
			wantError: regexp.MustCompile(`Invalid context label order`),
		},
		{
			name: "vars is read only",
			config: `
				provider "context" {}

				data "context_namespace" "env" {
				  name    = "production"
				  context = { stack = [] }
				}

				data "context_variables" "this" {
				  context = data.context_namespace.env.context
				  vars    = { a = "1" }
				}
			`,
			wantError: regexp.MustCompile(`(?s)Invalid Configuration for Read-Only Attribute.*vars`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config:      tt.config,
						ExpectError: tt.wantError,
					},
				},
			})
		})
	}
}
