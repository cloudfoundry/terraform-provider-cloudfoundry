data "cloudfoundry_org_role" "my_role" {
  id = "233959b9-c1fe-428d-8b53-5c903e5bd66b"
}

output "role_object" {
  value = data.cloudfoundry_space_role.my_role
}