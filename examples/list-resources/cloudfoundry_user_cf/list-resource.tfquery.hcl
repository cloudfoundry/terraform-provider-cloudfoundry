# This feature requires Terraform v1.14.0 or later (Stable as of 2026)
# List resources must be defined in .tfquery.hcl files.

# List block to discover all Cloud Controller users.
list "cloudfoundry_user_cf" "all_users" {
  provider = cloudfoundry
}

# List block to discover all Cloud Controller users and include resource data.
list "cloudfoundry_user_cf" "all_users_with_data" {
  provider         = cloudfoundry
  include_resource = true
}

# List block to discover Cloud Controller users in a specific organization.
list "cloudfoundry_user_cf" "org_users" {
  provider = cloudfoundry
  config {
    org = "b4da43cd-2055-4d4d-ae6e-4066ce34a563"
  }
}

# List block to discover Cloud Controller users in a specific space.
list "cloudfoundry_user_cf" "space_users" {
  provider         = cloudfoundry
  include_resource = true
  config {
    space = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  }
}
