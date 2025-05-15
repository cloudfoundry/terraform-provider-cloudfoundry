resource "cloudfoundry_service_plan_visibility" "test_visibility" {
  service_plan  = "f37176d7-39eb-4e80-a3c0-328dfe36902c"
  organizations = ["3533be5d-272f-42fe-bf70-fc4b108c2043"]
  type          = "organization"
}