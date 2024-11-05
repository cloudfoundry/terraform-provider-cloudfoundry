data "cloudfoundry_service_route_bindings" "rbs" {
  service_instance = "ab65cad9-73fa-4dd4-9c09-87f89b2e77ec"
}

output "bindings" {
  value = data.cloudfoundry_service_route_bindings.rbs
}