package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestServiceRouteBindingDataSource(t *testing.T) {
	serviceRouteBinding := "6961bc50-a694-4255-a5da-3ce4cdba7e54"
	t.Parallel()
	t.Run("happy path - read route binding", func(t *testing.T) {
		cfg := getCFHomeConf()
		dataSourceName := "data.cloudfoundry_service_route_binding.ds"
		rec := cfg.SetupVCR(t, "fixtures/datasource_service_route_binding")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclServiceRouteBinding(&ServiceRouteBindingModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Id:            &serviceRouteBinding,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(dataSourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(dataSourceName, "service_instance", regexpValidUUID),
						resource.TestMatchResourceAttr(dataSourceName, "route", regexpValidUUID),
						resource.TestMatchResourceAttr(dataSourceName, "created_at", regexpValidRFC3999Format),
						resource.TestMatchResourceAttr(dataSourceName, "updated_at", regexpValidRFC3999Format),
					),
				},
			},
		})
	})
	t.Run("error path - get unavailable route binding", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_service_route_binding_invalid")
		defer stopQuietly(rec)
		resource.UnitTest(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclServiceRouteBinding(&ServiceRouteBindingModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Id:            &invalidOrgGUID,
					}),
					ExpectError: regexp.MustCompile(`API Error Fetching Service Route Binding`),
				},
			},
		})
	})
}
