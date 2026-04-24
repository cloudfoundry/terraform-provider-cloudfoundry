# This feature requires Terraform v1.14.0 or later (Stable as of 2026)
# List resources must be defined in .tfquery.hcl files.

# Generic template for a list block
list "cloudfoundry_mta" "<label_name>" {
  # (Required) Provider instance to use
  provider = provider_name

  config {
    # Provider specific filters
  }
}

# List block to discover all MTAs in a space.
list "cloudfoundry_mta" "all_mtas" {
  provider = cloudfoundry
  config {
    space = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  }
}

# List block to discover all MTAs in a space and include the resource data in the output.
list "cloudfoundry_mta" "all_mtas_with_data" {
  provider         = cloudfoundry
  include_resource = true
  config {
    space = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  }
}

# List block to discover MTAs in a specific namespace.
list "cloudfoundry_mta" "namespaced_mtas" {
  provider = cloudfoundry
  config {
    space     = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
    namespace = "my-namespace"
  }
}

# List block to discover MTAs via a custom deploy service URL.
list "cloudfoundry_mta" "custom_deploy_url_mtas" {
  provider         = cloudfoundry
  include_resource = true
  config {
    space      = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
    deploy_url = "https://deploy-service.example.com"
  }
}
