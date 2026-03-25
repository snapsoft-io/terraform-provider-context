terraform {
  required_providers {
    context = {
      source = "snapsoft-io/context"
    }
  }
}

# Provider-level defaults apply to every data source unless overridden at the
# namespace level.  All three ID-shaping attributes (id_casing, id_prefix,
# include_resource_type_in_id) can also be set individually on each
# context_namespace data source for finer-grained control.
provider "context" {
  id_casing = "kebab-case"

  vars = {
    organization = "acme"
    team         = "platform"
  }

  # Optional: load JQ-based mapper functions from an external JSON file.
  # Inline mappers can alternatively be declared with the `mappers` block.
  mappers_file_path = "provider-mappers.json"
}

# ── Build the context stack ───────────────────────────────────────────────────

data "context_namespace" "env" {
  name    = "production"
  context = { stack = [] }
}

data "context_namespace" "region" {
  name    = "eu-west-1"
  context = data.context_namespace.env.context
}

# ── Produce a resource label ──────────────────────────────────────────────────

data "context_label" "bucket" {
  name          = "assets"
  resource_type = "s3-bucket"
  context       = data.context_namespace.region.context
}

output "bucket_id" {
  value = data.context_label.bucket.id
  # => "production-eu-west-1-assets"
}

output "bucket_tags" {
  value = data.context_label.bucket.tags
  # => { Name = "production-eu-west-1-assets", Team = "platform", Organization = "acme" }
}
