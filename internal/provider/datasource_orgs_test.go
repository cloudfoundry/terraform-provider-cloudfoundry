package provider

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type OrgsModelPtr struct {
	HclType       string
	HclObjectName string
	Name          *string
	Orgs          *string
}

func hclOrgs(odsmp *OrgsModelPtr) string {
	if odsmp != nil {
		s := `
		{{.HclType}} "cloudfoundry_orgs" {{.HclObjectName}} {
			{{- if .Name}}
				name  = "{{.Name}}"
			{{- end -}}
			{{if .Orgs}}
				orgs = {{.Orgs}}
			{{- end }}
			}`
		tmpl, err := template.New("orgs").Parse(s)
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
	return odsmp.HclType + ` "cloudfoundry_orgs" ` + odsmp.HclObjectName + ` {}`
}
func TestOrgsDataSource_Configure(t *testing.T) {
	t.Parallel()
	t.Run("error path - get unavailable orgs", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_orgs_invalid")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclOrgs(&OrgsModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Name:          strtostrptr("testunavailableorg"),
					}),
					ExpectError: regexp.MustCompile(`Error: Unable to find any org in list`),
				},
			},
		})
	})
	t.Run("get available datasource org", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_orgs")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclOrgs(&OrgsModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Name:          strtostrptr("PerformanceTeamBLR"),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.cloudfoundry_orgs.ds", "orgs.#", "1"),
					),
				},
			},
		})
	})
}
