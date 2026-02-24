# =============================================================================
# Data Sources - Fetch existing shared domain
# =============================================================================

data "cloudfoundry_domain" "default" {
  name = "127-0-0-1.nip.io"
}

# =============================================================================
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
# TODO: Service Broker Application deployment and configuration will go here in future iterations

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
}

# Organization Manager role for the first user
resource "cloudfoundry_org_role" "org_manager" {
  username = var.test_users[0].username
  type     = "organization_manager"
  org      = cloudfoundry_org.test_org.id
  origin   = var.test_users[0].origin

  depends_on = [cloudfoundry_org_role.org_users]
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
  username = var.test_users[0].username
  type     = "space_developer"
  space    = cloudfoundry_space.test_space.id
  origin   = var.test_users[0].origin

  depends_on = [cloudfoundry_org_role.org_users]
}

# =============================================================================
# 10. Deploy Go Application 
# =============================================================================

resource "cloudfoundry_app" "test_app" {
  name       = var.app_name
  org_name   = cloudfoundry_org.test_org.name
  space_name = cloudfoundry_space.test_space.name
  path       = "${path.module}/test-app"
  memory     = "64M"
  instances  = 1
  strategy   = "rolling"

  labels = {
    "purpose"     = "integration-test"
    "managed-by"  = "terraform"
    "environment" = "e2e-test"
  }
  annotations = {
    "description" = "Test GO application for E2E testing with service binding"
  }

  environment = {
    NODE_ENV   = "production"
    TEST_VAR   = "integration-test-value"
    CREATED_BY = "terraform-e2e-test"
  }

  processes = [
    {
      type       = "web"
      instances  = var.app_instances
      memory     = var.app_memory
      disk_quota = "512M"
    }
  ]

  depends_on = [
    cloudfoundry_space.test_space,
    cloudfoundry_space_role.space_developer,
    cloudfoundry_service_credential_binding.dummy_binding
  ]
}