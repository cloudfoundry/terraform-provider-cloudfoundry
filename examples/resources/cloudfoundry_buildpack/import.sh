# terraform import cloudfoundry_buildpack.<resource_name> <buildpack_guid>

terraform import cloudfoundry_buildpack.my_buildpack e3cef997-9ba5-4cb4-b25b-c79faa81a33f

#terraform import using id attribute in import block

import {
  to = cloudfoundry_buildpack.<resource_name>
  id = "<buildpack_guid>"
}

# this resource supports import using identity attribute from Terraform version 1.12 or higher

import {
to = cloudfoundry_buildpack.<resource_name>
identity = {
  buildpack_guid = "<buildpack_guid>"
  }
}