# terraform import cloudfoundry_security_group.<resource_name> <security_group_guid>

terraform import cloudfoundry_security_group.my_security_group 283f59d2-d660-45fb-9d96-b3e1aa92cfc7

#terraform import using id attribute in import block

import {
  to = cloudfoundry_security_group.<resource_name>
  id = "<security_group_guid>"
}

# this resource supports import using identity attribute from Terraform version 1.12 or higher

import {
to = cloudfoundry_security_group.<resource_name>
identity = {
  security_group_guid = "<security_group_guid>"
  }
}