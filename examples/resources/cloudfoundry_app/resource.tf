terraform {
  required_providers {
    cloudfoundry = {
      source = "cloudfoundry/cloudfoundry"
    }
    zipper = {
      source = "ArthurHlt/zipper"
    }
  }
}
provider "cloudfoundry" {}

provider "zipper" {}

resource "zipper_file" "fixture" {
  source      = "https://github.com/cloudfoundry-samples/cf-sample-app-nodejs.git"
  output_path = "cf-sample-app-nodejs.zip"
}

resource "cloudfoundry_app" "gobis-server" {
  name             = "tf-test-do-not-delete-nodejs"
  space_name       = "tf-space-1"
  org_name         = "PerformanceTeamBLR"
  path             = zipper_file.fixture.output_path
  source_code_hash = zipper_file.fixture.output_sha
  instances        = 1
  environment = {
    MY_ENV = "red",
  }
  strategy = "rolling"
  service_bindings = [
    {
      service_instance : "xsuaa-tf"
      params = <<EOT
{
  "xsappname": "tf-test-app",
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
  ]
  routes = [
    {
      route = "tf-test-do-not-delete-nodejs.cfapps.sap.hana.ondemand.com"
    }
  ]
}

resource "cloudfoundry_app" "http-bin-server" {
  name                                = "tf-test-do-not-delete-http-bin"
  space_name                          = "tf-space-1"
  org_name                            = "PerformanceTeamBLR"
  docker_image                        = "kennethreitz/httpbin"
  strategy                            = "blue-green"
  app_deployed_running_timeout        = 2
  app_deployed_running_check_interval = 10
  labels = {
    "app" = "backend",
    "env" = "production"
  }
  annotations = {
    "created-by" = "me",
  }
  processes = [
    {
      type                                 = "web",
      instances                            = 1
      memory                               = "256M"
      disk_quota                           = "1024M"
      health_check_type                    = "http"
      health_check_http_endpoint           = "/get"
      readiness_health_check_type          = "http"
      readiness_health_check_http_endpoint = "/get"
    }
  ]
  no_route = true
}

resource "cloudfoundry_app" "http-bin-sidecar" {
  name         = "tf-test-do-not-delete-http-bin-sidecar"
  space_name   = "tf-space-1"
  org_name     = "PerformanceTeamBLR"
  docker_image = "kennethreitz/httpbin"
  sidecars = [
    {
      name = "sidecar-1"
      process_types = [
        "worker"
      ]
      command = "sleep 5200"
      memory  = "256M"
    }
  ]
  no_route = true
}