# This feature requires Terraform v1.14.0 or later (Stable as of 2026)
# List resources must be defined in .tfquery.hcl files.

# Generic template for a list block
list "cloudfoundry_domain" "<label_name>" {
  # (Required) Provider instance to use
  provider = provider_name

  config {
    # Provider specific filters
  }
}

# List block to discover all domains.
list "cloudfoundry_domain" "all_domains" {
  provider = cloudfoundry
}

# List block to discover all domains and include the resource data in the output.
list "cloudfoundry_domain" "all_domains_with_data" {
  provider         = cloudfoundry
  include_resource = true
}

# List block to discover all domains scoped to a specific organization.
list "cloudfoundry_domain" "org_domains" {
  provider = cloudfoundry
  config {
    org = "b4da43cd-2055-4d4d-ae6e-4066ce34a563"
  }
}

# List block to discover all domains scoped to a specific organization and include the resource data in the output.
list "cloudfoundry_domain" "org_domains_with_data" {
  provider         = cloudfoundry
  include_resource = true
  config {
    org = "b4da43cd-2055-4d4d-ae6e-4066ce34a563"
  }
}
