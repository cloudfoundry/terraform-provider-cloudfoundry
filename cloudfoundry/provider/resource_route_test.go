package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestRouteResource_Configure(t *testing.T) {
	t.Parallel()
	t.Run("happy path - create/read/update/delete route", func(t *testing.T) {
		resourceName := "cloudfoundry_route.ds"
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_route_crud")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclResourceRoute(&RouteResourceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "ds",
						Space:         &testSpaceRouteGUID,
						Domain:        &testDomainRouteGUID,
						Host:          strtostrptr("testing123"),
						Path:          strtostrptr("/cart"),
						Destinations:  &createDestinations,
						Labels:        &testCreateLabel,
					},
					),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestCheckResourceAttr(resourceName, "protocol", "http"),
						resource.TestCheckResourceAttr(resourceName, "host", "testing123"),
						resource.TestCheckResourceAttr(resourceName, "path", "/cart"),
						resource.TestCheckNoResourceAttr(resourceName, "port"),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestMatchResourceAttr(resourceName, "updated_at", regexpValidRFC3999Format),
						resource.TestCheckResourceAttr(resourceName, "destinations.#", "2"),
						resource.TestCheckResourceAttr(resourceName, "labels.purpose", "testing"),
					),
				},
				{
					Config: hclProvider(nil) + hclResourceRoute(&RouteResourceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "ds",
						Space:         &testSpaceRouteGUID,
						Domain:        &testDomainRouteGUID,
						Host:          strtostrptr("testing123"),
						Path:          strtostrptr("/cart"),
						Destinations:  &updateDestinations1,
						Labels:        &testUpdateLabel,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "protocol", "http"),
						resource.TestCheckResourceAttr(resourceName, "host", "testing123"),
						resource.TestCheckResourceAttr(resourceName, "path", "/cart"),
						resource.TestCheckResourceAttr(resourceName, "destinations.#", "3"),
						resource.TestCheckResourceAttr(resourceName, "labels.purpose", "production"),
						resource.TestCheckResourceAttr(resourceName, "labels.%", "2"),
					),
				},
				{
					Config: hclProvider(nil) + hclResourceRoute(&RouteResourceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "ds",
						Space:         &testSpaceRouteGUID,
						Domain:        &testDomainRouteGUID,
						Host:          strtostrptr("testing123"),
						Path:          strtostrptr("/cart"),
						Destinations:  &updateDestinations2,
						Labels:        &testUpdateLabel,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "protocol", "http"),
						resource.TestCheckResourceAttr(resourceName, "host", "testing123"),
						resource.TestCheckResourceAttr(resourceName, "path", "/cart"),
						resource.TestCheckResourceAttr(resourceName, "destinations.#", "2"),
						resource.TestCheckResourceAttr(resourceName, "labels.purpose", "production"),
						resource.TestCheckResourceAttr(resourceName, "labels.%", "2"),
					),
				},
				{
					ResourceName:      resourceName,
					ImportStateIdFunc: getIdForImport(resourceName),
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})
	t.Run("happy path - route not replaced when space allow_ssh changes", func(t *testing.T) {
		testResourceNotReplacedOnSpaceUpdate(t,
			"fixtures/resource_route_space_allow_ssh_update",
			"cloudfoundry_route.rt_stability",
			func(allowSSH bool) string {
				return hclSpaceWithSSH("test-space-route-stability", allowSSH) + `
resource "cloudfoundry_route" "rt_stability" {
	space  = cloudfoundry_space.test.id
	domain = "` + testDomainRouteGUID + `"
	host   = "stability-test"
}
`
			},
		)
	})
	t.Run("error path - invalid domain or space when creating route", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_route_invalid")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclResourceRoute(&RouteResourceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "ds_invalid_name",
						Space:         &testSpaceRouteGUID,
						Domain:        &testSpaceRouteGUID,
					}),
					ExpectError: regexp.MustCompile(`API Error Creating Route`),
				},
			},
		})
	})

}
