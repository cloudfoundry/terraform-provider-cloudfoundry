resource "cloudfoundry_mta" "mtaone" {
  space                 = "02c0cc92-6ecc-44b1-b7b2-096ca19ee143"
  mtar_path             = "./my-mta_1.0.0.mtar"
  extension_descriptors = ["./prod.mtaext", "prod-scale-vertically.mtaext"]
  namespace             = "test"
  source_code_hash      = join("", [filesha256("./my-mta_1.0.0.mtar"), filesha256("./prod.mtaext"), filesha256("prod-scale-vertically.mtaext")])
  deploy_strategy       = "blue-green-deploy"
}

resource "cloudfoundry_mta" "mtatwo" {
  space     = "02c0cc92-6ecc-44b1-b7b2-096ca19ee143"
  mtar_path = "./my-mta_1.0.0.mtar"
  extension_descriptors_string = [
    <<EOT
_schema-version: 3.3.0
ID: my-mta-prod
extends: my-mta
version: 1.0.0

modules:
- name: my-app
  parameters:
    instances: 2

resources:
 - name: my-service
   parameters:
     service-plan: "lite"
EOT
    ,
    <<EOT
_schema-version: 3.3.0
ID: my-mta-prod-scale-vertically
extends: my-mta-prod
version: 1.0.0

modules:
- name: my-app
  parameters:
    memory: 2G
EOT  
  ]
  namespace        = "test"
  source_code_hash = filesha256("./my-mta_1.0.0.mtar")
  deploy_strategy  = "deploy"
}
