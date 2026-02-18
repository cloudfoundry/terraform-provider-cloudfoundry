package provider

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type RoleModelPtr struct {
	HclType       string
	HclObjectName string
	Type          *string
	User          *string
	Space         *string
	Id            *string
	Organization  *string
	CreatedAt     *string
	UpdatedAt     *string
}

func hclSpaceRoleDataSource(rrmp *RoleModelPtr) string {
	if rrmp != nil {
		s := `
		{{.HclType}} "cloudfoundry_space_role" {{.HclObjectName}} {
			{{- if .Type}}
				type = "{{.Type}}"
			{{- end -}}
			{{if .Id}}
				id = "{{.Id}}"
			{{- end -}}
			{{if .User}}
				user = "{{.User}}"
			{{- end -}}
			{{if .Space}}
				space = "{{.Space}}"
			{{- end -}}
			{{if .CreatedAt}}
				created_at = "{{.CreatedAt}}"
			{{- end -}}
			{{if .UpdatedAt}}
				updated_at = "{{.UpdatedAt}}"
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
	return rrmp.HclType + ` "cloudfoundry_space_role" ` + rrmp.HclObjectName + ` {}`
}

func TestSpaceRoleDataSource_Configure(t *testing.T) {
	testSpaceRoleGUID := "fcbadcb4-6d6c-41c8-a033-98fe24e41ff6"
	testOrgRoleGUID := "4c6849f2-6407-4385-a556-0840369f336b"
	t.Parallel()
	dataSourceName := "data.cloudfoundry_space_role.ds"
	t.Run("happy path - read space role", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_space_role")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclSpaceRoleDataSource(&RoleModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Id:            new(testSpaceRoleGUID),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(dataSourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(dataSourceName, "space", regexpValidUUID),
						resource.TestMatchResourceAttr(dataSourceName, "user", regexpValidUUID),
					),
				},
			},
		})
	})
	t.Run("error path - role does not exist", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_space_role_invalid")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclSpaceRoleDataSource(&RoleModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Id:            new(testOrgRoleGUID),
					}),
					ExpectError: regexp.MustCompile(`Invalid Space Role`),
				},
				{
					Config: hclProvider(nil) + hclSpaceRoleDataSource(&RoleModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Id:            new(invalidOrgGUID),
					}),
					ExpectError: regexp.MustCompile(`API Error Fetching Role`),
				},
			},
		})
	})
}
