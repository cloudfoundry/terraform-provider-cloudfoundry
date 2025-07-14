package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestIsolationSegmentDataSource_Configure(t *testing.T) {
	var (
		// in staging
		isolationSegmentName = "hifi"
		resourceName         = "data.cloudfoundry_isolation_segment.ds"
	)
	t.Parallel()
	t.Run("get available datasource isolation segment", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_isolation_segment")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclIsolationSegment(&IsolationSegmentModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Name:          &isolationSegmentName,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestCheckResourceAttr(resourceName, "name", isolationSegmentName),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestMatchResourceAttr(resourceName, "updated_at", regexpValidRFC3999Format),
						resource.TestCheckResourceAttr(resourceName, "labels.purpose", "testing"),
					),
				},
			},
		})
	})
	t.Run("error path - get unavailable isolation segment", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_isolation_segment_invalid")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclIsolationSegment(&IsolationSegmentModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Name:          strtostrptr("testunavailable"),
					}),
					ExpectError: regexp.MustCompile(`Unable to find any Isolation Segment`),
				},
			},
		})
	})
}
