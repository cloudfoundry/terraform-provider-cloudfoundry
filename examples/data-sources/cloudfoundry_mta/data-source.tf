data "cloudfoundry_mta" "mta" {
  space = "02c0cc92-6ecc-44b1-b7b2-096ca19ee143"
  id    = "a.cf.app"
}

output "data" {
  value = data.cloudfoundry_mta.mta
}