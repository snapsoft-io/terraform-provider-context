---
page_title: "context_label Data Source - context"
subcategory: ""
description: |-
  Reads the context stack built by one or more context_namespace data sources and produces a canonical resource ID and a tags map.
---

# context_label (Data Source)

The `context_label` data source consumes the context stack assembled by `context_namespace` calls and computes:

- **`id`** – a string composed of the namespace names (and optionally an ID prefix and the resource type), formatted according to the active `id_casing`.
- **`tags`** – a `Name`-keyed map defaulting to `{ Name = <id> }`, enriched by any mapper functions in the stack.

When a mapper sets `context.outputs.id`, that value is used directly and the default ID-computation logic is skipped.  When a mapper sets `context.outputs.tags`, those values are merged with any explicitly provided `tags`.

## Default ID computation

The ID is assembled from the following parts (in order), then cased:

1. `id_prefix` (if set on a namespace or the provider)
2. All namespace names in stack order
3. The `name` of this `context_label`
4. The `resource_type` value (only when `include_resource_type_in_id = true`)

The `id_casing`, `id_prefix`, and `include_resource_type_in_id` values are resolved from the **last** namespace in the stack that sets them, falling back to the provider-level defaults.

## Example Usage

```terraform
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

# ── Base namespace ────────────────────────────────────────────────────────────

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
# id = "acme-production-assets"

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
# id = "production-jobs-queue"

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
# id = "ProductionEuWest1Assets"

data "context_namespace" "env_pascal" {
  name      = "production"
  id_casing = "PascalCase"
  context   = { stack = [] }
}

data "context_namespace" "region_pascal" {
  name    = "eu-west-1"
  context = data.context_namespace.env_pascal.context
}

data "context_label" "with_pascal_case" {
  name          = "assets"
  resource_type = "s3-bucket"
  context       = data.context_namespace.region_pascal.context
}

# ── Combining id_prefix + include_resource_type_in_id + snake_case ───────────
# id = "snap_production_jobs_function"

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
# Mappers are declared on namespace data sources.  They are applied bottom-up
# when context_label runs.  A mapper can set context.outputs.id and/or
# context.outputs.tags; when id is set, the default computation is skipped.
#
# id   = "acme-production-assets"
# tags = { Name = "acme-production-assets", Prefix = "acme" }
# Note: the "Tag sandbox resources" mapper does NOT run because environment != "sbx".

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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `context` (Attributes) The context stack built by one or more `context_namespace` data sources. Must contain at least one namespace entry. (see [below for nested schema](#nestedatt--context))
- `name` (String) The name of this resource label. Included in the generated ID.
- `resource_type` (String) The type of the resource being labelled (e.g. `s3-bucket`). Included in the ID when `include_resource_type_in_id = true` on the active namespace or provider.

### Optional

- `tags` (Map of String) Additional tags merged with the computed tags. Explicit values win on key conflicts with mapper-produced tags.

### Read-Only

- `id` (String) The computed resource identifier, derived from namespace names and shaped by `id_casing`, `id_prefix`, and `include_resource_type_in_id`. Overridden by a mapper that sets `context.outputs.id`.
- `tags` (Map of String) The computed tags map. Defaults to `{ Name = <id> }`. Enriched by mapper functions and merged with any explicitly provided `tags`.

<a id="nestedatt--context"></a>
### Nested Schema for `context`

Required:

- `stack` (Attributes List) Ordered list of context stack elements. Must contain at least one `ns` (namespace) entry. (see [below for nested schema](#nestedatt--context--stack))

<a id="nestedatt--context--stack"></a>
### Nested Schema for `context.stack`

Required:

- `type` (String) The context type of this stack element. Must be `ns` (namespace) or `lb` (label).
- `name` (String) The name of this stack element.

Optional:

- `id_casing` (String) Casing style override for this stack element. One of: `kebab-case`, `snake_case`, `camelCase`, `PascalCase`.
- `id_prefix` (String) ID prefix override for this stack element.
- `include_resource_type_in_id` (Boolean) Whether to include the resource type in the ID for this stack element.
- `mappers` (Attributes List) Mapper functions attached to this stack element. (see [below for nested schema](#nestedatt--context--stack--mappers))
- `vars` (Map of String) Variables attached to this stack element.

<a id="nestedatt--context--stack--mappers"></a>
### Nested Schema for `context.stack.mappers`

Required:

- `function` (String) JQ expression that receives `{context, stack, vars}` and must return a new `context` object.
- `name` (String) Human-readable name for the mapper, used in error messages.

Optional:

- `run_condition` (String) JQ expression that receives `{context, stack, vars}` and must return a boolean. The mapper runs only when this evaluates to `true`.
