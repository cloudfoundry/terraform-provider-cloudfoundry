# terraform import cloudfoundry_app.<resource_name> <app_guid>

terraform import 'cloudfoundry_app.gobis-server' f71f4a6e-253c-4025-8e45-d2be1a0d9b15

#terraform import using id attribute in import block

import {
  to = cloudfoundry_app.<resource_name>
  id = "<app_guid>"
}