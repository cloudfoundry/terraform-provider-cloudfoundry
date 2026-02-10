package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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
	// This test verifies that updating a space's allow_ssh attribute does not cause
	// routes in that space to be replaced. This is a regression test for an issue
	// where changing space attributes caused the route's space reference to appear
	// as "known after apply", triggering unwanted replacement. The route references
	// cloudfoundry_space.test.id so that when the space is updated, the space ID
	// flows through as "known after apply" during planning. The UseStateForUnknown
	// plan modifier prevents this from triggering a replacement.
	t.Run("happy path - route not replaced when space allow_ssh changes", func(t *testing.T) {
		resourceName := "cloudfoundry_route.rt_stability"
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_route_space_allow_ssh_update")
		defer stopQuietly(rec)

		var routeID string

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					// Step 1: Create a space with allow_ssh=false and a route
					// that references cloudfoundry_space.test.id
					Config: hclProvider(nil) + `
resource "cloudfoundry_space" "test" {
	name      = "test-space-route-stability"
	org       = "` + testOrgGUID + `"
	allow_ssh = false
}
resource "cloudfoundry_route" "rt_stability" {
	space  = cloudfoundry_space.test.id
	domain = "` + testDomainRouteGUID + `"
	host   = "stability-test"
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestCheckResourceAttr(resourceName, "host", "stability-test"),
						func(s *terraform.State) error {
							rs, ok := s.RootModule().Resources[resourceName]
							if !ok {
								return fmt.Errorf("resource not found: %s", resourceName)
							}
							routeID = rs.Primary.ID
							return nil
						},
					),
				},
				{
					// Step 2: Change allow_ssh on the space. This causes the space
					// resource to be updated, making cloudfoundry_space.test.id
					// appear as "known after apply" during planning. Without
					// UseStateForUnknown on the route's space attribute,
					// this would trigger a replacement.
					Config: hclProvider(nil) + `
resource "cloudfoundry_space" "test" {
	name      = "test-space-route-stability"
	org       = "` + testOrgGUID + `"
	allow_ssh = true
}
resource "cloudfoundry_route" "rt_stability" {
	space  = cloudfoundry_space.test.id
	domain = "` + testDomainRouteGUID + `"
	host   = "stability-test"
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						func(s *terraform.State) error {
							rs, ok := s.RootModule().Resources[resourceName]
							if !ok {
								return fmt.Errorf("resource not found: %s", resourceName)
							}
							if rs.Primary.ID != routeID {
								return fmt.Errorf("route was unexpectedly replaced: old ID %s, new ID %s", routeID, rs.Primary.ID)
							}
							return nil
						},
					),
				},
			},
		})
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
