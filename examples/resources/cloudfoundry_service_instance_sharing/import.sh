# terraform import cloudfoundry_service_instance_sharing.<resource_name> <service_instance_guid>

terraform import cloudfoundry_service_instance_sharing.my_instance_sharing a1b2c3d4-5678-90ab-cdef-12345678abcd

# For Terraform 1.5+ users, you can also use the newer import blocks syntax:

import {
  to = cloudfoundry_service_instance_sharing.my_instance_sharing
  id = "a1b2c3d4-5678-90ab-cdef-12345678abcd"
}