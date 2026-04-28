# This feature requires Terraform v1.14.0 or later (Stable as of 2026)
# List resources must be defined in .tfquery.hcl files.

# Generic template for a list block
list "cloudfoundry_service_route_binding" "<label_name>" {
  # (Required) Provider instance to use
  provider = provider_name

  config {
    # Provider specific filters
  }
}

# List block to discover all service route bindings.
list "cloudfoundry_service_route_binding" "all_bindings" {
  provider = cloudfoundry
}

# List block to discover all service route bindings and include the resource data in the output.
list "cloudfoundry_service_route_binding" "all_bindings_with_data" {
  provider         = cloudfoundry
  include_resource = true
}

# List block to discover all route bindings for a specific service instance.
list "cloudfoundry_service_route_binding" "instance_bindings" {
  provider = cloudfoundry
  config {
    service_instance = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  }
}

# List block to discover all route bindings for a specific route and include the resource data in the output.
list "cloudfoundry_service_route_binding" "route_bindings_with_data" {
  provider         = cloudfoundry
  include_resource = true
  config {
    route = "b4da43cd-2055-4d4d-ae6e-4066ce34a563"
  }
}
