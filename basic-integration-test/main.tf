# Note:- Sleep resources are placed in several places to allow for UAA/CC
# etc to settle and also for apps to start up when necessary.
# Main reason is that this is a limited resource constrained environment and we want to avoid any unnecessary failures
# due to resources not being ready.

# =================================================
# 1. Service Broker Organization and Space
# =============================================================================

resource "cloudfoundry_org" "broker_org" {
  name = var.broker_org_name
  labels = {
    "purpose"     = "service-broker"
    "managed-by"  = "terraform"
    "environment" = "e2e-test"
  }
  annotations = {
    "description" = "Organization for hosting service broker application"
  }
}

resource "cloudfoundry_space" "broker_space" {
  name      = var.broker_space_name
  org       = cloudfoundry_org.broker_org.id
  allow_ssh = true
  labels = {
    "purpose"     = "service-broker"
    "managed-by"  = "terraform"
    "environment" = "e2e-test"
  }
  annotations = {
    "description" = "Space for hosting service broker application"
  }
}

# =============================================================================
# Service Broker Application Deployment
# =============================================================================

resource "zipper_file" "broker_fixture" {
  source      = "simple-service-broker"
  output_path = "simple-service-broker.zip"
}

resource "cloudfoundry_app" "service_broker" {

  name           = var.broker_app_name
  org_name       = cloudfoundry_org.broker_org.name
  space_name     = cloudfoundry_space.broker_space.name
  path           = zipper_file.broker_fixture.output_path
  source_code_hash = zipper_file.broker_fixture.output_sha
  memory         = var.broker_app_memory
  instances      = 1
  strategy       = "rolling"
  routes = [
    {
      route = "simple-service-broker.apps.127-0-0-1.nip.io"
    }
  ]

  labels = {
    "purpose"     = "service-broker"
    "managed-by"  = "terraform"
    "environment" = "e2e-test"
  }
  annotations = {
    "description" = "Simple service broker for E2E testing"
  }

  environment = {
    BROKER_USERNAME = var.broker_username
    BROKER_PASSWORD = var.broker_password
  }

  depends_on = [cloudfoundry_space.broker_space]
}


resource "time_sleep" "wait_sometime_broker_app" {
  depends_on = [cloudfoundry_app.service_broker]

  create_duration = "60s"
}

# =============================================================================
# Register Service Broker
# =============================================================================

resource "cloudfoundry_service_broker" "simple_broker" {
  name     = var.service_broker_name
  url = "https://simple-service-broker.apps.127-0-0-1.nip.io"
  
  username = var.broker_username
  password = var.broker_password

  depends_on = [
    cloudfoundry_app.service_broker,
    time_sleep.wait_sometime_broker_app]
}

data "cloudfoundry_service_plan" "simple_plan" {
   depends_on = [ cloudfoundry_service_broker.simple_broker ]
   name                  = "simple"
   service_offering_name = "simple-service"
}

#Make the service plan available to all orgs
resource "cloudfoundry_service_plan_visibility" "test_visibility" {
  service_plan  = data.cloudfoundry_service_plan.simple_plan.id
  type          = "public"
}

# =============================================================================
# 2. Create Organization (Test Org)
# =============================================================================

resource "cloudfoundry_org" "test_org" {
  name = var.org_name
  labels = {
    "purpose"     = "integration-test"
    "managed-by"  = "terraform"
    "environment" = "e2e-test"
  }
  annotations = {
    "description" = "Organization created for E2E integration testing"
  }
}
#Loop var.test_users and add to cloudfoundry_user
resource "cloudfoundry_user" "test_users" {
  for_each = { for idx, user in var.test_users : user.username => user }

  username = each.value.username
  email       = each.value.email
  password = "abc"
  origin   = each.value.origin
}
# =============================================================================
# 3. Organization Roles - Assign roles to test users
# =============================================================================

# Organization User role (required before other org roles)
resource "cloudfoundry_org_role" "org_users" {
  for_each = { for idx, user in var.test_users : user.username => user }

  username = each.value.username
  type     = "organization_user"
  org      = cloudfoundry_org.test_org.id
  origin   = each.value.origin
  depends_on = [cloudfoundry_user.test_users]
}
# Organization Manager role for the first user
resource "cloudfoundry_org_role" "org_manager" {
  for_each = { for idx, user in var.test_users : user.username => user }
  username = each.value.username
  type     = "organization_manager"
  org      = cloudfoundry_org.test_org.id
  origin   = each.value.origin

  depends_on = [cloudfoundry_user.test_users]
}
# =============================================================================
# 6. Create Space
# =============================================================================

resource "cloudfoundry_space" "test_space" {
  name      = var.space_name
  org       = cloudfoundry_org.test_org.id
  allow_ssh = true
  labels = {
    "purpose"     = "integration-test"
    "managed-by"  = "terraform"
    "environment" = "e2e-test"
  }
  annotations = {
    "description" = "Space created for E2E integration testing"
  }
}

# =============================================================================
# 7. Space Roles - Assign roles to test users
# =============================================================================

# Space Developer role for the first user
resource "cloudfoundry_space_role" "space_developer" {
  for_each = { for idx, user in var.test_users : user.username => user }

  username = each.value.username
  type     = "space_developer"
  space    = cloudfoundry_space.test_space.id
  origin   = each.value.origin

  depends_on = [cloudfoundry_org_role.org_users]
}

# =============================================================================
# 8. Create Service Instance from Broker
# =============================================================================

resource "cloudfoundry_service_instance" "simple_instance" {
  name            = "simple-service-simple"
  type = "managed"
  service_plan    = data.cloudfoundry_service_plan.simple_plan.id
  tags        = ["terraform-test", "managed-service"]
  space           = cloudfoundry_space.test_space.id

  depends_on = [
    cloudfoundry_service_broker.simple_broker,
    cloudfoundry_service_plan_visibility.test_visibility
  ]
}

# =============================================================================
# 9. Create User Provided Service
# =============================================================================

resource "cloudfoundry_service_instance" "simple_ups" {
  name  = "simple-ups"
  type  = "user-provided"
  space = cloudfoundry_space.test_space.id
  tags        = ["terraform-test", "usp"]
  credentials = <<EOT
  {
    "user": "user1",
    "password": "demo122"
  }
  EOT

  depends_on = [cloudfoundry_space.test_space]
}

# =============================================================================
# 10. Zipper - Package test application source code into a zip file for deployment
# =============================================================================


resource "zipper_file" "fixture" {
  source      = "test-app"
  output_path = "test-app.zip"
}

# =============================================================================
# 11. Deploy Go Application 
# =============================================================================

resource "time_sleep" "wait_sometime_ups" {
  depends_on = [cloudfoundry_service_instance.simple_ups,cloudfoundry_service_instance.simple_instance]

  create_duration = "60s"
}
resource "cloudfoundry_app" "test_app" {
  name       = var.app_name
  org_name   = cloudfoundry_org.test_org.name
  space_name = cloudfoundry_space.test_space.name
  path       = zipper_file.fixture.output_path
  source_code_hash = zipper_file.fixture.output_sha
  memory     = "64M"
  instances  = 1
  strategy   = "rolling"
  environment = {
    MY_ENV = "red",
  }

  labels = {
    "purpose"     = "integration-test"
    "managed-by"  = "terraform"
    "environment" = "e2e-test"
  }
  annotations = {
    "description" = "Test GO application for E2E testing with service bindings"
  }
  service_bindings = [
    {service_instance = cloudfoundry_service_instance.simple_ups.name}
  ]

  depends_on = [
    cloudfoundry_space.test_space,
    cloudfoundry_space_role.space_developer,
    cloudfoundry_service_instance.simple_instance,
    cloudfoundry_service_instance.simple_ups,
    time_sleep.wait_sometime_ups
  ]
}

# =============================================================================
# 12. Bind Applications to Services
# =============================================================================

resource "time_sleep" "wait_sometime_test_app" {
  depends_on = [cloudfoundry_app.test_app,cloudfoundry_service_instance.simple_instance]

  create_duration = "60s"
}

# Bind test app to service instance
resource "cloudfoundry_service_credential_binding" "service-credential-binding" {
  type             = "app"
  name             = "hifi"
  service_instance = cloudfoundry_service_instance.simple_instance.id
  app              = cloudfoundry_app.test_app.id
  depends_on = [ time_sleep.wait_sometime_test_app ]
}

resource "cloudfoundry_network_policy" "policy" {
  policies = [
    {
      source_app      = cloudfoundry_app.test_app.id
      destination_app = cloudfoundry_app.service_broker.id
      port            = "61443"
      protocol        = "tcp"
    }
  ]
}

resource "cloudfoundry_org_quota" "org_quota" {
  name                     = "tf-test-do-not-delete"
  allow_paid_service_plans = true
  instance_memory          = 2048
  total_memory             = 51200
  total_app_instances      = 100
  total_routes             = 50
  total_services           = 200
  total_service_keys       = 120
  total_private_domains    = 40
  total_app_tasks          = 10
  total_route_ports        = 5
  total_app_log_rate_limit = 1000
  orgs = [
    cloudfoundry_org.test_org.id,
  ]
}

resource "cloudfoundry_security_group" "my_security_group" {
  name                     = "tf-test"
  globally_enabled_running = false
  globally_enabled_staging = false
  rules = [{
    protocol    = "tcp"
    destination = "192.168.1.100"
    ports       = "1883,8883"
    log         = true
    }]
  staging_spaces = [cloudfoundry_space.broker_space.id, cloudfoundry_space.test_space.id]
  running_spaces = [cloudfoundry_space.test_space.id]

}

resource "cloudfoundry_domain" "mydomain" {
  name        = "test-domain.apps.127-0-0-1.nip.io"
  org         = cloudfoundry_org.test_org.id
  shared_orgs = [cloudfoundry_space.broker_space.id, cloudfoundry_space.test_space.id]
}
