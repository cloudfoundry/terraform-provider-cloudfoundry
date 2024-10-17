data "cloudfoundry_mtas" "mtas" {
  space     = "02c0cc92-6ecc-44b1-b7b2-096ca19ee143"
  namespace = "test"
}

output "data" {
  value = data.cloudfoundry_mtas.mtas
}