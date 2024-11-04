data "cloudfoundry_service_instances" "svcs" {
  org = "784b4cd0-4771-4e4d-9052-a07e178bae56"
}

output "svcs" {
  value = data.cloudfoundry_service_instances.svcs
}
