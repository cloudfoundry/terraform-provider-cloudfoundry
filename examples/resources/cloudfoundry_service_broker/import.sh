# terraform import cloudfoundry_service_broker.<resource_name> <service_broker_guid>

terraform import cloudfoundry_service_broker.service_broker 283f59d2-d660-45fb-9d96-b3e1aa92cfc7

#terraform import using id attribute in import block

import {
  to = cloudfoundry_service_broker.<resource_name>
  id = "<service_broker_guid>"
}

# this resource supports import using identity attribute from Terraform version 1.12 or higher

import {
to = cloudfoundry_service_broker.<resource_name>
identity = {
  service_broker_guid = "<service_broker_guid>"
  }
}