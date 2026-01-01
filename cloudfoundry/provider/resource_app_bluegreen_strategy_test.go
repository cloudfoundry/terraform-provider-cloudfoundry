package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAppResource_UpdateWithBlueGreen(t *testing.T) {
	t.Parallel()
	resourceName := "cloudfoundry_app.app"

	t.Run("happy path - update app with bits using blue green", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_app_blue_green")
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
  app_deployed_running_timeout = 1
  app_deployed_running_check_interval = 30
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
  app_deployed_running_timeout = 1
  app_deployed_running_check_interval = 30
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
}
