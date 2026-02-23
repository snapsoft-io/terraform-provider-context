terraform {
  required_providers {
    context = {
      source = "snapsoft/context"
    }
  }
}

provider "context" {
  mappers_file_path = "config.json"
  vars = {
    organization_id = "0123456789"
  }
}

data "context_variables" "this" {
  context = data.context_example_module_metadata.test.context
}

data "context_example_module_metadata" "test" {
  name = "test_example"
  vars = {
    id_prefix   = "product"
    environment = "sbx"
  }
}

data "context_root_module_metadata" "test" {
  name    = "test_root"
  context = data.context_example_module_metadata.test.context
}

data "context_label_namespace" "test1" {
  name    = "test_namespace1"
  context = data.context_root_module_metadata.test.context
}

data "context_component_module_metadata" "test" {
  name    = "test_component"
  context = data.context_label_namespace.test1.context
}

data "context_label_namespace" "test2" {
  name    = "test_namespace2"
  context = data.context_component_module_metadata.test.context
}

data "context_label_item" "test" {
  name          = "test_item"
  resource_type = "test_type"
  context       = data.context_label_namespace.test2.context
}

output "id" {
  value = data.context_label_item.test.id
}

output "tags" {
  value = data.context_label_item.test.tags
}

output "vars" {
  value = data.context_variables.this.vars
}