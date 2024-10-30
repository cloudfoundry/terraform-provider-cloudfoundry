data "cloudfoundry_space_roles" "roles" {
  space = "02c0cc92-6ecc-44b1-b7b2-096ca19ee143"
}

output "role_objects" {
  value = data.cloudfoundry_space_roles.roles
}