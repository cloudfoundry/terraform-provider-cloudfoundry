# terraform import cloudfoundry_org_quota.<resource_name> <org_quota_guid>

terraform import cloudfoundry_org_quota.my_org_quota e3cef997-9ba5-4cb4-b25b-c79faa81a33f

#terraform import using id attribute in import block

import {
  to = cloudfoundry_org_quota.<resource_name>
  id = "<org_quota_guid>"
}

# this resource supports import using identity attribute from Terraform version 1.12 or higher

import {
to = cloudfoundry_org_quota.<resource_name>
identity = {
  org_quota_guid = "<org_quota_guid>"
  }
}