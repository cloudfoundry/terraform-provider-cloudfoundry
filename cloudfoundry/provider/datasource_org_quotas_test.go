package provider

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type OrgQuotasModelPtr struct {
	HclType       string
	HclObjectName string
	Org           *string
	OrgQuotas     *string
}

func hclOrgQuotas(oqdsmp *OrgQuotasModelPtr) string {
	if oqdsmp != nil {
		s := `
			{{.HclType}} "cloudfoundry_org_quotas" "{{.HclObjectName}}" {
			{{if .Org}}
				org = "{{.Org}}"
			{{- end }}
			{{if .OrgQuotas}}
				org_quotas = {{.OrgQuotas}}
			{{- end }}
			}`
		tmpl, err := template.New("org_quotas").Parse(s)
		if err != nil {
			panic(err)
		}
		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, oqdsmp)
		if err != nil {
			panic(err)
		}
		return buf.String()
	}
	return oqdsmp.HclType + ` cloudfoundry_org_quotas" ` + oqdsmp.HclObjectName + `  {}`
}
func TestOrgQuotasDataSource_Configure(t *testing.T) {
	resourceName := "data.cloudfoundry_org_quotas.ds"
	orgId := "261e5031-3e54-4b12-b316-94b3195b5f8e"
	t.Parallel()
	t.Run("error path - get unavailable datasource org quota", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_org_quotas_invalid")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclOrgQuotas(&OrgQuotasModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Org:           &invalidOrgGUID,
					}),
					ExpectError: regexp.MustCompile(`API Error Fetching Organization`),
				},
			},
		})
	})
	t.Run("get available datasource org quota", func(t *testing.T) {
		cfg := getCFHomeConf()

		rec := cfg.SetupVCR(t, "fixtures/datasource_org_quotas")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclOrgQuotas(&OrgQuotasModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Org:           new(orgId),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "org_quotas.#", "1"),
					),
				},
			},
		})
	})
}
