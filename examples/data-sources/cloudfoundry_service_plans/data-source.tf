data "cloudfoundry_service_plans" "xsuaa-offering" {
  service_offering_name = "xsuaa"
}

output "serviceplans" {
  value = data.cloudfoundry_service_plans.xsuaa-offering.service_plans
}