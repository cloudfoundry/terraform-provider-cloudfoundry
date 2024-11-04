package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAppsDataSource_Configure(t *testing.T) {
	t.Parallel()
	t.Run("happy path - read apps", func(t *testing.T) {
		cfg := getCFHomeConf()
		resourceName := "data.cloudfoundry_apps.apps"
		rec := cfg.SetupVCR(t, "fixtures/datasource_apps")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + `
					data "cloudfoundry_apps" "apps" {
						space = "02c0cc92-6ecc-44b1-b7b2-096ca19ee143"
					}`,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "apps.#", "8"),
					),
				},
				{
					Config: hclProvider(nil) + `
					data "cloudfoundry_apps" "apps" {
						space = "02c0cc92-6ecc-44b1-b7b2-096ca19ee143"
						name = "foo-hello-backend"
					}`,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "apps.#", "1"),
					),
				},
			},
		})
	})
	t.Run("error path - get unavailable apps", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_apps_invalid")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + `
						data "cloudfoundry_apps" "apps" {
							space = "02c0cc92-6ecc-44b1-b7b2-096ca19ee143"
							name = "blah"
						}`,
					ExpectError: regexp.MustCompile(`Unable to find any app in list`),
				},
			},
		})
	})
}
