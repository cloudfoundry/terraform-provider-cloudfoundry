package provider

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type IsolationSegmentModelPtr struct {
	HclType       string
	HclObjectName string
	Name          *string
	Id            *string
	Labels        *string
	Annotations   *string
	CreatedAt     *string
	UpdatedAt     *string
}

func hclIsolationSegment(ismp *IsolationSegmentModelPtr) string {
	if ismp != nil {
		s := `
		{{.HclType}} "cloudfoundry_isolation_segment" {{.HclObjectName}} {
			{{- if .Name}}
				name = "{{.Name}}"
			{{- end -}}
			{{if .Id}}
				id = "{{.Id}}"
			{{- end -}}
			{{if .Labels}}
				labels = {{.Labels}}
			{{- end -}}
			{{if .Annotations}}
				annotations = {{.Annotations}}
			{{- end -}}
			{{if .CreatedAt}}
				created_at = "{{.CreatedAt}}"
			{{- end -}}
			{{if .UpdatedAt}}
				updated_at = "{{.UpdatedAt}}"
			{{- end }}
			}`
		tmpl, err := template.New("resource_isolation_segment").Parse(s)
		if err != nil {
			panic(err)
		}
		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, ismp)
		if err != nil {
			panic(err)
		}
		return buf.String()
	}
	return ismp.HclType + ` "cloudfoundry_isolation_segment "` + ismp.HclObjectName + ` {}`
}

func TestIsolationSegmentResource_Configure(t *testing.T) {
	var (
		// in staging
		createName   = "segment"
		createName2  = "segment-new"
		updateName   = "segment-2"
		resourceName = "cloudfoundry_isolation_segment.rs"
		existingName = "hifi"
	)
	t.Parallel()
	t.Run("happy path - create/read/update/delete isolation segment", func(t *testing.T) {

		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_isolation_segment_crud")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclIsolationSegment(&IsolationSegmentModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rs",
						Name:          &createName,
						Labels:        &testCreateLabel,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestCheckResourceAttr(resourceName, "name", createName),
						resource.TestCheckResourceAttr(resourceName, "labels.purpose", "testing"),
					),
				},
				{
					Config: hclProvider(nil) + hclIsolationSegment(&IsolationSegmentModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rs",
						Name:          &updateName,
						Labels:        &testUpdateLabel,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestCheckResourceAttr(resourceName, "name", updateName),
						resource.TestCheckResourceAttr(resourceName, "labels.purpose", "production"),
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

	t.Run("error path - create isolation segment with existing name", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_isolation_segment_invalid_create")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclIsolationSegment(&IsolationSegmentModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rs",
						Name:          &existingName,
					}),
					ExpectError: regexp.MustCompile(`API Error Creating Isolation Segment`),
				},
			},
		})
	})
	t.Run("error path - update isolation segment with existing name", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_isolation_segment_invalid_update")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclIsolationSegment(&IsolationSegmentModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rs",
						Name:          &createName2,
						Labels:        &testCreateLabel,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestCheckResourceAttr(resourceName, "name", createName2),
						resource.TestCheckResourceAttr(resourceName, "labels.purpose", "testing"),
					),
				},
				{
					Config: hclProvider(nil) + hclIsolationSegment(&IsolationSegmentModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rs",
						Name:          &existingName,
						Labels:        &testUpdateLabel,
					}),
					ExpectError: regexp.MustCompile(`API Error Updating Isolation Segment`),
				},
			},
		})
	})
}
