package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAppResource_Configure(t *testing.T) {
	t.Parallel()
	resourceName := "cloudfoundry_app.app"
	params := `{"xsappname":"tf-test-app","tenant-mode":"dedicated","description":"tf test123","foreign-scope-references":["user_attributes"],"scopes":[{"name":"uaa.user","description":"UAA"}],"role-templates":[{"name":"Token_Exchange","description":"UAA","scope-references":["uaa.user"]}]}`
	t.Run("happy path - create app with bits", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_app_bits")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + `
resource "cloudfoundry_app" "app" {
	name                                 = "cf-nodejs"
  space_name                           = "tf-space-1" 
  org_name                             = "PerformanceTeamBLR"
  path                                 = "../../assets/cf-sample-app-nodejs.zip"
	memory                               = "256M"
	disk_quota                           = "1024mb"
	health_check_type                    = "http"
	health_check_http_endpoint           = "/"
	readiness_health_check_type          = "http"
	readiness_health_check_http_endpoint = "/"
  instances                            = 2
	service_bindings = [
    {
      service_instance : "xsuaa-tf"
      params = <<EOT
` + params + `
EOT
		}
	]
  environment = {
    MY_ENV = "red",
  }
  strategy = "rolling"
  routes = [
    {
      route = "cf-sample-test.cfapps.sap.hana.ondemand.com"
    }
  ]
}
					`,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "name", "cf-nodejs"),
						resource.TestCheckResourceAttr(resourceName, "space_name", "tf-space-1"),
						resource.TestCheckResourceAttr(resourceName, "org_name", "PerformanceTeamBLR"),
						resource.TestCheckResourceAttr(resourceName, "instances", "2"),
						resource.TestCheckResourceAttr(resourceName, "memory", "256M"),
						resource.TestCheckResourceAttr(resourceName, "disk_quota", "1024mb"),
						resource.TestCheckResourceAttr(resourceName, "health_check_type", "http"),
						resource.TestCheckResourceAttr(resourceName, "health_check_http_endpoint", "/"),
						resource.TestCheckResourceAttr(resourceName, "strategy", "rolling"),
						resource.TestCheckResourceAttr(resourceName, "service_bindings.0.service_instance", "xsuaa-tf"),
						resource.TestCheckResourceAttr(resourceName, "service_bindings.0.params", params+"\n"),
						resource.TestCheckResourceAttr(resourceName, "environment.MY_ENV", "red"),
						resource.TestCheckResourceAttr(resourceName, "routes.0.route", "cf-sample-test.cfapps.sap.hana.ondemand.com"),
						resource.TestCheckResourceAttr(resourceName, "routes.0.protocol", "http1"),
					),
				},
			},
		})
	})
	t.Run("happy path - update app with bits", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_app_bits_update")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + `
resource "cloudfoundry_app" "app" {
	name                                 = "cf-nodejs-update"
  space_name                           = "tf-space-1" 
  org_name                             = "PerformanceTeamBLR"
  path                                 = "../../assets/cf-sample-app-nodejs.zip"
	source_code_hash                     = "1234567890"
	memory                               = "0.5gb"
	disk_quota                           = "1024M"
  instances                            = 1
  environment = {
    MY_ENV = "red",
  }
	labels = {
		MY_LABEL = "red",
	}
  strategy = "blue-green"
}
					`,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "name", "cf-nodejs-update"),
						resource.TestCheckResourceAttr(resourceName, "space_name", "tf-space-1"),
						resource.TestCheckResourceAttr(resourceName, "org_name", "PerformanceTeamBLR"),
						resource.TestCheckResourceAttr(resourceName, "instances", "1"),
						resource.TestCheckResourceAttr(resourceName, "memory", "0.5gb"),
						resource.TestCheckResourceAttr(resourceName, "disk_quota", "1024M"),
						resource.TestCheckResourceAttr(resourceName, "strategy", "blue-green"),
						resource.TestCheckResourceAttr(resourceName, "environment.MY_ENV", "red"),
						resource.TestCheckResourceAttr(resourceName, "labels.MY_LABEL", "red"),
					),
				},
				{
					Config: hclProvider(nil) + `
resource "cloudfoundry_app" "app" {
	name                                 = "cf-nodejs-update"
  space_name                           = "tf-space-1" 
  org_name                             = "PerformanceTeamBLR"
  path                                 = "../../assets/cf-sample-app-nodejs.zip"
	source_code_hash                     = "999999"
	memory                               = "256M"
	disk_quota                           = "1024mB"
  instances                            = 2
  labels = {
		MY_LABEL = "blue",
	}
  strategy = "blue-green"
}
					`,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "name", "cf-nodejs-update"),
						resource.TestCheckResourceAttr(resourceName, "space_name", "tf-space-1"),
						resource.TestCheckResourceAttr(resourceName, "org_name", "PerformanceTeamBLR"),
						resource.TestCheckResourceAttr(resourceName, "instances", "2"),
						resource.TestCheckResourceAttr(resourceName, "memory", "256M"),
						resource.TestCheckResourceAttr(resourceName, "disk_quota", "1024mB"),
						resource.TestCheckResourceAttr(resourceName, "strategy", "blue-green"),
						resource.TestCheckResourceAttr(resourceName, "labels.MY_LABEL", "blue"),
					),
				},
			},
		})
	})

	t.Run("happy path - create app with docker and process", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_app_docker")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + `
resource "cloudfoundry_app" "app" {
	name         = "http-bin"
	space_name   = "tf-space-1"
	org_name     = "PerformanceTeamBLR"
	docker_image = "kennethreitz/httpbin"
	strategy		 = "blue-green"
	processes = [
		{
			type                                 = "web",
			instances                            = 1
			memory                               = "256M"
			disk_quota                           = "1024M"
			health_check_type                    = "http"
			health_check_http_endpoint           = "/get"
			readiness_health_check_type          = "http"
			readiness_health_check_http_endpoint = "/get"
		}
	]
	no_route = true
}
					`,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "docker_image", "kennethreitz/httpbin"),
						resource.TestCheckResourceAttr(resourceName, "strategy", "blue-green"),
						resource.TestCheckResourceAttr(resourceName, "no_route", "true"),
						resource.TestCheckResourceAttr(resourceName, "processes.0.instances", "1"),
						resource.TestCheckResourceAttr(resourceName, "processes.0.memory", "256M"),
						resource.TestCheckResourceAttr(resourceName, "processes.0.disk_quota", "1024M"),
						resource.TestCheckResourceAttr(resourceName, "processes.0.health_check_type", "http"),
						resource.TestCheckResourceAttr(resourceName, "processes.0.health_check_http_endpoint", "/get"),
						resource.TestCheckResourceAttr(resourceName, "processes.0.readiness_health_check_type", "http"),
						resource.TestCheckResourceAttr(resourceName, "processes.0.readiness_health_check_http_endpoint", "/get"),
						resource.TestCheckResourceAttr(resourceName, "processes.0.type", "web"),
					),
				},
			},
		})
	})
	t.Run("happy path - create app with sidecar", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_app_sidecar")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + `
resource "cloudfoundry_app" "http-bin-sidecar" {
	name         = "http-bin-sidecar"
	space_name   = "tf-space-1"
	org_name     = "PerformanceTeamBLR"
	docker_image = "kennethreitz/httpbin"
	sidecars = [
		{
			name         = "sidecar-1"
			process_types = [
				"worker"
			]
			command      = "sleep 5200"
			memory       = "256M"
		}
	]
	no_route = true
}
					`,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("cloudfoundry_app.http-bin-sidecar", "sidecars.0.name", "sidecar-1"),
						resource.TestCheckResourceAttr("cloudfoundry_app.http-bin-sidecar", "sidecars.0.command", "sleep 5200"),
						resource.TestCheckResourceAttr("cloudfoundry_app.http-bin-sidecar", "sidecars.0.process_types.#", "1"),
					),
				},
			},
		})
	})

	t.Run("happy path - create app with non-standard process type", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_app_scheduler")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + `
resource "cloudfoundry_app" "app" {
	name         = "http-bin-scheduler"
	space_name   = "tf-space-1"
	org_name     = "PerformanceTeamBLR"
	docker_image = "kennethreitz/httpbin"
	strategy		 = "blue-green"
	processes = [
		{
			type                                 = "web",
			instances                            = 1
			memory                               = "512M"
			disk_quota                           = "2048M"
			health_check_type                    = "port"
		},
		{
			type                                 = "scheduler",
			instances                            = 0
			memory                               = "256M"
			disk_quota                           = "1024M"
			health_check_type                    = "process"
		}
	]
	no_route = true
}
					`,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "docker_image", "kennethreitz/httpbin"),
						resource.TestCheckResourceAttr(resourceName, "strategy", "blue-green"),
						resource.TestCheckResourceAttr(resourceName, "no_route", "true"),
						resource.TestCheckResourceAttr(resourceName, "processes.#", "2"),
						// Check for the web process
						resource.TestCheckTypeSetElemNestedAttrs(resourceName, "processes.*", map[string]string{
							"type":              "web",
							"instances":         "1",
							"memory":            "512M",
							"disk_quota":        "2048M",
							"health_check_type": "port",
						}),
						// Check for the scheduler process (the main test goal)
						resource.TestCheckTypeSetElemNestedAttrs(resourceName, "processes.*", map[string]string{
							"type":              "scheduler",
							"instances":         "0",
							"memory":            "256M",
							"disk_quota":        "1024M",
							"health_check_type": "process",
						}),
					),
				},
			},
		})
	})
	// This test verifies that updating a space's allow_ssh attribute does not cause
	// apps in that space to be replaced. This is a regression test for an issue where
	// changing a referenced resource caused the app's space_name attribute to appear
	// as "known after apply", triggering unwanted replacement. The app's space_name
	// references cloudfoundry_space.test.name so that when the space is updated, the
	// name flows through as "known after apply" during planning. The UseStateForUnknown
	// plan modifier prevents this from triggering a replacement.
	t.Run("happy path - app not replaced when space allow_ssh changes", func(t *testing.T) {
		resourceName := "cloudfoundry_app.app_stability"
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_app_space_allow_ssh_update")
		defer stopQuietly(rec)

		var appID string

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					// Step 1: Create a space with allow_ssh=false and a docker app
					// that references cloudfoundry_space.test.name for space_name
					Config: hclProvider(nil) + `
resource "cloudfoundry_space" "test" {
	name      = "tf-space-1"
	org       = "` + testOrgGUID + `"
	allow_ssh = false
}
resource "cloudfoundry_app" "app_stability" {
	name         = "stability-test-app"
	space_name   = cloudfoundry_space.test.name
	org_name     = "PerformanceTeamBLR"
	docker_image = "kennethreitz/httpbin"
	no_route     = true
}
`,
					Check: resource.ComposeTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestCheckResourceAttr(resourceName, "name", "stability-test-app"),
						resource.TestCheckResourceAttr(resourceName, "space_name", "tf-space-1"),
						func(s *terraform.State) error {
							rs, ok := s.RootModule().Resources[resourceName]
							if !ok {
								return fmt.Errorf("resource not found: %s", resourceName)
							}
							appID = rs.Primary.ID
							return nil
						},
					),
				},
				{
					// Step 2: Change allow_ssh on the space. This causes the space
					// resource to be updated, making cloudfoundry_space.test.name
					// appear as "known after apply" during planning. Without
					// UseStateForUnknown on the app's space_name attribute,
					// this would trigger a replacement.
					Config: hclProvider(nil) + `
resource "cloudfoundry_space" "test" {
	name      = "tf-space-1"
	org       = "` + testOrgGUID + `"
	allow_ssh = true
}
resource "cloudfoundry_app" "app_stability" {
	name         = "stability-test-app"
	space_name   = cloudfoundry_space.test.name
	org_name     = "PerformanceTeamBLR"
	docker_image = "kennethreitz/httpbin"
	no_route     = true
}
`,
					Check: resource.ComposeTestCheckFunc(
						func(s *terraform.State) error {
							rs, ok := s.RootModule().Resources[resourceName]
							if !ok {
								return fmt.Errorf("resource not found: %s", resourceName)
							}
							if rs.Primary.ID != appID {
								return fmt.Errorf("app was unexpectedly replaced: old ID %s, new ID %s", appID, rs.Primary.ID)
							}
							return nil
						},
					),
				},
			},
		})
	})
}
