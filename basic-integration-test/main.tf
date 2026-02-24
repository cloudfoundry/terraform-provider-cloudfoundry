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
# 8. Zipper - Package test application source code into a zip file for deployment
# =============================================================================


resource "zipper_file" "fixture" {
  source      = "test-app"
  output_path = "test-app.zip"
}

# =============================================================================
# 10. Deploy Go Application 
# =============================================================================

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
    "description" = "Test GO application for E2E testing with service binding"
  }

  depends_on = [
    cloudfoundry_space.test_space,
    cloudfoundry_space_role.space_developer,
  ]
}