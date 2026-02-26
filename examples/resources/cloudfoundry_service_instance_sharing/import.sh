# terraform import cloudfoundry_service_instance_sharing.<resource_name> <service_instance_guid>

terraform import cloudfoundry_service_instance_sharing.my_instance_sharing a1b2c3d4-5678-90ab-cdef-12345678abcd

#terraform import using id attribute in import block

import {
  to = cloudfoundry_service_instance_sharing.<resource_name>
  id = "<service_instance_guid>"
}

# this resource supports import using identity attribute from Terraform version 1.12 or higher

import {
to = cloudfoundry_service_instance_sharing.<resource_name>
identity = {
  service_instance_guid = "<service_instance_guid>"
  }
}