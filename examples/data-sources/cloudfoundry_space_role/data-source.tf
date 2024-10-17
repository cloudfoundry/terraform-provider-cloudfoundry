data "cloudfoundry_space_role" "my_role" {
  id = "4c6849f2-6407-4385-a556-0840369f336b"
}

output "role_object" {
  value = data.cloudfoundry_space_role.my_role
}