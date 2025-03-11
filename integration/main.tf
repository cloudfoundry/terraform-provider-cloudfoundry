resource "cloudfoundry_org" "org" {
  name      = "tf-test-iso"
  suspended = false
  labels = {
    env = "test"
  }
  annotations = {
    env-ann = "test-ann"
  }
}

resource "cloudfoundry_space" "space" {
  name      = "space"
  org       = cloudfoundry_org.org.id
  allow_ssh = "true"
  labels    = { test : "pass", purpose : "prod" }
}

resource "cloudfoundry_app" "http-bin-sidecar" {
  name         = "cf-app-1"
  space_name   = cloudfoundry_space.space.name
  org_name     = cloudfoundry_org.org.name
  docker_image = "kennethreitz/httpbin"
#   sidecars = [
#     {
#       name = "sidecar-1"
#       process_types = [
#         "worker"
#       ]
#       command = "sleep 5200"
#       memory  = "256M"
#     }
#   ]
#   no_route = true
}