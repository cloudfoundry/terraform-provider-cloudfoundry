package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestSpaceResource_Configure(t *testing.T) {
	t.Parallel()
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
						HclType:          hclObjectResource,
						HclObjectName:    "ds",
						Name:             strtostrptr("tf-unit-test"),
						OrgId:            strtostrptr(testOrgGUID),
						AllowSSH:         booltoboolptr(true),
						Labels:           strtostrptr(testCreateLabel),
						IsolationSegment: strtostrptr(testIsolationSegmentGUID),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestCheckNoResourceAttr(resourceName, "quota"),
						resource.TestCheckResourceAttr(resourceName, "allow_ssh", "true"),
						resource.TestCheckResourceAttr(resourceName, "labels.purpose", "testing"),
						resource.TestCheckResourceAttr(resourceName, "isolation_segment", testIsolationSegmentGUID),
					),
				},
				{
					Config: hclProvider(nil) + hclSpace(&SpaceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "ds",
						Name:          strtostrptr("tf-unit-test"),
						OrgId:         strtostrptr(testOrgGUID),
						AllowSSH:      booltoboolptr(false),
						Labels:        strtostrptr(testUpdateLabel),
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
						Name:          strtostrptr("tf-unit-test-import"),
						OrgId:         strtostrptr(testOrgGUID),
						Labels:        strtostrptr(testCreateLabel),
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
						Name:             strtostrptr("tf-unit-test123"),
						OrgId:            strtostrptr(testOrgGUID),
						AllowSSH:         booltoboolptr(true),
						Labels:           strtostrptr(testCreateLabel),
						IsolationSegment: strtostrptr(invalidOrgGUID),
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
						Name:          strtostrptr("tf-unit-test"),
						OrgId:         strtostrptr(invalidOrgGUID),
						AllowSSH:      booltoboolptr(true),
						Labels:        strtostrptr(testCreateLabel),
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
						Name:          strtostrptr("tf-unit-test"),
						OrgId:         strtostrptr(testOrgGUID),
						AllowSSH:      booltoboolptr(true),
						Labels:        strtostrptr(testCreateLabel),
						Quota:         strtostrptr(invalidOrgGUID),
					}),
					ExpectError: regexp.MustCompile(`Error: Invalid Configuration for Read-Only Attribute`),
				},
			},
		})
	})
}
