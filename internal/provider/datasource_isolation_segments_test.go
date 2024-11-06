package provider

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type IsolationSegmentsModelPtr struct {
	HclType           string
	HclObjectName     string
	Name              *string
	IsolationSegments *string
}

func hclIsolationSegments(ismp *IsolationSegmentsModelPtr) string {
	if ismp != nil {
		s := `
			{{.HclType}} "cloudfoundry_isolation_segments" "{{.HclObjectName}}" {
			{{- if .Name}}
				name  = "{{.Name}}"
			{{- end -}}
			{{if .IsolationSegments}}
				isolation_segments = {{.IsolationSegments}}
			{{- end }}
			}`
		tmpl, err := template.New("isolation_segments").Parse(s)
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
	return ismp.HclType + ` cloudfoundry_isolation_segments" ` + ismp.HclObjectName + `  {}`
}

func TestIsolationSegmentsDataSource_Configure(t *testing.T) {
	var (
		// in staging
		isolationSegmentName = "trial"
		resourceName         = "data.cloudfoundry_isolation_segments.ds"
	)
	t.Parallel()
	t.Run("get available datasource isolation segments", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_isolation_segments")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclIsolationSegments(&IsolationSegmentsModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Name:          &isolationSegmentName,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "isolation_segments.#", "1"),
					),
				},
			},
		})
	})
	t.Run("error path - get unavailable isolation segments", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_isolation_segments_invalid")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclIsolationSegments(&IsolationSegmentsModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Name:          strtostrptr("testunavailable"),
					}),
					ExpectError: regexp.MustCompile(`Unable to find any Isolation Segment in given list`),
				},
			},
		})
	})
}
