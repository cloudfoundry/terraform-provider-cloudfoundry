data "cloudfoundry_service_plan" "xsuaa-application" {
  service_offering_name = "xsuaa"
  name                  = "application"
}

output "serviceplan_id" {
  value = data.cloudfoundry_service_plan.xsuaa-application.id
}
