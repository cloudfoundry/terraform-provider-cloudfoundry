package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
)

func TestResourceSpaceQuota(t *testing.T) {
	t.Parallel()
	resourceName := "cloudfoundry_space_quota.rs"
	var testOrgGUID = "b4da43cd-2055-4d4d-ae6e-4066ce53a8b9"
	t.Run("happy path - create space quota", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_space_quota")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclSpaceQuota(&SpaceQuotaModelPtr{
						Name:                  new("tf-unit-test"),
						Org:                   new(testOrgGUID),
						AllowPaidServicePlans: new(true),
						HclType:               "resource",
						HclObjectName:         "rs",
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestMatchResourceAttr(resourceName, "updated_at", regexpValidRFC3999Format),
					),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectIdentity("cloudfoundry_space_quota.rs", map[string]knownvalue.Check{
							"space_quota_guid": knownvalue.NotNull(),
						}),
					},
				},
				{
					ResourceName:    "cloudfoundry_space_quota.rs",
					ImportState:     true,
					ImportStateKind: resource.ImportBlockWithResourceIdentity,
				},
			},
		})
	})
	t.Run("happy path - import space quota", func(t *testing.T) {
		resourceName := "cloudfoundry_space_quota.rs_import"
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_space_quota_import")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclSpaceQuota(&SpaceQuotaModelPtr{
						HclType:               hclObjectResource,
						HclObjectName:         "rs_import",
						Org:                   new(testOrgGUID),
						AllowPaidServicePlans: new(false),
						Name:                  new("tf-unit-test-import"),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
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
	t.Run("happy path - update space quota", func(t *testing.T) {
		resourceName := "cloudfoundry_space_quota.rs_update"
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_space_quota_update")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclSpaceQuota(&SpaceQuotaModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rs_update",
						Org:           new(testOrgGUID),
						TotalServices: new(10),
						TotalRoutes:   new(20),
						//TotalRoutePorts:       inttointptr(4),
						TotalAppTasks:         new(10),
						TotalServiceKeys:      new(10),
						AllowPaidServicePlans: new(false),
						Name:                  new("tf-unit-test-update"),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestCheckResourceAttr(resourceName, "total_services", "10"),
						resource.TestCheckResourceAttr(resourceName, "total_routes", "20"),
						//resource.TestCheckResourceAttr(resourceName, "total_route_ports", "4"),
						resource.TestCheckResourceAttr(resourceName, "total_app_tasks", "10"),
						resource.TestCheckResourceAttr(resourceName, "total_service_keys", "10"),
						resource.TestCheckResourceAttr(resourceName, "allow_paid_service_plans", "false"),
					),
				},
				{
					Config: hclProvider(nil) + hclSpaceQuota(&SpaceQuotaModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rs_update",
						Org:           new(testOrgGUID),
						TotalRoutes:   new(10),
						//TotalRoutePorts:       inttointptr(3),
						TotalAppTasks:         new(10),
						TotalServiceKeys:      new(10),
						AllowPaidServicePlans: new(true),
						Name:                  new("tf-unit-test-update"),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestCheckResourceAttr(resourceName, "total_routes", "10"),
						//resource.TestCheckResourceAttr(resourceName, "total_route_ports", "3"),
						resource.TestCheckResourceAttr(resourceName, "total_app_tasks", "10"),
						resource.TestCheckResourceAttr(resourceName, "total_service_keys", "10"),
						resource.TestCheckResourceAttr(resourceName, "allow_paid_service_plans", "true"),
					),
				},
			},
		})
	})
}
