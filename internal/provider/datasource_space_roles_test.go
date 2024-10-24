package provider

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type SpaceRolesModelPtr struct {
	HclType       string
	HclObjectName string
	Type          *string
	User          *string
	Space         *string
	Roles         *string
}

func hclSpaceRolesDataSource(rrmp *SpaceRolesModelPtr) string {
	if rrmp != nil {
		s := `
		{{.HclType}} "cloudfoundry_space_roles" {{.HclObjectName}} {
			{{- if .Type}}
				type = "{{.Type}}"
			{{- end -}}
			{{if .User}}
				user = "{{.User}}"
			{{- end -}}
			{{if .Space}}
				space = "{{.Space}}"
			{{- end -}}
			{{if .Roles}}
				roles = "{{.Roles}}"
			{{- end }}
			}`
		tmpl, err := template.New("datasource_role").Parse(s)
		if err != nil {
			panic(err)
		}
		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, rrmp)
		if err != nil {
			panic(err)
		}
		return buf.String()
	}
	return rrmp.HclType + ` "cloudfoundry_space_roles" ` + rrmp.HclObjectName + ` {}`
}

func TestSpaceRolesDataSource_Configure(t *testing.T) {
	testSpaceGuid := "02c0cc92-6ecc-44b1-b7b2-096ca19ee143"
	t.Parallel()
	dataSourceName := "data.cloudfoundry_space_roles.ds"
	t.Run("happy path - read space role", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_space_roles")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclSpaceRolesDataSource(&SpaceRolesModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Space:         strtostrptr(testSpaceGuid),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(dataSourceName, "space", testSpaceGuid),
						resource.TestCheckResourceAttr(dataSourceName, "roles.#", "8")),
				},
				{
					Config: hclProvider(nil) + hclSpaceRolesDataSource(&SpaceRolesModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Space:         strtostrptr(testSpaceGuid),
						Type:          strtostrptr("space_developer"),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(dataSourceName, "space", testSpaceGuid),
						resource.TestCheckResourceAttr(dataSourceName, "roles.#", "5")),
				},
			},
		})
	})
	t.Run("error path - role does not exist", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_space_roles_invalid")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclSpaceRolesDataSource(&SpaceRolesModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Type:          strtostrptr("space_developers"),
						Space:         strtostrptr(invalidOrgGUID),
					}),
					ExpectError: regexp.MustCompile(`Invalid Attribute Value Match`),
				},
				{
					Config: hclProvider(nil) + hclSpaceRolesDataSource(&SpaceRolesModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Space:         strtostrptr(invalidOrgGUID),
					}),
					ExpectError: regexp.MustCompile(`API Error Fetching Space`),
				},
			},
		})
	})
}
