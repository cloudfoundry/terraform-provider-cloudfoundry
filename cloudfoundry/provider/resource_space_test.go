package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
)

func TestSpaceResource_Configure(t *testing.T) {
	t.Parallel()
	testOrgGUID := "b4da43cd-2055-4d4d-ae6e-4066ce53a8b9"
	t.Run("happy path - create/read/update/delete space", func(t *testing.T) {
		resourceName := "cloudfoundry_space.ds"
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_space_crud")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclSpace(&SpaceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "ds",
						Name:          new("tf-unit-test"),
						OrgId:         new(testOrgGUID),
						AllowSSH:      new(true),
						Labels:        new(testCreateLabel),
						//IsolationSegment: strtostrptr(testIsolationSegmentGUID),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestCheckNoResourceAttr(resourceName, "quota"),
						resource.TestCheckResourceAttr(resourceName, "allow_ssh", "true"),
						resource.TestCheckResourceAttr(resourceName, "labels.purpose", "testing"),
						//resource.TestCheckResourceAttr(resourceName, "isolation_segment", testIsolationSegmentGUID),
					),
				},
				{
					Config: hclProvider(nil) + hclSpace(&SpaceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "ds",
						Name:          new("tf-unit-test"),
						OrgId:         new(testOrgGUID),
						AllowSSH:      new(false),
						Labels:        new(testUpdateLabel),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "allow_ssh", "false"),
						resource.TestCheckResourceAttr(resourceName, "labels.purpose", "production"),
						resource.TestCheckResourceAttr(resourceName, "labels.%", "2"),
						resource.TestCheckNoResourceAttr(resourceName, "isolation_segment"),
					),
				},
			},
		})
	})
	t.Run("happy path - import space", func(t *testing.T) {
		resourceName := "cloudfoundry_space.ds_import"
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_space_crud_import")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclSpace(&SpaceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "ds_import",
						Name:          new("tf-unit-test-import"),
						OrgId:         new(testOrgGUID),
						Labels:        new(testCreateLabel),
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

	t.Run("happy path - import space by resource identity", func(t *testing.T) {
		resourceName := "cloudfoundry_space.ds"
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_space_import")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclSpace(&SpaceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "ds",
						Name:          new("tf-unit-test"),
						OrgId:         new(testOrgGUID),
						AllowSSH:      new(true),
						Labels:        new(testCreateLabel),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestCheckNoResourceAttr(resourceName, "quota"),
						resource.TestCheckResourceAttr(resourceName, "allow_ssh", "true"),
						resource.TestCheckResourceAttr(resourceName, "labels.purpose", "testing"),
					),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectIdentity("cloudfoundry_space.ds", map[string]knownvalue.Check{
							"space_guid": knownvalue.NotNull(),
						}),
					},
				},
				{
					ResourceName:    "cloudfoundry_space.ds",
					ImportState:     true,
					ImportStateKind: resource.ImportBlockWithResourceIdentity,
				},
			},
		})
	})

	t.Run("error path - invalid isolation segment when creating space", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_space_invalid_isolation")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclSpace(&SpaceModelPtr{
						HclType:          hclObjectResource,
						HclObjectName:    "ds_isol",
						Name:             new("tf-unit-test123"),
						OrgId:            new(testOrgGUID),
						AllowSSH:         new(true),
						Labels:           new(testCreateLabel),
						IsolationSegment: new(invalidOrgGUID),
					}),
					ExpectError: regexp.MustCompile(`API Error Assigning Isolation Segment`),
				},
			},
		})
	})
	t.Run("error path - invalid organization when creating space", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_space_invalid_org")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclSpace(&SpaceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "ds_invalid_org",
						Name:          new("tf-unit-test"),
						OrgId:         new(invalidOrgGUID),
						AllowSSH:      new(true),
						Labels:        new(testCreateLabel),
					}),
					ExpectError: regexp.MustCompile(`API Error Creating Space`),
				},
			},
		})
	})
	t.Run("error path - invalid quota attribute", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_space_invalid_quota")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclSpace(&SpaceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "ds_invalid_attribute",
						Name:          new("tf-unit-test"),
						OrgId:         new(testOrgGUID),
						AllowSSH:      new(true),
						Labels:        new(testCreateLabel),
						Quota:         new(invalidOrgGUID),
					}),
					ExpectError: regexp.MustCompile(`Error: Invalid Configuration for Read-Only Attribute`),
				},
			},
		})
	})
}
