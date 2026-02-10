# terraform import cloudfoundry_user_cf.<resource_name> <user_guid>

terraform import cloudfoundry_user_cf.my_user 283f59d2-d660-45fb-9d96-b3e1aa92cfc7

#terraform import using id attribute in import block

import {
  to = cloudfoundry_user_cf.<resource_name>
  id = "<user_guid>"
}

# this resource supports import using identity attribute from Terraform version 1.12 or higher

import {
to = cloudfoundry_user_cf.<resource_name>
identity = {
  user_guid = "<user_guid>"
  }
}