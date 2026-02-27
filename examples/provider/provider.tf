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

data "context_variables" "this" {
  context = data.context_example_module_metadata.test.context
}

output "vars" {
  value = data.context_variables.this.vars
}
