# This feature requires Terraform v1.14.0 or later (Stable as of 2026)
# List resources must be defined in .tfquery.hcl files.

# Generic template for a list block
list "cloudfoundry_org_quota" "<label_name>" {
  # (Required) Provider instance to use
  provider = provider_name

}

# List block to discover all organization quotas.
list "cloudfoundry_org_quota" "all_org_quotas" {
  provider = cloudfoundry
}

# List block to discover all organization quotas and include the resource data in the output.
list "cloudfoundry_org_quota" "all_org_quotas_with_data" {
  provider         = cloudfoundry
  include_resource = true
}
