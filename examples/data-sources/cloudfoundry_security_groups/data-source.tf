data "cloudfoundry_security_groups" "sgroups" {
  name = "riemann"
}

output "sgroup" {
  value = data.cloudfoundry_security_groups.sgroups
}