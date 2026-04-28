# This feature requires Terraform v1.14.0 or later (Stable as of 2026)
# List resources must be defined in .tfquery.hcl files.

# Generic template for a list block
list "cloudfoundry_service_credential_binding" "<label_name>" {
  # (Required) Provider instance to use
  provider = provider_name

  config {
    # Provider specific filters
  }
}

# List block to discover all service credential bindings.
list "cloudfoundry_service_credential_binding" "all_bindings" {
  provider = cloudfoundry
}

# List block to discover all service credential bindings and include the resource data in the output.
list "cloudfoundry_service_credential_binding" "all_bindings_with_data" {
  provider         = cloudfoundry
  include_resource = true
}

# List block to discover all bindings for a specific service instance and include the resource data in the output.
list "cloudfoundry_service_credential_binding" "instance_bindings_with_data" {
  provider         = cloudfoundry
  include_resource = true
  config {
    service_instance = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  }
}

# List block to discover all bindings for a specific app.
list "cloudfoundry_service_credential_binding" "app_bindings" {
  provider = cloudfoundry
  config {
    app = "b4da43cd-2055-4d4d-ae6e-4066ce34a563"
  }
}

# List block to discover all bindings for a specific app and include the resource data in the output.
list "cloudfoundry_service_credential_binding" "app_bindings_with_data" {
  provider         = cloudfoundry
  include_resource = true
  config {
    app = "b4da43cd-2055-4d4d-ae6e-4066ce34a563"
  }
}