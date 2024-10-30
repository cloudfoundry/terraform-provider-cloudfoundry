data "cloudfoundry_apps" "apps" {
  org = "02c0cc92-6ecc-44b1-b7b2-096ca19ee143"
}

output "apps" {
  value = data.cloudfoundry_apps.apps
}