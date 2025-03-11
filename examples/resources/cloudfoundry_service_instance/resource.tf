data "cloudfoundry_org" "team_org" {
  name = "PerformanceTeamBLR"
}

data "cloudfoundry_space" "team_space" {
  name = "tf-space-1"
  org  = data.cloudfoundry_org.team_org.id
}

data "cloudfoundry_service_plan" "xsuaa_svc" {
  name                  = "application"
  service_offering_name = "xsuaa"
}
data "cloudfoundry_service_plan" "autoscaler_svc" {
  name                  = "standard"
  service_offering_name = "autoscaler"
}
resource "cloudfoundry_service_instance" "xsuaa_svc" {
  name         = "xsuaa_sv1"
  type         = "managed"
  tags         = ["terraform-test", "test1"]
  space        = data.cloudfoundry_space.team_space.id
  service_plan = data.cloudfoundry_service_plan.xsuaa_svc.service_plans.id
  parameters   = <<EOT
  {
  "xsappname": "tf-test2",
  "tenant-mode": "dedicated",
  "description": "tf test123",
  "foreign-scope-references": ["user_attributes"],
  "scopes": [
    {
      "name": "uaa.user",
      "description": "UAA"
    }
  ],
  "role-templates": [
    {
      "name": "Token_Exchange",
      "description": "UAA",
      "scope-references": ["uaa.user"]
    }
  ]
}
EOT
}

# Managed service instance without parameters
resource "cloudfoundry_service_instance" "dev-autoscaler" {
  name         = "tf-autoscaler-test"
  type         = "managed"
  tags         = ["terraform-test", "autoscaler"]
  space        = data.cloudfoundry_space.team_space.id
  service_plan = data.cloudfoundry_service_plan.autoscaler_svc.service_plans.id
  timeouts = {
    create = "10m"
  }
}
# User provided service instance
resource "cloudfoundry_service_instance" "dev-usp" {
  name        = "tf-usp-test1"
  type        = "user-provided"
  tags        = ["terraform-test", "usp"]
  space       = data.cloudfoundry_space.team_space.id
  credentials = <<EOT
  {
    "user": "user1",
    "password": "demo122"
  }
  EOT
}
