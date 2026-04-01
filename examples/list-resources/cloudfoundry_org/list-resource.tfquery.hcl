# This feature requires Terraform v1.14.0 or later (Stable as of 2026)
# List resources must be defined in .tfquery.hcl files.

# Generic template for a list block
list "cloudfoundry_org" "<label_name>" {
  # (Required) Provider instance to use
  provider = provider_name

}

# List block to discover all organizations.
list "cloudfoundry_org" "all_orgs" {
  provider = cloudfoundry
}

# List block to discover all organizations and include the resource data in the output.
list "cloudfoundry_org" "all_orgs_with_data" {
  provider         = cloudfoundry
  include_resource = true
}
