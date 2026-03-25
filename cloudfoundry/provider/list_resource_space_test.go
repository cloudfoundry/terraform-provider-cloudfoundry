package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/list"
	res "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/querycheck/queryfilter"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestSpaceListResource(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/list_resource_space")
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
					Config: hclProvider(nil) + listSpaceQueryConfig(
						"space_list",
						"cloudfoundry",
						"7c8b9705-bc7c-4a96-9af0-6674c8290615",
					),

					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("cloudfoundry_space.space_list", 1),

						querycheck.ExpectIdentity(
							"cloudfoundry_space.space_list",
							map[string]knownvalue.Check{
								"space_guid": knownvalue.StringRegexp(regexpValidUUID),
							},
						),
					},
				},
				{
					// List Query with include_resource = true
					Query: true,
					Config: hclProvider(nil) + listSpaceQueryConfigWithIncludeResource(
						"space_list",
						"cloudfoundry",
						"7c8b9705-bc7c-4a96-9af0-6674c8290615",
					),

					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("cloudfoundry_space.space_list", 1),

						// Verify full resource data is populated
						querycheck.ExpectResourceKnownValues(
							"cloudfoundry_space.space_list",
							queryfilter.ByResourceIdentity(map[string]knownvalue.Check{
								"space_guid": knownvalue.StringExact("10b8f9bd-0731-49f3-9e88-a17b6eec090e"),
							}),
							[]querycheck.KnownValueCheck{
								{
									Path:       tfjsonpath.New("name"),
									KnownValue: knownvalue.StringExact("test"),
								},
								{
									Path:       tfjsonpath.New("id"),
									KnownValue: knownvalue.StringRegexp(regexpValidUUID),
								},
								{
									Path:       tfjsonpath.New("org"),
									KnownValue: knownvalue.StringExact("7c8b9705-bc7c-4a96-9af0-6674c8290615"),
								},
							},
						),
					},
				},
			},
		})
	})

	t.Run("error path - configure", func(t *testing.T) {
		r, ok := NewSpaceListResource().(list.ListResourceWithConfigure)
		if !ok {
			t.Fatalf("Resource does not implement ListResourceWithConfigure")
		}
		resp := &res.ConfigureResponse{}
		req := res.ConfigureRequest{
			ProviderData: struct{}{}, // Wrong type
		}

		r.Configure(context.Background(), req, resp)

		if !resp.Diagnostics.HasError() {
			t.Error("Expected error for invalid provider data type")
		}
	})

}

func listSpaceQueryConfig(label, providerName string, orgGuID string) string {
	return fmt.Sprintf(`
list "cloudfoundry_space" "%s" {
  provider = "%s"
  config {
   org = "%s"
  }
}`, label, providerName, orgGuID)
}

func listSpaceQueryConfigWithIncludeResource(label, providerName string, orgGuID string) string {
	return fmt.Sprintf(`
list "cloudfoundry_space" "%s" {
  provider = "%s"
  config {
   org = "%s"
  }
  include_resource = true
}`, label, providerName, orgGuID)
}
