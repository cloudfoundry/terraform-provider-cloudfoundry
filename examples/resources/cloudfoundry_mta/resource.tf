resource "cloudfoundry_mta" "mtar" {
  space                 = "02c0cc92-6ecc-44b1-b7b2-096ca19ee143"
  mtar_path             = "./my-mta_1.0.0.mtar"
  extension_descriptors = ["./prod.mtaext", "prod-scale-vertically.mtaext"]
  namespace             = "test"
  source_code_hash      = join("", [filesha256("./my-mta_1.0.0.mtar"), filesha256("./prod.mtaext"), filesha256("prod-scale-vertically.mtaext")])
  deploy_strategy       = "blue-green-deploy"
}
