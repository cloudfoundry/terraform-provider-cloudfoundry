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

const testRouteSpaceGUID = "10b8f9bd-0731-49f3-9e88-a17b6eec090e"

func TestRouteListResource(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/list_resource_route")
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
					Config: hclProvider(nil) + listRouteQueryConfig(
						"route_list",
						"cloudfoundry",
					),
					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("cloudfoundry_route.route_list", 2),
						querycheck.ExpectIdentity(
							"cloudfoundry_route.route_list",
							map[string]knownvalue.Check{
								"route_guid": knownvalue.StringRegexp(regexpValidUUID),
							},
						),
					},
				},
				{
					// List Query with include_resource = true
					Query: true,
					Config: hclProvider(nil) + listRouteQueryConfigWithIncludeResource(
						"route_list",
						"cloudfoundry",
						testRouteSpaceGUID,
					),
					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("cloudfoundry_route.route_list", 2),
						// Verify full resource data is populated
						querycheck.ExpectResourceKnownValues(
							"cloudfoundry_route.route_list",
							queryfilter.ByResourceIdentity(map[string]knownvalue.Check{
								"route_guid": knownvalue.StringExact("6a18e548-e862-4cfc-b194-cc1d5a53346d"),
							}),
							[]querycheck.KnownValueCheck{
								{
									Path:       tfjsonpath.New("id"),
									KnownValue: knownvalue.StringRegexp(regexpValidUUID),
								},
								{
									Path:       tfjsonpath.New("space"),
									KnownValue: knownvalue.StringExact(testRouteSpaceGUID),
								},
								{
									Path:       tfjsonpath.New("host"),
									KnownValue: knownvalue.StringExact("delayed-broker"),
								},
							},
						),
					},
				},
			},
		})
	})

	t.Run("error path - configure", func(t *testing.T) {
		r, ok := NewRouteListResource().(list.ListResourceWithConfigure)
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

func listRouteQueryConfig(label, providerName string) string {
	return fmt.Sprintf(`
list "cloudfoundry_route" "%s" {
  provider = %s
}`, label, providerName)
}

func listRouteQueryConfigWithIncludeResource(label, providerName, spaceGUID string) string {
	return fmt.Sprintf(`
list "cloudfoundry_route" "%s" {
  provider         = %s
  include_resource = true
  config {
    space = "%s"
  }
}`, label, providerName, spaceGUID)
}
