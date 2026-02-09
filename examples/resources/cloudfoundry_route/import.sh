# terraform import cloudfoundry_route.<resource_name> <route_guid>

terraform import cloudfoundry_route.my_route 283f59d2-d660-45fb-9d96-b3e1aa92cfc7

import {
  to = cloudfoundry_route.<resource_name>
  id = "<route_guid>"
}

# this resource supports import using identity attribute from Terraform version 1.12 or higher

import {
to = cloudfoundry_route.<resource_name>
identity = {
  route_guid = "<route_guid>"
  }
}