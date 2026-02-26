# terraform import cloudfoundry_service_credential_binding.<resource_name> <service_credential_binding_guid>

terraform import cloudfoundry_service_credential_binding.xsuaa_svc 68fea1b6-11b9-4737-ad79-74e49832533f

#terraform import using id attribute in import block

import {
  to = cloudfoundry_service_credential_binding.<resource_name> 
  id = "<service_credential_binding_guid>"
}

# this resource supports import using identity attribute from Terraform version 1.12 or higher

import {
to = cloudfoundry_service_credential_binding.<resource_name>
identity = {
  service_credential_binding_guid = "<service_credential_binding_guid>"
  }
}
