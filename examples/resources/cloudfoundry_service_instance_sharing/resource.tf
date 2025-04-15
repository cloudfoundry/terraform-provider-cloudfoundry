data "cloudfoundry_org" "my-org" {
  name = var.org_name
}

data "cloudfoundry_space" "current-space" {
  name = "current-space"
  org  = data.cloudfoundry_org.my-org-org.id
}

data "cloudfoundry_space" "space-to-share-with" {
  name = var.space_name
  org  = data.cloudfoundry_org.my-org.id
}

data "cloudfoundry_space" "another-space-to-share-with" {
  name = var.space_name
  org  = data.cloudfoundry_org.my-org.id
}

# existing service instance
data "cloudfoundry_service_instance" "svc" {
  name  = var.service_instance_name
  space = data.cloudfoundry_space.current-space.id
}

resource "cloudfoundry_service_instance_sharing" "instance_sharing" {
  service_instance = data.cloudfoundry_service_instance.svc.id
  spaces           = [data.cloudfoundry_space.space-to-share-with.id, data.cloudfoundry_space.another-space-to-share-with.id]
}