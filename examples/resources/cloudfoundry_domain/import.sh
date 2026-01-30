# terraform import cloudfoundry_domain.<resource_name> <domain_guid>

terraform import cloudfoundry_domain.my_domain e3cef997-9ba5-4cb4-b25b-c79faa81a33f

#terraform import using id attribute in import block

import {
  to = cloudfoundry_domain.<resource_name>
  id = "<domain_guid>"
}

# this resource supports import using identity attribute from Terraform version 1.12 or higher

import {
to = cloudfoundry_domain.<resource_name>
identity = {
  domain_guid = "<domain_guid>"
  }
}