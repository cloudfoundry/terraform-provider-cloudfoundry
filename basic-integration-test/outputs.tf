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

# Service broker outputs
output "service_broker_id" {
  value       = cloudfoundry_service_broker.simple_broker.id
  description = "The GUID of the registered service broker"
}

output "service_broker_name" {
  value       = cloudfoundry_service_broker.simple_broker.name
  description = "The name of the registered service broker"
}

# Service instance outputs
output "service_instance_id" {
  value       = cloudfoundry_service_instance.simple_instance.id
  description = "The GUID of the created service instance"
}

output "service_instance_name" {
  value       = cloudfoundry_service_instance.simple_instance.name
  description = "The name of the created service instance"
}

# User provided service outputs
output "ups_id" {
  value       = cloudfoundry_service_instance.simple_ups.id
  description = "The GUID of the user provided service"
}

output "ups_name" {
  value       = cloudfoundry_service_instance.simple_ups.name
  description = "The name of the user provided service"
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
    service_broker = {
      id   = cloudfoundry_service_broker.simple_broker.id
      name = cloudfoundry_service_broker.simple_broker.name
    }
    service_instance = {
      id   = cloudfoundry_service_instance.simple_instance.id
      name = cloudfoundry_service_instance.simple_instance.name
    }
    user_provided_service = {
      id   = cloudfoundry_service_instance.simple_ups.id
      name = cloudfoundry_service_instance.simple_ups.name
    }
    test_app = {
      id   = cloudfoundry_app.test_app.id
      name = cloudfoundry_app.test_app.name
    }
    users_with_roles = var.test_users[*].username
  }
  description = "Summary of all created resources"
}
