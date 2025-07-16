package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAppDataSource_Configure(t *testing.T) {
	t.Parallel()
	t.Run("happy path - read app with docker", func(t *testing.T) {
		cfg := getCFHomeConf()
		resourceName := "data.cloudfoundry_app.app"
		rec := cfg.SetupVCR(t, "fixtures/datasource_app_docker")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + `
data "cloudfoundry_app" "app" {
	name = "tf-test-do-not-delete-http-bin"
	space_name = "tf-space-1"
	org_name = "PerformanceTeamBLR"
}`,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "name", "tf-test-do-not-delete-http-bin"),
						resource.TestCheckResourceAttr(resourceName, "space_name", "tf-space-1"),
						resource.TestCheckResourceAttr(resourceName, "org_name", "PerformanceTeamBLR"),
						resource.TestCheckResourceAttr(resourceName, "processes.0.instances", "1"),
						resource.TestCheckResourceAttr(resourceName, "processes.0.memory", "256M"),
						resource.TestCheckResourceAttr(resourceName, "processes.0.disk_quota", "1024M"),
						resource.TestCheckResourceAttr(resourceName, "processes.0.health_check_type", "http"),
						resource.TestCheckResourceAttr(resourceName, "processes.0.health_check_http_endpoint", "/get"),
						resource.TestCheckResourceAttr(resourceName, "processes.0.readiness_health_check_type", "http"),
						resource.TestCheckResourceAttr(resourceName, "processes.0.readiness_health_check_http_endpoint", "/get"),
						resource.TestCheckResourceAttr(resourceName, "processes.0.type", "web"),
						resource.TestCheckResourceAttr(resourceName, "docker_image", "kennethreitz/httpbin"),
						resource.TestCheckResourceAttr(resourceName, "labels.%", "2"),
						resource.TestCheckResourceAttr(resourceName, "annotations.%", "1"),
					),
				},
			},
		})
	})
	t.Run("happy path - read app with bits", func(t *testing.T) {
		cfg := getCFHomeConf()
		resourceName := "data.cloudfoundry_app.app"
		rec := cfg.SetupVCR(t, "fixtures/datasource_app_bits")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + `
data "cloudfoundry_app" "app" {
	name = "tf-test-do-not-delete-nodejs"
	space_name = "tf-space-1"
	org_name = "PerformanceTeamBLR"
}`,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "name", "tf-test-do-not-delete-nodejs"),
						resource.TestCheckResourceAttr(resourceName, "space_name", "tf-space-1"),
						resource.TestCheckResourceAttr(resourceName, "org_name", "PerformanceTeamBLR"),
						resource.TestCheckResourceAttr(resourceName, "processes.0.instances", "1"),
						resource.TestCheckResourceAttr(resourceName, "processes.0.type", "web"),
						resource.TestCheckResourceAttr(resourceName, "service_bindings.#", "1"),
						resource.TestCheckResourceAttr(resourceName, "service_bindings.0.service_instance", "xsuaa-tf"),
						resource.TestCheckResourceAttr(resourceName, "routes.#", "1"),
						resource.TestCheckResourceAttr(resourceName, "routes.0.protocol", "http1"),
						resource.TestCheckResourceAttr(resourceName, "environment.MY_ENV", "red"),
					),
				},
			},
		})
	})
	t.Run("error path - get unavailable datasource app", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_app_invalid_app_name")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + `
data "cloudfoundry_app" "app" {
	name = "testunavailableapp"
	space_name = "tf-space-1"
	org_name = "PerformanceTeamBLR"
}`,
					ExpectError: regexp.MustCompile(`Error: Error finding given app`),
				},
			},
		})
	})
}

func TestAppDataSource_Lifecycle(t *testing.T) {
	t.Parallel()

	t.Run("read app with docker lifecycle", func(t *testing.T) {
		cfg := getCFHomeConf()
		resourceName := "data.cloudfoundry_app.app_lifecycle"
		rec := cfg.SetupVCR(t, "fixtures/datasource_app_lifecycle_docker")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + `
data "cloudfoundry_app" "app_lifecycle" {
	name = "tf-test-do-not-delete-http-bin"
	space_name = "tf-space-1"
	org_name = "PerformanceTeamBLR"
}`,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "name", "tf-test-do-not-delete-http-bin"),
						resource.TestCheckResourceAttr(resourceName, "app_lifecycle", "docker"),
						resource.TestCheckResourceAttr(resourceName, "docker_image", "kennethreitz/httpbin"),
					),
				},
			},
		})
	})

	t.Run("read app with buildpack lifecycle", func(t *testing.T) {
		cfg := getCFHomeConf()
		resourceName := "data.cloudfoundry_app.app_lifecycle"
		rec := cfg.SetupVCR(t, "fixtures/datasource_app_lifecycle_buildpack")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + `
data "cloudfoundry_app" "app_lifecycle" {
	name = "tf-test-do-not-delete-nodejs"
	space_name = "tf-space-1"
	org_name = "PerformanceTeamBLR"
}`,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "name", "tf-test-do-not-delete-nodejs"),
						resource.TestCheckResourceAttr(resourceName, "app_lifecycle", "buildpack"),
					),
				},
			},
		})
	})
}
