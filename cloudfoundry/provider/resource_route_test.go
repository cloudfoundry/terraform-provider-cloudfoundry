package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
)

func TestRouteResource_Configure(t *testing.T) {
	var (
		testSpace1RouteGUID  = "50655f6a-a66c-4276-b544-1d1aa864effd"
		testDomain1RouteGUID = "d042f463-bda6-4036-95a8-f2abfa7287ee"
	)
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
						Host:          new("testing123"),
						Path:          new("/cart"),
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
						Host:          new("testing123"),
						Path:          new("/cart"),
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
						Host:          new("testing123"),
						Path:          new("/cart"),
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

	t.Run("happy path - import with identity", func(t *testing.T) {
		resourceName := "cloudfoundry_route.ds"
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_route_import_identity")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclResourceRoute(&RouteResourceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "ds",
						Space:         &testSpace1RouteGUID,
						Domain:        &testDomain1RouteGUID,
					},
					),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
					),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectIdentity("cloudfoundry_route.ds", map[string]knownvalue.Check{
							"route_guid": knownvalue.NotNull(),
						}),
					},
				},
				{
					ResourceName:    resourceName,
					ImportState:     true,
					ImportStateKind: resource.ImportBlockWithResourceIdentity,
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
