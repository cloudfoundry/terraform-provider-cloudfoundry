resource "cloudfoundry_security_group_space_bindings" "my_security_group_spaces" {
  security_group = "56eedab7-cb97-469b-a3e9-89521827c039"
  staging_spaces = ["02c0cc92-6ecc-44b1-b7b2-096ca19ee143"]
  running_spaces = ["121c3a95-0f82-45a6-8ff2-1920b2067edb", "02c0cc92-6ecc-44b1-b7b2-096ca19ee143"]
}
