// Copyright SnapSoft 2026
// SPDX-License-Identifier: MIT

package provider

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

// labelConfig renders a minimal provider + namespace chain + label configuration.
// providerBody and namespaceBody are injected verbatim into the respective blocks.
func labelConfig(providerBody, namespaceBody, labelBody string) string {
	return `
		provider "context" {
		  ` + providerBody + `
		}

		data "context_namespace" "env" {
		  name    = "production"
		  context = { stack = [] }
		  ` + namespaceBody + `
		}

		data "context_label" "this" {
		  context = data.context_namespace.env.context
		  ` + labelBody + `
		}
	`
}

func TestAcc_LabelDataSource_DefaultId(t *testing.T) {
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

					data "context_namespace" "region" {
					  name    = "eu-west-1"
					  context = data.context_namespace.env.context
					}

					data "context_label" "bucket" {
					  name          = "assets"
					  resource_type = "s3-bucket"
					  context       = data.context_namespace.region.context
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					expectId("data.context_label.bucket", "production-eu-west-1-assets"),
					// Default tags are just {"Name": <id>}.
					expectTags("data.context_label.bucket", map[string]string{"Name": "production-eu-west-1-assets"}),
					statecheck.ExpectKnownValue("data.context_label.bucket", tfjsonpath.New("name"), knownvalue.StringExact("assets")),
					statecheck.ExpectKnownValue("data.context_label.bucket", tfjsonpath.New("resource_type"), knownvalue.StringExact("s3-bucket")),
				},
			},
		},
	})
}

func TestAcc_LabelDataSource_IdCasing(t *testing.T) {
	testAccPreCheck(t)

	tests := []struct {
		casing string
		wantId string
	}{
		{"kebab-case", "production-eu-west-1-assets"},
		{"snake_case", "production_eu-west-1_assets"},
		{"PascalCase", "ProductionEuWest1Assets"},
		{"camelCase", "productionEuWest1Assets"},
	}

	for _, tt := range tests {
		t.Run(tt.casing, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: `
							provider "context" {
							  id_casing = "` + tt.casing + `"
							}

							data "context_namespace" "env" {
							  name    = "production"
							  context = { stack = [] }
							}

							data "context_namespace" "region" {
							  name    = "eu-west-1"
							  context = data.context_namespace.env.context
							}

							data "context_label" "bucket" {
							  name          = "assets"
							  resource_type = "s3-bucket"
							  context       = data.context_namespace.region.context
							}
						`,
						ConfigStateChecks: []statecheck.StateCheck{
							expectId("data.context_label.bucket", tt.wantId),
						},
					},
				},
			})
		})
	}
}

func TestAcc_LabelDataSource_IdPrefix(t *testing.T) {
	testAccPreCheck(t)

	t.Run("from provider", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: labelConfig(`id_prefix = "acme"`, ``, `name = "assets"
						resource_type = "s3-bucket"`),
					ConfigStateChecks: []statecheck.StateCheck{
						expectId("data.context_label.this", "acme-production-assets"),
					},
				},
			},
		})
	})

	t.Run("from namespace", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: labelConfig(``, `id_prefix = "acme"`, `name = "assets"
						resource_type = "s3-bucket"`),
					ConfigStateChecks: []statecheck.StateCheck{
						expectId("data.context_label.this", "acme-production-assets"),
					},
				},
			},
		})
	})

	t.Run("prefix is cased together with the rest of the id", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: labelConfig(`id_prefix = "ACME"
						id_casing = "snake_case"`, ``, `name = "assets"
						resource_type = "s3-bucket"`),
					ConfigStateChecks: []statecheck.StateCheck{
						expectId("data.context_label.this", "acme_production_assets"),
					},
				},
			},
		})
	})
}

func TestAcc_LabelDataSource_IncludeResourceTypeInId(t *testing.T) {
	testAccPreCheck(t)

	t.Run("from provider", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: labelConfig(`include_resource_type_in_id = true`, ``, `name = "jobs"
						resource_type = "queue"`),
					ConfigStateChecks: []statecheck.StateCheck{
						expectId("data.context_label.this", "production-jobs-queue"),
					},
				},
			},
		})
	})

	t.Run("from namespace", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: labelConfig(``, `include_resource_type_in_id = true`, `name = "jobs"
						resource_type = "queue"`),
					ConfigStateChecks: []statecheck.StateCheck{
						expectId("data.context_label.this", "production-jobs-queue"),
					},
				},
			},
		})
	})

	t.Run("explicit false on provider", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: labelConfig(`include_resource_type_in_id = false`, ``, `name = "jobs"
						resource_type = "queue"`),
					ConfigStateChecks: []statecheck.StateCheck{
						expectId("data.context_label.this", "production-jobs"),
					},
				},
			},
		})
	})
}

func TestAcc_LabelDataSource_CombinedIdAttributes(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: labelConfig(``, `id_prefix                   = "snap"
					id_casing                   = "snake_case"
					include_resource_type_in_id = true`, `name = "jobs"
					resource_type = "function"`),
				ConfigStateChecks: []statecheck.StateCheck{
					expectId("data.context_label.this", "snap_production_jobs_function"),
					expectTags("data.context_label.this", map[string]string{"Name": "snap_production_jobs_function"}),
				},
			},
		},
	})
}

// TestAcc_LabelDataSource_NamespaceOverridesProvider verifies that id-shaping attributes declared on a
// namespace take precedence over the provider defaults, and that the deepest namespace wins.
func TestAcc_LabelDataSource_NamespaceOverridesProvider(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "context" {
					  id_casing                   = "kebab-case"
					  id_prefix                   = "provider"
					  include_resource_type_in_id = false
					}

					data "context_namespace" "root" {
					  name                        = "root"
					  context                     = { stack = [] }
					  id_casing                   = "snake_case"
					  id_prefix                   = "rootprefix"
					  include_resource_type_in_id = true
					}

					data "context_namespace" "leaf" {
					  name      = "leaf"
					  context   = data.context_namespace.root.context
					  id_casing = "PascalCase"
					  id_prefix = "leafprefix"
					}

					data "context_label" "this" {
					  name          = "svc"
					  resource_type = "lambda"
					  context       = data.context_namespace.leaf.context
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					// casing+prefix from leaf, include_resource_type from root (the only namespace setting it).
					expectId("data.context_label.this", "LeafprefixRootLeafSvcLambda"),
				},
			},
		},
	})
}

func TestAcc_LabelDataSource_DeeperNamespaceCanClearPrefix(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "context" {
					  id_prefix = "acme"
					}

					data "context_namespace" "root" {
					  name    = "production"
					  context = { stack = [] }
					}

					data "context_namespace" "leaf" {
					  name      = "api"
					  context   = data.context_namespace.root.context
					  id_prefix = ""
					}

					data "context_label" "this" {
					  name          = "svc"
					  resource_type = "lambda"
					  context       = data.context_namespace.leaf.context
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					expectId("data.context_label.this", "production-api-svc"),
				},
			},
		},
	})
}

func TestAcc_LabelDataSource_EmptyLabelName(t *testing.T) {
	testAccPreCheck(t)

	// Mirrors the landing-zone pattern `name = ""` with include_resource_type_in_id enabled.
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: labelConfig(`include_resource_type_in_id = true`, ``, `name = ""
					resource_type = "vpc"`),
				ConfigStateChecks: []statecheck.StateCheck{
					expectId("data.context_label.this", "production-vpc"),
					expectTags("data.context_label.this", map[string]string{"Name": "production-vpc"}),
				},
			},
		},
	})
}

func TestAcc_LabelDataSource_EmptyNamespaceNamesAreSkipped(t *testing.T) {
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

					data "context_label" "this" {
					  name          = "svc"
					  resource_type = "lambda"
					  context       = data.context_namespace.anonymous.context
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					expectId("data.context_label.this", "production-svc"),
				},
			},
		},
	})
}

func TestAcc_LabelDataSource_ExplicitTags(t *testing.T) {
	testAccPreCheck(t)

	t.Run("non-conflicting tags are merged with computed tags", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: labelConfig(``, ``, `name = "assets"
						resource_type = "s3-bucket"
						tags = {
						  CostCenter = "cc-42"
						  Owner      = "platform-team"
						}`),
					ConfigStateChecks: []statecheck.StateCheck{
						expectTags("data.context_label.this", map[string]string{
							"Name":       "production-assets",
							"CostCenter": "cc-42",
							"Owner":      "platform-team",
						}),
					},
				},
			},
		})
	})

	t.Run("explicit tags win over default Name tag", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: labelConfig(``, ``, `name = "assets"
						resource_type = "s3-bucket"
						tags = {
						  Name = "custom-name"
						}`),
					ConfigStateChecks: []statecheck.StateCheck{
						expectId("data.context_label.this", "production-assets"),
						expectTags("data.context_label.this", map[string]string{"Name": "custom-name"}),
					},
				},
			},
		})
	})

	t.Run("explicit tags win over mapper-produced tags", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: labelConfig(`
						mappers = [
						  {
						    name     = "tags"
						    function = "(.context.outputs.tags += {Team: \"from-mapper\", Env: \"prod\"}).context"
						  }
						]`, ``, `name = "assets"
						resource_type = "s3-bucket"
						tags = {
						  Team = "explicit"
						}`),
					ConfigStateChecks: []statecheck.StateCheck{
						expectTags("data.context_label.this", map[string]string{
							"Team": "explicit",
							"Env":  "prod",
						}),
					},
				},
			},
		})
	})

	t.Run("empty explicit tags map", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: labelConfig(``, ``, `name = "assets"
						resource_type = "s3-bucket"
						tags = {}`),
					ConfigStateChecks: []statecheck.StateCheck{
						expectTags("data.context_label.this", map[string]string{"Name": "production-assets"}),
					},
				},
			},
		})
	})
}

func TestAcc_LabelDataSource_Mappers(t *testing.T) {
	testAccPreCheck(t)

	t.Run("mapper id overrides default id generation", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: labelConfig(`id_prefix = "ignored"`, `
						vars = {
						  environment = "production"
						  prefix      = "acme"
						}
						mappers = [
						  {
						    name     = "Build custom ID"
						    function = "(.context.outputs += {id: ([.vars.prefix, .vars.environment, (.stack[] | select(.type == \"lb\") | .name)] | join(\"-\"))}).context"
						  }
						]`, `name = "assets"
						resource_type = "s3-bucket"`),
					ConfigStateChecks: []statecheck.StateCheck{
						expectId("data.context_label.this", "acme-production-assets"),
						// When a mapper sets only the id, tags default to {"Name": <mapper id>}.
						expectTags("data.context_label.this", map[string]string{"Name": "acme-production-assets"}),
					},
				},
			},
		})
	})

	t.Run("mapper tags replace default Name tag", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: labelConfig(``, `
						vars = { prefix = "acme" }
						mappers = [
						  {
						    name     = "Always tag with the ID prefix"
						    function = "(.context.outputs.tags += {Prefix: .vars.prefix}).context"
						  }
						]`, `name = "assets"
						resource_type = "s3-bucket"`),
					ConfigStateChecks: []statecheck.StateCheck{
						expectId("data.context_label.this", "production-assets"),
						// Once a mapper produces tags, the implicit Name tag is no longer added.
						expectTags("data.context_label.this", map[string]string{"Prefix": "acme"}),
					},
				},
			},
		})
	})

	t.Run("run_condition gates mapper execution", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: labelConfig(``, `
						vars = { environment = "production" }
						mappers = [
						  {
						    name          = "Tag sandbox resources"
						    run_condition = ".vars.environment == \"sbx\""
						    function      = "(.context.outputs.tags += {CostCentre: \"sandbox\"}).context"
						  },
						  {
						    name          = "Tag production resources"
						    run_condition = ".vars.environment == \"production\""
						    function      = "(.context.outputs.tags += {Tier: \"prod\"}).context"
						  }
						]`, `name = "assets"
						resource_type = "s3-bucket"`),
					ConfigStateChecks: []statecheck.StateCheck{
						expectTags("data.context_label.this", map[string]string{"Tier": "prod"}),
					},
				},
			},
		})
	})

	t.Run("mapper can access resource_type and label vars", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: labelConfig(``, `
						mappers = [
						  {
						    name     = "id from resource type"
						    function = "(.context.outputs += {id: (.vars.resource_type + \"/\" + (.stack[-1].vars.resource_type) + \"/\" + .stack[-1].type)}).context"
						  }
						]`, `name = "assets"
						resource_type = "s3-bucket"`),
					ConfigStateChecks: []statecheck.StateCheck{
						expectId("data.context_label.this", "s3-bucket/s3-bucket/lb"),
					},
				},
			},
		})
	})

	t.Run("mapper can access stack element vars", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: labelConfig(``, `
						vars = { environment = "production" }
						mappers = [
						  {
						    name     = "id from first stack element var"
						    function = "(.context.outputs += {id: .stack[0].vars.environment}).context"
						  }
						]`, `name = "assets"
						resource_type = "s3-bucket"`),
					ConfigStateChecks: []statecheck.StateCheck{
						expectId("data.context_label.this", "production"),
					},
				},
			},
		})
	})
}

// TestAcc_LabelDataSource_MapperEvaluationOrder verifies the documented ordering: namespace mappers are
// evaluated bottom-up (leaf namespace first, root namespace last) and provider mappers run after all of them.
func TestAcc_LabelDataSource_MapperEvaluationOrder(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "context" {
					  mappers = [
					    {
					      name     = "provider"
					      function = "(.context.outputs += {id: ((.context.outputs.id // \"\") + \"provider\")}).context"
					    }
					  ]
					}

					data "context_namespace" "root" {
					  name    = "root"
					  context = { stack = [] }
					  mappers = [
					    {
					      name     = "root-1"
					      function = "(.context.outputs += {id: ((.context.outputs.id // \"\") + \"root1,\")}).context"
					    },
					    {
					      name     = "root-2"
					      function = "(.context.outputs += {id: ((.context.outputs.id // \"\") + \"root2,\")}).context"
					    }
					  ]
					}

					data "context_namespace" "leaf" {
					  name    = "leaf"
					  context = data.context_namespace.root.context
					  mappers = [
					    {
					      name     = "leaf"
					      function = "(.context.outputs += {id: ((.context.outputs.id // \"\") + \"leaf,\")}).context"
					    }
					  ]
					}

					data "context_label" "this" {
					  name          = "svc"
					  resource_type = "lambda"
					  context       = data.context_namespace.leaf.context
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					expectId("data.context_label.this", "leaf,root1,root2,provider"),
				},
			},
		},
	})
}

// TestAcc_LabelDataSource_VarsPrecedence verifies the `.vars` object seen by mappers: provider vars are the
// lowest priority, then each namespace top-down, and the label's implicit resource_type var is on top.
func TestAcc_LabelDataSource_VarsPrecedence(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "context" {
					  vars = {
					    a             = "provider"
					    b             = "provider"
					    c             = "provider"
					    resource_type = "provider"
					  }
					  mappers = [
					    {
					      name     = "dump vars"
					      function = "(.context.outputs += {id: ([.vars.a, .vars.b, .vars.c, .vars.resource_type] | join(\"/\"))}).context"
					    }
					  ]
					}

					data "context_namespace" "root" {
					  name    = "root"
					  context = { stack = [] }
					  vars = {
					    b = "root"
					    c = "root"
					  }
					}

					data "context_namespace" "leaf" {
					  name    = "leaf"
					  context = data.context_namespace.root.context
					  vars = {
					    c = "leaf"
					  }
					}

					data "context_label" "this" {
					  name          = "svc"
					  resource_type = "lambda"
					  context       = data.context_namespace.leaf.context
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					expectId("data.context_label.this", "provider/root/leaf/lambda"),
				},
			},
		},
	})
}

// TestAcc_LabelDataSource_LandingZoneStyleMapper exercises the real-world provider-level mapper used by the
// SnapSoft AWS landing zone: '#'-prefixed namespaces are excluded from the id, the label name and the
// resource type are appended, and empty parts are dropped.
func TestAcc_LabelDataSource_LandingZoneStyleMapper(t *testing.T) {
	testAccPreCheck(t)

	const mapperFunction = `(.context.outputs += {id: ([.vars.prefix] + [.vars.environment] + [.stack[] | select(.type == \"ns\" and (.name | startswith(\"#\") | not)) | .name] + [.stack[] | select(.type == \"lb\") | .name] + (if (.vars.resource_type // \"\" | startswith(\"#\")) then [] else [.vars.resource_type] end) | map(select(. != null and . != \"\")) | join(\"-\"))}).context`

	config := func(disabled string) string {
		return `
			provider "context" {
			  id_casing                   = "kebab-case"
			  include_resource_type_in_id = true
			  vars = {
			    prefix      = "snap"
			    environment = "dev"
			  }
			  mappers = [
			    {
			      name          = "LabelIdOverride"
			      run_condition = ".vars.context_config_default_id_generation_disabled != \"true\""
			      function      = "` + mapperFunction + `"
			    }
			  ]
			}

			data "context_namespace" "network" {
			  name    = "network"
			  context = { stack = [] }
			}

			data "context_namespace" "component" {
			  name    = "#aws-vpc"
			  context = data.context_namespace.network.context
			  vars = {
			    context_config_default_id_generation_disabled = "` + disabled + `"
			  }
			}

			data "context_label" "vpc" {
			  name          = ""
			  resource_type = "vpc"
			  context       = data.context_namespace.component.context
			}

			data "context_label" "flow_log_role" {
			  name          = "vpc-flow-log"
			  resource_type = "role"
			  context       = data.context_namespace.component.context
			}

			data "context_label" "hidden_type" {
			  name          = "secret"
			  resource_type = "#secret"
			  context       = data.context_namespace.component.context
			}
		`
	}

	t.Run("mapper enabled", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: config("false"),
					ConfigStateChecks: []statecheck.StateCheck{
						expectId("data.context_label.vpc", "snap-dev-network-vpc"),
						expectId("data.context_label.flow_log_role", "snap-dev-network-vpc-flow-log-role"),
						expectId("data.context_label.hidden_type", "snap-dev-network-secret"),
					},
				},
			},
		})
	})

	t.Run("mapper disabled falls back to default id generation", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: config("true"),
					ConfigStateChecks: []statecheck.StateCheck{
						// Default generation does not know about the '#' convention and includes every namespace.
						expectId("data.context_label.vpc", "network-#aws-vpc-vpc"),
						expectId("data.context_label.flow_log_role", "network-#aws-vpc-vpc-flow-log-role"),
					},
				},
			},
		})
	})
}

func TestAcc_LabelDataSource_MultipleLabelsShareNamespace(t *testing.T) {
	testAccPreCheck(t)

	// Mirrors modules declaring many labels against one namespace: evaluating one label must not leak
	// its stack element or vars into another.
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "context" {
					  include_resource_type_in_id = true
					  mappers = [
					    {
					      name     = "count labels"
					      function = "(.context.outputs.tags += {LabelCount: ([.stack[] | select(.type == \"lb\")] | length | tostring)}).context"
					    }
					  ]
					}

					data "context_namespace" "env" {
					  name    = "production"
					  context = { stack = [] }
					}

					data "context_label" "a" {
					  name          = "alpha"
					  resource_type = "sg"
					  context       = data.context_namespace.env.context
					}

					data "context_label" "b" {
					  name          = "beta"
					  resource_type = "subnet"
					  context       = data.context_namespace.env.context
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					expectId("data.context_label.a", "production-alpha-sg"),
					expectId("data.context_label.b", "production-beta-subnet"),
					expectTags("data.context_label.a", map[string]string{"LabelCount": "1"}),
					expectTags("data.context_label.b", map[string]string{"LabelCount": "1"}),
				},
			},
		},
	})
}

func TestAcc_LabelDataSource_ContextIsClearedInStateByDefault(t *testing.T) {
	testAccPreCheck(t)
	if strings.EqualFold(os.Getenv("TF_LOG"), "DEBUG") {
		t.Skip("context is intentionally retained in state when TF_LOG=DEBUG")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: labelConfig(``, ``, `name = "assets"
					resource_type = "s3-bucket"`),
				ConfigStateChecks: []statecheck.StateCheck{
					// The (potentially large) context stack is dropped from state to keep plans readable.
					statecheck.ExpectKnownValue("data.context_label.this", stackPath(), knownvalue.ListSizeExact(0)),
				},
			},
		},
	})
}

func TestAcc_LabelDataSource_ContextIsRetainedInStateWithDebugLog(t *testing.T) {
	testAccPreCheck(t)
	// The provider runs in-process, so its os.Getenv("TF_LOG") sees this test's environment.
	t.Setenv("TF_LOG", "DEBUG")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: labelConfig(``, ``, `name = "assets"
					resource_type = "s3-bucket"`),
				ConfigStateChecks: []statecheck.StateCheck{
					// Namespace element + the label's own element.
					statecheck.ExpectKnownValue("data.context_label.this", stackPath(), knownvalue.ListSizeExact(2)),
					statecheck.ExpectKnownValue("data.context_label.this", stackElementPath(0, "name"), knownvalue.StringExact("production")),
					statecheck.ExpectKnownValue("data.context_label.this", stackElementPath(0, "type"), knownvalue.StringExact("ns")),
					statecheck.ExpectKnownValue("data.context_label.this", stackElementPath(1, "name"), knownvalue.StringExact("assets")),
					statecheck.ExpectKnownValue("data.context_label.this", stackElementPath(1, "type"), knownvalue.StringExact("lb")),
					statecheck.ExpectKnownValue("data.context_label.this", stackElementPath(1, "vars"), knownvalue.MapExact(map[string]knownvalue.Check{
						"resource_type": knownvalue.StringExact("s3-bucket"),
					})),
				},
			},
		},
	})
}

func TestAcc_LabelDataSource_ValidationErrors(t *testing.T) {
	testAccPreCheck(t)

	tests := []struct {
		name      string
		config    string
		wantError *regexp.Regexp
	}{
		{
			name: "label requires at least one namespace",
			config: `
				provider "context" {}

				data "context_label" "this" {
				  name          = "assets"
				  resource_type = "s3-bucket"
				  context       = { stack = [] }
				}
			`,
			wantError: regexp.MustCompile(`Invalid context label order`),
		},
		{
			name: "label cannot follow another label",
			config: `
				provider "context" {}

				data "context_label" "this" {
				  name          = "assets"
				  resource_type = "s3-bucket"
				  context = {
				    stack = [
				      { name = "production", type = "ns" },
				      { name = "other", type = "lb" },
				    ]
				  }
				}
			`,
			wantError: regexp.MustCompile(`Invalid context label order`),
		},
		{
			name: "resource_type must be at least 2 characters",
			config: `
				provider "context" {}

				data "context_namespace" "env" {
				  name    = "production"
				  context = { stack = [] }
				}

				data "context_label" "this" {
				  name          = "assets"
				  resource_type = "s"
				  context       = data.context_namespace.env.context
				}
			`,
			wantError: regexp.MustCompile(`(?s)Invalid Attribute Value Length.*resource_type`),
		},
		{
			name: "name is required",
			config: `
				provider "context" {}

				data "context_namespace" "env" {
				  name    = "production"
				  context = { stack = [] }
				}

				data "context_label" "this" {
				  resource_type = "s3-bucket"
				  context       = data.context_namespace.env.context
				}
			`,
			wantError: regexp.MustCompile(`(?s)Missing required argument.*"name" is required`),
		},
		{
			name: "resource_type is required",
			config: `
				provider "context" {}

				data "context_namespace" "env" {
				  name    = "production"
				  context = { stack = [] }
				}

				data "context_label" "this" {
				  name    = "assets"
				  context = data.context_namespace.env.context
				}
			`,
			wantError: regexp.MustCompile(`(?s)Missing required argument.*"resource_type" is required`),
		},
		{
			name: "context is required",
			config: `
				provider "context" {}

				data "context_label" "this" {
				  name          = "assets"
				  resource_type = "s3-bucket"
				}
			`,
			wantError: regexp.MustCompile(`(?s)Missing required argument.*"context" is required`),
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

func TestAcc_LabelDataSource_MapperEvaluationErrors(t *testing.T) {
	testAccPreCheck(t)

	tests := []struct {
		name      string
		mappers   string
		wantError *regexp.Regexp
	}{
		{
			name: "run_condition returns non boolean",
			mappers: `[
				{ name = "bad-cond", run_condition = "\"yes\"", function = ".context" }
			]`,
			wantError: regexp.MustCompile(`(?s)Failed to evaluate the context with the given mappers.*invalid return type for "bad-cond" run condition`),
		},
		{
			name: "function raises an error",
			mappers: `[
				{ name = "boom", function = "error(\"boom\")" }
			]`,
			wantError: regexp.MustCompile(`(?s)Failed to evaluate the context with the given mappers.*mapper function "boom" has evaluation error.*boom`),
		},
		{
			name: "function returns a non object",
			mappers: `[
				{ name = "string", function = "\"just-a-string\"" }
			]`,
			wantError: regexp.MustCompile(`(?s)Failed to retrieve required output values from the 'context.main' data structure.*invalid context.main type`),
		},
		{
			name: "outputs is not an object",
			mappers: `[
				{ name = "bad-outputs", function = "(.context += {outputs: \"x\"}).context" }
			]`,
			wantError: regexp.MustCompile(`(?s)Failed to retrieve required output values.*invalid 'outputs' type`),
		},
		{
			name: "tags contain non string values",
			mappers: `[
				{ name = "bad-tags", function = "(.context.outputs.tags += {Count: 1}).context" }
			]`,
			wantError: regexp.MustCompile(`(?s)Failed to retrieve required output values.*expected to be a map of strings`),
		},
		{
			name: "namespace mapper with invalid jq syntax fails at evaluation",
			mappers: `[
				{ name = "ns-broken", function = ".context | |" }
			]`,
			wantError: regexp.MustCompile(`(?s)Failed to evaluate the context with the given mappers.*mapper function "ns-broken" has an invalid jq syntax`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: labelConfig(``, `mappers = `+tt.mappers, `name = "assets"
							resource_type = "s3-bucket"`),
						ExpectError: tt.wantError,
					},
				},
			})
		})
	}
}
