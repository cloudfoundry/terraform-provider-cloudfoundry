data "cloudfoundry_spaces" "spaces" {
  org = "784b4cd0-4771-4e4d-9052-a07e178bae56f"
}

output "id" {
  value = data.cloudfoundry_spaces.spaces
}