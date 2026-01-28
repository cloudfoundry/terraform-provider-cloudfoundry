# terraform import cloudfoundry_isolation_segment.<resource_name> <segment_guid>

terraform import cloudfoundry_isolation_segment.my_segment e3cef997-9ba5-4cb4-b25b-c79faa81a33f

#terraform import using id attribute in import block

import {
  to = cloudfoundry_isolation_segment.<resource_name> 
  id = "<segment_guid>"
}

# this resource supports import using identity attribute from Terraform version 1.12 or higher

import {
to = cloudfoundry_isolation_segment.<resource_name>
identity = {
  segment_guid = "<segment_guid>"
  }
}