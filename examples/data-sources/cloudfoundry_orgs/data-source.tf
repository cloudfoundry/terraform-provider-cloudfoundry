data "cloudfoundry_orgs" "orgs" {
}

output "orgs" {
  value = data.cloudfoundry_orgs.orgs
}
