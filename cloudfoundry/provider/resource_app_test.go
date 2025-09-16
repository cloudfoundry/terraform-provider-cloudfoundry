package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAppResource_Configure(t *testing.T) {
	t.Parallel()
	resourceName := "cloudfoundry_app.app"
	params := `{"xsappname":"tf-test-app","tenant-mode":"dedicated","description":"tf test123","foreign-scope-references":["user_attributes"],"scopes":[{"name":"UAA","description":"UAA"}],"role-templates":[{"name":"Token_Exchange","description":"UAA","scope-references":["uaa.user"]}]}`
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
}

func TestAppResource_Lifecycle(t *testing.T) {
	t.Parallel()

	t.Run("buildpack lifecycle", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_app_lifecycle_buildpack")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + `
resource "cloudfoundry_app" "app_buildpack" {
	name                                 = "cf-buildpack-lifecycle"
    space_name                           = "tf-space-1" 
    org_name                             = "PerformanceTeamBLR"
    path                                 = "../../assets/cf-sample-app-nodejs.zip"
	app_lifecycle                        = "buildpack"
	memory                               = "256M"
	disk_quota                           = "1024M"
    instances                            = 1
}
					`,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("cloudfoundry_app.app_buildpack", "name", "cf-buildpack-lifecycle"),
						resource.TestCheckResourceAttr("cloudfoundry_app.app_buildpack", "app_lifecycle", "buildpack"),
					),
				},
			},
		})
	})

	t.Run("docker lifecycle", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_app_lifecycle_docker")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + `
resource "cloudfoundry_app" "app_docker" {
	name         = "cf-docker-lifecycle"
	space_name   = "tf-space-1"
	org_name     = "PerformanceTeamBLR"
	docker_image = "kennethreitz/httpbin"
	app_lifecycle = "docker"
	no_route     = true
}
					`,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("cloudfoundry_app.app_docker", "name", "cf-docker-lifecycle"),
						resource.TestCheckResourceAttr("cloudfoundry_app.app_docker", "app_lifecycle", "docker"),
					),
				},
			},
		})
	})

	t.Run("cnb lifecycle", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_app_lifecycle_cnb")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + `
resource "cloudfoundry_app" "app_cnb" {
	name                                 = "cf-cnb-lifecycle"
    space_name                           = "tf-space-1" 
    org_name                             = "PerformanceTeamBLR"
	app_lifecycle                        = "cnb"
	buildpacks                           = ["docker://docker.io/paketobuildpacks/nodejs"]
	stack                                = "cflinuxfs4"
	memory                               = "512M"
	disk_quota                           = "1024M"
    instances                            = 1
    random_route                         = true
}
					`,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("cloudfoundry_app.app_cnb", "name", "cf-cnb-lifecycle"),
						resource.TestCheckResourceAttr("cloudfoundry_app.app_cnb", "app_lifecycle", "cnb"),
						resource.TestCheckResourceAttr("cloudfoundry_app.app_cnb", "buildpacks.0", "docker://docker.io/paketobuildpacks/nodejs"),
					),
				},
			},
		})
	})
}
