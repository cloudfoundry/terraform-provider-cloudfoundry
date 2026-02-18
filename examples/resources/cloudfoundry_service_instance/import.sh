# terraform import cloudfoundry_service_instance.<resource_name> <service_instance_guid>

terraform import cloudfoundry_service_instance.xsuaa_svc 68fea1b6-11b9-4737-ad79-74e49832533f

#terraform import using id attribute in import block

import {
  to = cloudfoundry_service_instance.<resource_name>
  id = "<service_instance_guid>"
}

# this resource supports import using identity attribute from Terraform version 1.12 or higher

import {
to = cloudfoundry_service_instance.<resource_name>
identity = {
  service_instance_guid = "<service_instance_guid>"
  }
}
