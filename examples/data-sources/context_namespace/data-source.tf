terraform {
  required_providers {
    context = {
      source = "snapsoft-io/context"
    }
  }
}

provider "context" {}

# ── Basic: start a new context stack ─────────────────────────────────────────
# Pass { stack = [] } as the context to begin an empty stack.

data "context_namespace" "env" {
  name    = "production"
  context = { stack = [] }
}

# ── Stack a second namespace on top of the first ──────────────────────────────
# Each context_namespace appends one entry to context.stack.

data "context_namespace" "region" {
  name    = "eu-west-1"
  context = data.context_namespace.env.context
}

# ── id_prefix: prepend a fixed string to all IDs produced downstream ──────────
# A context_label that receives this namespace's context will have "acme"
# prepended to the generated ID.

data "context_namespace" "env_with_prefix" {
  name      = "production"
  id_prefix = "acme"
  context   = { stack = [] }
}

# ── id_casing: control the casing style of the generated ID ──────────────────
# Accepted values: "kebab-case" (default), "snake_case", "camelCase", "PascalCase".

data "context_namespace" "env_snake" {
  name      = "production"
  id_casing = "snake_case"
  context   = { stack = [] }
}

data "context_namespace" "env_pascal" {
  name      = "production"
  id_casing = "PascalCase"
  context   = { stack = [] }
}

# ── include_resource_type_in_id: append the resource_type to the ID ───────────
# The resource_type value comes from the context_label data source.

data "context_namespace" "env_with_rtype" {
  name                     = "production"
  include_resource_type_in_id = true
  context                  = { stack = [] }
}

# ── Combining all three ID-shaping attributes ─────────────────────────────────

data "context_namespace" "env_full" {
  name                     = "staging"
  id_prefix                = "acme"
  id_casing                = "snake_case"
  include_resource_type_in_id = true
  context                  = { stack = [] }
}

# ── vars: namespace-scoped variables, available to mappers at evaluation time ──

data "context_namespace" "env_with_vars" {
  name    = "production"
  context = { stack = [] }
  vars = {
    cost_center = "cc-42"
    owner       = "platform-team"
  }
}

# ── mappers: JQ expressions applied when context_label evaluates the stack ────
# Multiple namespaces can each carry their own mappers; they are applied
# bottom-up (leaf namespace first, root namespace last).

data "context_namespace" "env_with_mappers" {
  name    = "production"
  context = { stack = [] }
  vars = {
    environment = "production"
  }
  mappers = [
    {
      name     = "Tag with environment"
      function = "(.context.outputs.tags += {Environment: .vars.environment}).context"
    },
    {
      name          = "Tag with cost centre (non-production only)"
      run_condition = ".vars.environment != \"production\""
      function      = "(.context.outputs.tags += {CostCentre: \"sandbox\"}).context"
    }
  ]
}
