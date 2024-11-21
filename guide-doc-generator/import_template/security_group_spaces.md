# Application Security Group Spaces


-----------------
#RES.DESC
Application Security Group `asg`  is now identified as `security_group`. The newer resource also exposes some additional parameters that are introduced as part of the v3 specification.
##RES.DESC

------------------
#RES.COMM
resource "cloudfoundry_space_asgs" "spaceasgs" {
    space        = "e4ccb84e-5d8b-4ca2-b59a-012f4cf45c5d"
    running_asgs = ["38109de6-8059-41dd-b9fa-d802d8a35271"]
    staging_asgs = ["531dd667-0fcf-44a0-8c6a-a541a062750d"]
}
##RES.COMM

--------------------
#RES.SAP
resource "cloudfoundry_security_group_spaces" "space_bindings1" {
  security_group = "38109de6-8059-41dd-b9fa-d802d8a35271"
  running_spaces = ["e4ccb84e-5d8b-4ca2-b59a-012f4cf45c5d"]
}

resource "cloudfoundry_security_group_spaces" "space_bindings2" {
  security_group = "531dd667-0fcf-44a0-8c6a-a541a062750d"
  staging_spaces = ["e4ccb84e-5d8b-4ca2-b59a-012f4cf45c5d"]
}
##RES.SAP

---------------

#DS.DESC
##DS.DESC
----------------

#DS.COMM
##DS.COMM
-----------------

#DS.SAP
##DS.SAP