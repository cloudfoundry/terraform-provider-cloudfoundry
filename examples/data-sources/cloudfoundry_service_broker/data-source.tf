data "cloudfoundry_service_broker" "broker" {
  name = "hi"
}

output "se_br" {
  value = data.cloudfoundry_service_broker.broker
}
