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
		orgGuid := "261e5031-3e54-4b12-b316-94b3195b5f8e"
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclSpaces(&SpacesModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						OrgId:         strtostrptr(orgGuid),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(dataSourceName, "org", orgGuid),
						resource.TestCheckResourceAttr(dataSourceName, "spaces.#", "2"),
					),
				},
				{
					Config: hclProvider(nil) + hclSpaces(&SpacesModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						OrgId:         strtostrptr(orgGuid),
						Name:          strtostrptr("test-space"),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(dataSourceName, "org", orgGuid),
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
		testSpace := "wrong-space-name"
		orgGuid := "261e5031-3e54-4b12-b316-94b3195b5f8e"

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclSpaces(&SpacesModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Name:          strtostrptr(testSpace),
						OrgId:         strtostrptr(orgGuid),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(dataSourceName, "spaces.#", "0"),
					),
				},
			},
		})
	})

}
