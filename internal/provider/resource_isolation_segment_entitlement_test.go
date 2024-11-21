package provider

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type IsolationSegmentEntitlementModelPtr struct {
	HclType       string
	HclObjectName string
	Segment       *string
	Orgs          *string
	Default       *bool
}

func hclIsolationSegmentEntitlement(isemp *IsolationSegmentEntitlementModelPtr) string {
	if isemp != nil {
		s := `
		{{.HclType}} "cloudfoundry_isolation_segment_entitlement" {{.HclObjectName}} {
			{{if .Segment}}
				segment = "{{.Segment}}"
			{{- end -}}
			{{if .Orgs}}
				orgs = {{.Orgs}}
			{{- end -}}
			{{if .Default}}
				default = {{.Default}}
			{{- end }}
			}`
		tmpl, err := template.New("resource_isolation_segment_entitlement").Parse(s)
		if err != nil {
			panic(err)
		}
		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, isemp)
		if err != nil {
			panic(err)
		}
		return buf.String()
	}
	return isemp.HclType + ` "cloudfoundry_isolation_segment_entitlement" ` + isemp.HclObjectName + ` {}`
}

func TestIsolationSegmentEntitlementResource_Configure(t *testing.T) {
	var (
		// in staging
		resourceName      = "cloudfoundry_isolation_segment_entitlement.rs"
		segmentGUID       = "63ae51b9-9073-4409-81b0-3704b8de85dd"
		entitleOrgsCreate = `["db29f4b8-d39e-4f5c-b24d-cb34cde27abf", "3c6ee045-7791-4d65-a12b-f42ab1b283eb"]`
		entitleOrgsUpdate = `["db29f4b8-d39e-4f5c-b24d-cb34cde27abf", "b2ab6999-260f-4db8-8e9e-098d7f55ecaf"]`
		invalidOrgs       = `["da7f4f0d-0488-4880-9fbd-99ebfda12bed"]`
	)
	t.Parallel()
	t.Run("happy path - create/read/update/delete isolation segment entitlement", func(t *testing.T) {

		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_isolation_segment_entitlement_crud")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclIsolationSegmentEntitlement(&IsolationSegmentEntitlementModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rs",
						Segment:       &segmentGUID,
						Orgs:          &entitleOrgsCreate,
						Default:       booltoboolptr(true),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "segment", segmentGUID),
						resource.TestCheckResourceAttr(resourceName, "orgs.#", "2"),
						resource.TestCheckResourceAttr(resourceName, "default", "true"),
					),
				},
				{
					Config: hclProvider(nil) + hclIsolationSegmentEntitlement(&IsolationSegmentEntitlementModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rs",
						Segment:       &segmentGUID,
						Orgs:          &entitleOrgsCreate,
						Default:       booltoboolptr(false),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "segment", segmentGUID),
						resource.TestCheckResourceAttr(resourceName, "orgs.#", "2"),
						resource.TestCheckResourceAttr(resourceName, "default", "false"),
					),
				},
				{
					Config: hclProvider(nil) + hclIsolationSegmentEntitlement(&IsolationSegmentEntitlementModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rs",
						Segment:       &segmentGUID,
						Orgs:          &entitleOrgsUpdate,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "segment", segmentGUID),
						resource.TestCheckResourceAttr(resourceName, "orgs.#", "2"),
						resource.TestCheckNoResourceAttr(resourceName, "default"),
					),
				},
			},
		})
	})

	t.Run("error path - create isolation segment entitlement with invalid org", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_isolation_segment_entitlement_invalid")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclIsolationSegmentEntitlement(&IsolationSegmentEntitlementModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rs",
						Segment:       &segmentGUID,
						Orgs:          &invalidOrgs,
					}),
					ExpectError: regexp.MustCompile(`API Error Entitling Isolation Segment to Organizations`),
				},
			},
		})
	})
}
