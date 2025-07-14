package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestIsolationSegmentEntitlementDataSource_Configure(t *testing.T) {
	var (
		// in staging
		resourceName = "data.cloudfoundry_isolation_segment_entitlement.ds"
		segmentGUID  = "63ae51b9-9073-4409-81b0-3704b8de85dd"
	)
	t.Parallel()
	t.Run("happy path - get available orgs entitled with isolation segment", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_isolation_segment_entilement")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclIsolationSegmentEntitlement(&IsolationSegmentEntitlementModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Segment:       &segmentGUID,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "segment", segmentGUID),
						resource.TestCheckResourceAttr(resourceName, "orgs.#", "1"),
					),
				},
			},
		})
	})
	t.Run("error path - get orgs entitled for invalid isolation segment", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_isolation_segment_entilement_invalid")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclIsolationSegmentEntitlement(&IsolationSegmentEntitlementModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Segment:       &invalidOrgGUID,
					}),
					ExpectError: regexp.MustCompile(`API Error Fetching Entitled Organizations`),
				},
			},
		})
	})
}
