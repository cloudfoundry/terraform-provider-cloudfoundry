# terraform import cloudfoundry_org_role.<resource_name> <role_guid>

terraform import cloudfoundry_org_role.my_role e17839d9-cd4f-4e4b-baf0-18786f12fede

#terraform import using id attribute in import block

import {
  to = cloudfoundry_org_role.<resource_name>
  id = "<role_guid>"
}

# this resource supports import using identity attribute from Terraform version 1.12 or higher

import {
to = cloudfoundry_org_role.<resource_name>
identity = {
  role_guid = "<role_guid>"
  }
}