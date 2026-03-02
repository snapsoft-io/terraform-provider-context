terraform {
  required_providers {
    context = {
      source = "snapsoft-io/context"
    }
  }
}

provider "context" {
  mappers_file_path = "provider-mappers.json"
  vars = {
    organization_id = "0123456789"
  }
}

data "context_example_module_metadata" "test" {
  name = "test-example"
  vars = {
    id_prefix   = "snap"
    environment = "sbx"
  }
}

data "context_label_namespace" "test" {
  name    = "test-namespace"
  context = data.context_example_module_metadata.test.context
}

data "context_label_item" "test" {
  name          = "test-item"
  resource_type = "test-type"
  context       = data.context_label_namespace.test.context
}

output "id" {
  value = data.context_label_item.test.id
}

output "tags" {
  value = data.context_label_item.test.tags
}
