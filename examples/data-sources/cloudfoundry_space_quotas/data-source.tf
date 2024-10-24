data "cloudfoundry_space_quotas" "my_space_quotas" {
  org = "784b4cd0-4771-4e4d-9052-a07e178bae56"
}

output "quotas" {
  value = data.cloudfoundry_space_quotas.my_space_quotas
}