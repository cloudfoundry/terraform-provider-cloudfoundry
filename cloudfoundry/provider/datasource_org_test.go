package provider

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type OrgModelPtr struct {
	HclType       string
	HclObjectName string
	Name          *string
	Id            *string
	Labels        *string
	Annotations   *string
	Suspended     *bool
}

func hclOrg(odsmp *OrgModelPtr) string {
	if odsmp != nil {
		s := `
		{{.HclType}} "cloudfoundry_org" {{.HclObjectName}} {
			{{- if .Name}}
				name  = "{{.Name}}"
			{{- end -}}
			{{if .Id}}
				id = "{{.Id}}"
			{{- end -}}
			{{if .Suspended}}
				suspended = "{{.Suspended}}"
			{{- end -}}
			{{if .Labels}}
				labels = {{.Labels}}
			{{- end -}}
			{{if .Annotations}}
				annotations = {{.Annotations}}
			{{- end }}
			}`
		tmpl, err := template.New("org").Parse(s)
		if err != nil {
			panic(err)
		}
		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, odsmp)
		if err != nil {
			panic(err)
		}
		return buf.String()
	}
	return odsmp.HclType + ` "cloudfoundry_org  "` + odsmp.HclObjectName + ` {}`
}
func TestOrgDataSource_Configure(t *testing.T) {
	t.Parallel()
	t.Run("error path - get unavailable datasource org", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_org_invalid_orgname")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclOrg(&OrgModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Name:          new("testunavailableorg"),
					}),
					ExpectError: regexp.MustCompile(`Error: Unable to find org data in list`),
				},
			},
		})
	})
	t.Run("get available datasource org", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_org")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclOrg(&OrgModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Name:          new(testOrg),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr("data.cloudfoundry_org.ds", "id", regexpValidUUID),
						resource.TestCheckResourceAttr("data.cloudfoundry_org.ds", labelsKey+".env", "canary"),
						resource.TestCheckResourceAttr("data.cloudfoundry_org.ds", annotationsKey+".%", "0"),
					),
				},
			},
		})
	})
}
