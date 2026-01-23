# terraform import cloudfoundry_space.<resource_name> <space_guid>

terraform import cloudfoundry_space.my_space 283f59d2-d660-45fb-9d96-b3e1aa92cfc7

#terraform import using id attribute in import block

import {
  to = cloudfoundry_space.<resource_name>
  id = "<space_guid>"
}

# this resource supports import using identity attribute from Terraform version 1.12 or higher

import {
to = cloudfoundry_space.<resource_name>
identity = {
  space_guid = "<space_guid>"
  }
}