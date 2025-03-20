resource "cloudfoundry_user_cf" "my_user" {
  username = "tf-test1212123@example.com"
  origin   = "hi"
}

resource "cloudfoundry_org_role" "my_role" {
  user = cloudfoundry_user_cf.my_user.id
  type = "organization_user"
  org  = "db29f4b8-d39e-4f5c-b24d-cb34cde27abf"
} 