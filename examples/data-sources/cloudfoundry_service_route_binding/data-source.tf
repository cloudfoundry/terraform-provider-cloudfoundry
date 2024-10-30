data "cloudfoundry_service_route_binding" "rb" {
  id = "ab65cad9-73fa-4dd4-9c09-87f89b2e77ec"
}

output "binding" {
  value = data.cloudfoundry_service_route_binding.rb
}