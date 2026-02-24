# =============================================================================
# Outputs - Integration Test Results
# =============================================================================

# Organization outputs
output "org_id" {
  value       = cloudfoundry_org.test_org.id
  description = "The GUID of the created organization"
}

output "org_name" {
  value       = cloudfoundry_org.test_org.name
  description = "The name of the created organization"
}

# Space outputs
output "space_id" {
  value       = cloudfoundry_space.test_space.id
  description = "The GUID of the created space"
}

output "space_name" {
  value       = cloudfoundry_space.test_space.name
  description = "The name of the created space"
}

# Role outputs
output "org_user_role_ids" {
  value       = { for k, v in cloudfoundry_org_role.org_users : k => v.id }
  description = "The GUIDs of the organization user roles"
}

# Application outputs
output "app_id" {
  value       = cloudfoundry_app.test_app.id
  description = "The GUID of the deployed application"
}

output "app_name" {
  value       = cloudfoundry_app.test_app.name
  description = "The name of the deployed application"
}


# Summary output
output "test_summary" {
  value = {
    organization = {
      id   = cloudfoundry_org.test_org.id
      name = cloudfoundry_org.test_org.name
    }
    space = {
      id   = cloudfoundry_space.test_space.id
      name = cloudfoundry_space.test_space.name
    }
    users_with_roles = var.test_users[*].username
  }
  description = "Summary of all created resources"
}
