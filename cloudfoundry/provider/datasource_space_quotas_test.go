package provider

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type SpaceQuotasModelPtr struct {
	HclType       string
	HclObjectName string
	Name          *string
	Org           *string
	SpaceQuotas   *string
}

func hclSpaceQuotas(sqdsmp *SpaceQuotasModelPtr) string {
	if sqdsmp != nil {
		s := `
			{{.HclType}} "cloudfoundry_space_quotas" "{{.HclObjectName}}" {
			{{- if .Name}}
				name  = "{{.Name}}"
			{{- end -}}
			{{if .Org}}
				org = "{{.Org}}"
			{{- end }}
			{{if .SpaceQuotas}}
				space_quotas = {{.SpaceQuotas}}
			{{- end }}
			}`
		tmpl, err := template.New("space_quotas").Parse(s)
		if err != nil {
			panic(err)
		}
		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, sqdsmp)
		if err != nil {
			panic(err)
		}
		return buf.String()
	}
	return sqdsmp.HclType + ` cloudfoundry_space_quotas" ` + sqdsmp.HclObjectName + `  {}`
}
func TestSpaceQuotasDataSource_Configure(t *testing.T) {
	resourceName := "data.cloudfoundry_space_quotas.ds"
	t.Parallel()
	testOrg4GUID := "3d3fcf37-a526-45f6-be96-4d7c4cee92a5"
	t.Run("error path - get unavailable datasource space quota", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_space_quotas_invalid")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclSpaceQuotas(&SpaceQuotasModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Org:           &invalidOrgGUID,
					}),
					ExpectError: regexp.MustCompile(`API Error Fetching Organization`),
				},
				{
					Config: hclProvider(nil) + hclSpaceQuotas(&SpaceQuotasModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Org:           &testOrg4GUID,
						Name:          strtostrptr("hi"),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "space_quotas.#", "0"),
					),
				},
			},
		})
	})
	t.Run("get available datasource space quota", func(t *testing.T) {
		cfg := getCFHomeConf()

		rec := cfg.SetupVCR(t, "fixtures/datasource_space_quotas")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclSpaceQuotas(&SpaceQuotasModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Org:           &testOrg4GUID,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "space_quotas.#", "2"),
					),
				},
				{
					Config: hclProvider(nil) + hclSpaceQuotas(&SpaceQuotasModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Org:           &testOrg4GUID,
						Name:          strtostrptr("space-quota-1"),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "space_quotas.#", "1"),
					),
				},
			},
		})
	})
}
