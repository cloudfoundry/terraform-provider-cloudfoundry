# terraform import cloudfoundry_space_quota.<resource_name> <space_quota_guid>

terraform import cloudfoundry_space_quota.my_space_quota e3cef997-9ba5-4cb4-b25b-c79faa81a33f

#terraform import using id attribute in import block

import {
  to = cloudfoundry_space_quota.<resource_name>
  id = "<space_quota_guid>"
}

# this resource supports import using identity attribute from Terraform version 1.12 or higher

import {
to = cloudfoundry_space_quota.<resource_name>
identity = {
  space_quota_guid = "<space_quota_guid>"
  }
}