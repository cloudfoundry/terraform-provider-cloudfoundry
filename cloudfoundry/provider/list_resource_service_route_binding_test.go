package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/querycheck/queryfilter"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

const testServiceRouteBindingListServiceInstanceGUID = "ab65cad9-73fa-4dd4-9c09-87f89b2e77ec"

func TestServiceRouteBindingListResource(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/list_resource_service_route_binding")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			TerraformVersionChecks: []tfversion.TerraformVersionCheck{
				tfversion.SkipBelow(tfversion.Version1_14_0),
			},
			Steps: []resource.TestStep{
				{
					Query: true,
					Config: hclProvider(nil) + listServiceRouteBindingQueryConfig(
						"binding_list",
						"cloudfoundry",
						testServiceRouteBindingListServiceInstanceGUID,
					),
					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("cloudfoundry_service_route_binding.binding_list", 1),
						querycheck.ExpectIdentity(
							"cloudfoundry_service_route_binding.binding_list",
							map[string]knownvalue.Check{
								"service_route_binding_guid": knownvalue.StringRegexp(regexpValidUUID),
							},
						),
					},
				},
				{
					// List Query with include_resource = true
					Query: true,
					Config: hclProvider(nil) + listServiceRouteBindingQueryConfigWithIncludeResource(
						"binding_list",
						"cloudfoundry",
						testServiceRouteBindingListServiceInstanceGUID,
					),
					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("cloudfoundry_service_route_binding.binding_list", 1),
						querycheck.ExpectResourceKnownValues(
							"cloudfoundry_service_route_binding.binding_list",
							queryfilter.ByResourceIdentity(map[string]knownvalue.Check{
								"service_route_binding_guid": knownvalue.StringExact("6961bc50-a694-4255-a5da-3ce4cdba7e54"),
							}),
							[]querycheck.KnownValueCheck{
								{
									Path:       tfjsonpath.New("service_instance"),
									KnownValue: knownvalue.StringExact(testServiceRouteBindingListServiceInstanceGUID),
								},
								{
									Path:       tfjsonpath.New("route"),
									KnownValue: knownvalue.StringExact("8c7cdbd4-98ef-4e2f-97d1-4704f435022d"),
								},
								{
									Path:       tfjsonpath.New("route_service_url"),
									KnownValue: knownvalue.StringExact("https://nginx-route-service.cfapps.sap.hana.ondemand.com"),
								},
							},
						),
					},
				},
			},
		})
	})
}

func listServiceRouteBindingQueryConfig(label, providerName, serviceInstanceGUID string) string {
	return fmt.Sprintf(`
list "cloudfoundry_service_route_binding" "%s" {
  provider = %s
  config {
    service_instance = "%s"
  }
}`, label, providerName, serviceInstanceGUID)
}

func listServiceRouteBindingQueryConfigWithIncludeResource(label, providerName, serviceInstanceGUID string) string {
	return fmt.Sprintf(`
list "cloudfoundry_service_route_binding" "%s" {
  provider         = %s
  include_resource = true
  config {
    service_instance = "%s"
  }
}`, label, providerName, serviceInstanceGUID)
}
