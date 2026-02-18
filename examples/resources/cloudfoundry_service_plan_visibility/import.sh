# terraform import cloudfoundry_service_plan_visibility.<resource_name> <service_plan_guid>

terraform import cloudfoundry_service_plan_visibility.test_visibility f37176d7-39eb-4e80-a3c0-328dfe36902c

#terraform import using id attribute in import block

import {
  to = cloudfoundry_service_plan_visibility.<resource_name>
  id = "<service_plan_guid>"
}

# this resource supports import using identity attribute from Terraform version 1.12 or higher

import {
to =cloudfoundry_service_plan_visibility.<resource_name>
identity = {
  service_plan_guid = "<service_plan_guid>"
  }
}