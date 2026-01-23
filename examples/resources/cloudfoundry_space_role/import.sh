# terraform import cloudfoundry_space_role.<resource_name> <role_guid>

terraform import cloudfoundry_space_role.my_role 8f5b7b45-83f4-41c1-b99e-0f9582c31209

#terraform import using id attribute in import block

import {
  to = cloudfoundry_space_role.<resource_name>
  id = "<role_guid>"
}

# this resource supports import using identity attribute from Terraform version 1.12 or higher

import {
to = cloudfoundry_space_role.<resource_name>
identity = {
  role_guid = "<role_guid>"
  }
}