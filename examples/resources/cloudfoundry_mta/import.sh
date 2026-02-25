# terraform import cloudfoundry_mta.<resource_name> <space_guid/mta_id> OR 
# terraform import cloudfoundry_mta.<resource_name> <space_guid/mta_id/namespace> if MTA in custom namespace

terraform import cloudfoundry_mta.my_mtar 02c0cc92-6ecc-44b1-b7b2-096ca19ee143/a.cf.app/hello

#terraform import using id attribute in import block

import {
  to = cloudfoundry_mta.<resource_name>
  id = "<space_guid/mta_id>" 
  # OR id = <space_guid/mta_id/namespace>
}

# this resource supports import using identity attribute from Terraform version 1.12 or higher

import {
to = cloudfoundry_mta.<resource_name>
identity = {
  space_guid = "<space_guid>"
  mta_id = "<mta_id>"
  namespace = "<namespace>"
  }
}

