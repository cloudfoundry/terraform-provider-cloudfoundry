data "cloudfoundry_service_brokers" "brokers" {
  space = "02c0cc92-6ecc-44b1-b7b2-096ca19ee143"
}

output "se_brs" {
  value = data.cloudfoundry_service_brokers.brokers
}
