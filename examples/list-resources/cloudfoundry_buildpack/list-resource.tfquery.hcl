# This feature requires Terraform v1.14.0 or later (Stable as of 2026)
# List resources must be defined in .tfquery.hcl files.

# Generic template for a list block
list "cloudfoundry_buildpack" "<label_name>" {
  # (Required) Provider instance to use
  provider = provider_name

  config {
    # Provider specific filters
  }
}

# List block to discover all buildpacks.
list "cloudfoundry_buildpack" "all_buildpacks" {
  provider = cloudfoundry
}

# List block to discover all buildpacks and include the resource data in the output.
list "cloudfoundry_buildpack" "all_buildpacks_with_data" {
  provider         = cloudfoundry
  include_resource = true
}

# List block to discover all buildpacks for a specific stack.
list "cloudfoundry_buildpack" "stack_buildpacks" {
  provider = cloudfoundry
  config {
    stack = "cflinuxfs4"
  }
}
