data "cloudfoundry_buildpacks" "buildpack" {
  name = "java_buildpack"
}

output "buildpack" {
  value = data.cloudfoundry_buildpacks.buildpack
}