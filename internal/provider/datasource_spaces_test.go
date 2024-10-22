package provider

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type SpacesModelPtr struct {
	HclType       string
	HclObjectName string
	Name          *string
	OrgId         *string
	Spaces        *string
}

func hclSpaces(smp *SpacesModelPtr) string {
	if smp != nil {
		s := `
		{{.HclType}} "cloudfoundry_spaces" {{.HclObjectName}} {
			{{- if .Name}}
				name = "{{.Name}}"
			{{- end -}}
			{{if .OrgId}}
				org = "{{.OrgId}}"
			{{- end -}}
			{{if .Spaces}}
				spaces = {{.Spaces}}
			{{- end }}
			}`
		tmpl, err := template.New("datasource_spaces").Parse(s)
		if err != nil {
			panic(err)
		}
		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, smp)
		if err != nil {
			panic(err)
		}
		return buf.String()
	}
	return smp.HclType + ` "cloudfoundry_spaces" ` + smp.HclObjectName + ` {}`
}

func TestSpacesDataSource_Configure(t *testing.T) {
	t.Parallel()
	dataSourceName := "data.cloudfoundry_spaces.ds"
	t.Run("happy path - read spaces", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_spaces")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclSpaces(&SpacesModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						OrgId:         strtostrptr(testOrg2GUID),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(dataSourceName, "org", testOrg2GUID),
						resource.TestCheckResourceAttr(dataSourceName, "spaces.#", "16"),
					),
				},
				{
					Config: hclProvider(nil) + hclSpaces(&SpacesModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						OrgId:         strtostrptr(testOrg2GUID),
						Name:          strtostrptr("wow"),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(dataSourceName, "org", testOrg2GUID),
						resource.TestCheckResourceAttr(dataSourceName, "spaces.#", "1"),
					),
				},
			},
		})
	})
	t.Run("error path - org does not exist", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_spaces_invalid_org")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclSpaces(&SpacesModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						OrgId:         strtostrptr(invalidOrgGUID),
					}),
					ExpectError: regexp.MustCompile(`API Error Fetching Organization`),
				},
			},
		})
	})
	t.Run("error path - spaces do not exist", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_spaces_invalid_space")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclSpaces(&SpacesModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Name:          strtostrptr(testSpace),
						OrgId:         strtostrptr(testOrg2GUID),
					}),
					ExpectError: regexp.MustCompile(`Unable to find any space in list`),
				},
			},
		})
	})

}
