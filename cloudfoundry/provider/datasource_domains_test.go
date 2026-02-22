package provider

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type DomainsModelPtr struct {
	HclType       string
	HclObjectName string
	Name          *string
	Org           *string
	Domains       *string
}

func hclDomains(ddsmp *DomainsModelPtr) string {
	if ddsmp != nil {
		s := `
			{{.HclType}} "cloudfoundry_domains" "{{.HclObjectName}}" {
			{{- if .Name}}
				name  = "{{.Name}}"
			{{- end -}}
			{{if .Org}}
				org = "{{.Org}}"
			{{- end }}
			{{if .Domains}}
				domains = {{.Domains}}
			{{- end }}
			}`
		tmpl, err := template.New("domains").Parse(s)
		if err != nil {
			panic(err)
		}
		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, ddsmp)
		if err != nil {
			panic(err)
		}
		return buf.String()
	}
	return ddsmp.HclType + ` cloudfoundry_domains" ` + ddsmp.HclObjectName + `  {}`
}
func TestDomainsDataSource_Configure(t *testing.T) {
	resourceName := "data.cloudfoundry_domains.ds"
	orgId := "261e5031-3e54-4b12-b316-94b3195b5f8e"
	t.Parallel()
	t.Run("error path - get domains for invalid org", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_domains_invalid")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclDomains(&DomainsModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Org:           &invalidOrgGUID,
					}),
					ExpectError: regexp.MustCompile(`API Error Fetching Organization`),
				},
			},
		})
	})
	t.Run("get domains within the organization", func(t *testing.T) {
		cfg := getCFHomeConf()

		rec := cfg.SetupVCR(t, "fixtures/datasource_domains")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclDomains(&DomainsModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Org:           new(orgId),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "domains.#", "2"),
					),
				},
			},
		})
	})
}
