terraform {
  required_providers {
    context = {
      source = "snapsoft-io/context"
    }
  }
}

provider "context" {
  id_casing = "kebab-case"
}

# ── Base namespace (reused across examples below) ─────────────────────────────

data "context_namespace" "env" {
  name    = "production"
  context = { stack = [] }
}

data "context_namespace" "region" {
  name    = "eu-west-1"
  context = data.context_namespace.env.context
}

# ── Basic label ───────────────────────────────────────────────────────────────
# id   = "production-eu-west-1-assets"
# tags = { Name = "production-eu-west-1-assets" }

data "context_label" "basic" {
  name          = "assets"
  resource_type = "s3-bucket"
  context       = data.context_namespace.region.context
}

# ── With id_prefix ────────────────────────────────────────────────────────────
# id_prefix is set on the namespace; context_label inherits it automatically.
# id   = "acme-production-assets"
# tags = { Name = "acme-production-assets" }

data "context_namespace" "env_prefixed" {
  name      = "production"
  id_prefix = "acme"
  context   = { stack = [] }
}

data "context_label" "with_prefix" {
  name          = "assets"
  resource_type = "s3-bucket"
  context       = data.context_namespace.env_prefixed.context
}

# ── With include_resource_type_in_id ─────────────────────────────────────────
# id   = "production-jobs-queue"
# tags = { Name = "production-jobs-queue" }

data "context_namespace" "env_with_rtype" {
  name                        = "production"
  include_resource_type_in_id = true
  context                     = { stack = [] }
}

data "context_label" "with_resource_type" {
  name          = "jobs"
  resource_type = "queue"
  context       = data.context_namespace.env_with_rtype.context
}

# ── With PascalCase id_casing ─────────────────────────────────────────────────
# id   = "ProductionEuWest1Assets"
# tags = { Name = "ProductionEuWest1Assets" }

data "context_namespace" "env_pascal" {
  name      = "production"
  id_casing = "PascalCase"
  context   = { stack = [] }
}

data "context_namespace" "region_pascal" {
  name      = "eu-west-1"
  id_casing = "PascalCase"
  context   = data.context_namespace.env_pascal.context
}

data "context_label" "with_pascal_case" {
  name          = "assets"
  resource_type = "s3-bucket"
  context       = data.context_namespace.region_pascal.context
}

# ── Combining id_prefix + include_resource_type_in_id + snake_case ───────────
# id   = "snap_production_jobs_function"
# tags = { Name = "snap_production_jobs_function" }

data "context_namespace" "env_combined" {
  name                        = "production"
  id_prefix                   = "snap"
  id_casing                   = "snake_case"
  include_resource_type_in_id = true
  context                     = { stack = [] }
}

data "context_label" "combined_attrs" {
  name          = "jobs"
  resource_type = "function"
  context       = data.context_namespace.env_combined.context
}

# ── With extra tags merged in ─────────────────────────────────────────────────
# Explicitly provided tags are merged with the computed tags.
# Explicitly provided values win on key conflicts.
# tags = { Name = "production-eu-west-1-assets", CostCenter = "cc-42", Owner = "platform-team" }

data "context_label" "with_extra_tags" {
  name          = "assets"
  resource_type = "s3-bucket"
  context       = data.context_namespace.region.context

  tags = {
    CostCenter = "cc-42"
    Owner      = "platform-team"
  }
}

# ── With JQ mappers ───────────────────────────────────────────────────────────
# Mappers are JQ expressions declared on namespace data sources.
# They are evaluated bottom-up when context_label runs.
# A mapper can set context.outputs.id and/or context.outputs.tags.
# When a mapper sets the id, the default id-computation logic is skipped.
#
# id   = "acme-production-assets"
# tags = { Name = "acme-production-assets", Prefix = "acme" }
# Note: the second mapper does NOT run because environment is "production", not "sbx".

data "context_namespace" "env_mapped" {
  name    = "production"
  context = { stack = [] }
  vars = {
    environment = "production"
    prefix      = "acme"
  }
  mappers = [
    {
      name     = "Build custom ID from prefix and label name"
      function = "(.context.outputs += {id: ([.vars.prefix, .vars.environment, (.stack[] | select(.type == \"lb\") | .name)] | join(\"-\"))}).context"
    },
    {
      name          = "Tag sandbox resources with a cost centre"
      run_condition = ".vars.environment == \"sbx\""
      function      = "(.context.outputs.tags += {CostCentre: \"sandbox\"}).context"
    },
    {
      name     = "Always tag with the ID prefix"
      function = "(.context.outputs.tags += {Prefix: .vars.prefix}).context"
    }
  ]
}

data "context_label" "with_mappers" {
  name          = "assets"
  resource_type = "s3-bucket"
  context       = data.context_namespace.env_mapped.context
}

output "basic_id" {
  value = data.context_label.basic.id
}

output "with_prefix_id" {
  value = data.context_label.with_prefix.id
}

output "with_resource_type_id" {
  value = data.context_label.with_resource_type.id
}

output "with_pascal_case_id" {
  value = data.context_label.with_pascal_case.id
}

output "combined_id" {
  value = data.context_label.combined_attrs.id
}

output "with_extra_tags" {
  value = data.context_label.with_extra_tags.tags
}

output "mapped_id" {
  value = data.context_label.with_mappers.id
}

output "mapped_tags" {
  value = data.context_label.with_mappers.tags
}
