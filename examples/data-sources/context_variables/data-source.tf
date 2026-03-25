terraform {
  required_providers {
    context = {
      source = "snapsoft-io/context"
    }
  }
}

# Provider-level vars are the lowest-priority source.
provider "context" {
  vars = {
    organization = "acme"
    team         = "platform"
  }
}

# Namespace-level vars override provider vars on key conflicts.
# Vars accumulate top-down: each namespace can add or override keys.

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
    # Overrides the value set in the parent namespace.
    region = "eu-west-1-overridden"
  }
}

# context_variables merges vars from the provider and all stack elements in
# top-down order.  Keys from deeper namespaces override keys from higher ones.
data "context_variables" "this" {
  context = data.context_namespace.region.context
}

output "all_vars" {
  value = data.context_variables.this.vars
  # => {
  #   organization = "acme"
  #   team         = "platform"
  #   environment  = "production"
  #   region       = "eu-west-1-overridden"
  # }
}
