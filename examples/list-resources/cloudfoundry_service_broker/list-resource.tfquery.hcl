# This feature requires Terraform v1.14.0 or later (Stable as of 2026)
# List resources must be defined in .tfquery.hcl files.

# Generic template for a list block
list "cloudfoundry_service_broker" "<label_name>" {
  # (Required) Provider instance to use
  provider = provider_name

  config {
    # Provider specific filters
  }
}

# List block to discover all service brokers.
list "cloudfoundry_service_broker" "all_service_brokers" {
  provider = cloudfoundry
}

# List block to discover all service brokers and include the resource data in the output.
list "cloudfoundry_service_broker" "all_service_brokers_with_data" {
  provider         = cloudfoundry
  include_resource = true
}

# List block to discover all service brokers scoped to a specific space and include the resource data in the output.
list "cloudfoundry_service_broker" "space_service_brokers_with_data" {
  provider         = cloudfoundry
  include_resource = true
  config {
    space = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  }
}
