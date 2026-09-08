package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAppResource_Lifecycle(t *testing.T) {
	t.Parallel()
	resourceName := "cloudfoundry_app.app"

	t.Run("happy path - create app with buildpack lifecycle", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_app_buildpack")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + `
resource "cloudfoundry_app" "app" {
	name           = "cf-nodejs-buildpack"
	space_name     = "tf-space-1"
	org_name       = "PerformanceTeamBLR"
	path           = "../../assets/cf-sample-app-nodejs.zip"
	lifecycle_type = "buildpack"
	buildpacks     = ["nodejs_buildpack"]
	memory         = "512M"
	instances      = 1
}
					`,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "name", "cf-nodejs-buildpack"),
						resource.TestCheckResourceAttr(resourceName, "lifecycle_type", "buildpack"),
						resource.TestCheckResourceAttr(resourceName, "buildpacks.0", "nodejs_buildpack"),
					),
				},
			},
		})
	})

	t.Run("happy path - create app with docker lifecycle", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_app_docker_lifecycle")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + `
resource "cloudfoundry_app" "app" {
	name           = "http-bin-lifecycle"
	space_name     = "tf-space-1"
	org_name       = "PerformanceTeamBLR"
	docker_image   = "kennethreitz/httpbin"
	lifecycle_type = "docker"
	no_route       = true
}
					`,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "docker_image", "kennethreitz/httpbin"),
						resource.TestCheckResourceAttr(resourceName, "lifecycle_type", "docker"),
					),
				},
			},
		})
	})

	t.Run("happy path - create app with cnb lifecycle", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_app_cnb")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + `
resource "cloudfoundry_app" "app" {
	name           = "cf-nodejs-cnb"
	space_name     = "tf-space-1"
	org_name       = "PerformanceTeamBLR"
	path           = "../../assets/cf-sample-app-nodejs.zip"
	lifecycle_type = "cnb"
	buildpacks     = ["docker://docker.io/paketobuildpacks/nodejs"]
	memory         = "512M"
	instances      = 1
}
					`,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "name", "cf-nodejs-cnb"),
						resource.TestCheckResourceAttr(resourceName, "lifecycle_type", "cnb"),
						resource.TestCheckResourceAttr(resourceName, "buildpacks.0", "docker://docker.io/paketobuildpacks/nodejs"),
					),
				},
			},
		})
	})

	t.Run("error path - docker_image conflicts with buildpack lifecycle_type", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_app_lifecycle_conflict")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + `
resource "cloudfoundry_app" "invalid" {
	name           = "invalid-lifecycle-buildpack-app"
	space_name     = "tf-space-1"
	org_name       = "PerformanceTeamBLR"
	docker_image   = "kennethreitz/httpbin"
	lifecycle_type = "buildpack"
}
					`,
					ExpectError: regexp.MustCompile(`lifecycle_type must be docker or omitted when docker_image is set`),
				},
			},
		})
	})

	t.Run("error path - docker_image conflicts with cnb lifecycle_type", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_app_lifecycle_conflict")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + `
resource "cloudfoundry_app" "invalid" {
	name           = "invalid-lifecycle-cnb-app"
	space_name     = "tf-space-1"
	org_name       = "PerformanceTeamBLR"
	docker_image   = "kennethreitz/httpbin"
	lifecycle_type = "cnb"
}
					`,
					ExpectError: regexp.MustCompile(`lifecycle_type must be docker or omitted when docker_image is set`),
				},
			},
		})
	})

	t.Run("error path - lifecycle_type must be a known value", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_app_lifecycle_invalid_value")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + `
resource "cloudfoundry_app" "invalid" {
	name           = "invalid-lifecycle-value-app"
	space_name     = "tf-space-1"
	org_name       = "PerformanceTeamBLR"
	path           = "../../assets/cf-sample-app-nodejs.zip"
	lifecycle_type = "not-a-real-lifecycle"
}
					`,
					ExpectError: regexp.MustCompile(`Attribute lifecycle_type value must be one of`),
				},
			},
		})
	})
}
