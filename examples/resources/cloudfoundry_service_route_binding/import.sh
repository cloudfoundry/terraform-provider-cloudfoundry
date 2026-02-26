# terraform import cloudfoundry_service_route_binding.<resource_name> <service_route_binding_guid>

terraform import cloudfoundry_service_route_binding.xsuaa_bi 68fea1b6-11b9-4737-ad79-74e49832533f

#terraform import using id attribute in import block

import {
  to = cloudfoundry_service_route_binding.<resource_name>
  id = "<service_route_binding_guid>"
}

# this resource supports import using identity attribute from Terraform version 1.12 or higher

import {
to = cloudfoundry_service_route_binding.<resource_name>
identity = {
  service_route_binding_guid = "<service_route_binding_guid>"
  }
}
