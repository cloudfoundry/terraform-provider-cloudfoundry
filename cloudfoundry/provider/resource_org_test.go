package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestResourceOrg(t *testing.T) {
	t.Parallel()
	t.Run("happy path - create/update/delete/import org", func(t *testing.T) {
		cfg := getCFHomeConf()
		resourceName := "cloudfoundry_org.crud_org"
		rec := cfg.SetupVCR(t, "fixtures/resource_org")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclOrg(&OrgModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "crud_org",
						Name:          strtostrptr("tf-unit-test"),
						Labels:        strtostrptr(testCreateLabel),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestMatchResourceAttr(resourceName, "updated_at", regexpValidRFC3999Format),
						resource.TestMatchResourceAttr(resourceName, "quota", regexpValidUUID),
						resource.TestCheckResourceAttr(resourceName, "labels.purpose", "testing"),
					),
				},
				{
					Config: hclProvider(nil) + hclOrg(&OrgModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "crud_org",
						Name:          strtostrptr("tf-org-test"),
						Labels:        strtostrptr(testUpdateLabel),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestMatchResourceAttr(resourceName, "updated_at", regexpValidRFC3999Format),
						resource.TestMatchResourceAttr(resourceName, "quota", regexpValidUUID),
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

	t.Run("happy path - create suspended org ", func(t *testing.T) {
		cfg := getCFHomeConf()
		resourceName := "cloudfoundry_org.suspended_org"
		rec := cfg.SetupVCR(t, "fixtures/resource_org_suspended")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclOrg(&OrgModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "suspended_org",
						Name:          strtostrptr("tf-org-suspended-test"),
						Labels:        strtostrptr(testCreateLabel),
						Suspended:     booltoboolptr(true),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestMatchResourceAttr(resourceName, "updated_at", regexpValidRFC3999Format),
						resource.TestMatchResourceAttr(resourceName, "quota", regexpValidUUID),
						resource.TestCheckResourceAttr(resourceName, "labels.purpose", "testing"),
						resource.TestCheckResourceAttr(resourceName, "suspended", "true"),
					),
				},
			},
		})

	})

}
