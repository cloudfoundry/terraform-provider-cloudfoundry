# terraform import cloudfoundry_org.<resource_name> <org_guid>

terraform import cloudfoundry_org.my_org e3cef997-9ba5-4cb4-b25b-c79faa81a33f

#terraform import using id attribute in import block

import {
  to = cloudfoundry_org.<resource_name>
  id = "<org_guid>"
}

# this resource supports import using identity attribute from Terraform version 1.12 or higher

import {
to = cloudfoundry_org.<resource_name>
identity = {
  org_guid = "<org_guid>"
  }
}