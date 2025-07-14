package provider

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type OrgRolesModelPtr struct {
	HclType       string
	HclObjectName string
	Type          *string
	User          *string
	Org           *string
	Roles         *string
}

func hclOrgRolesDataSource(rrmp *OrgRolesModelPtr) string {
	if rrmp != nil {
		s := `
		{{.HclType}} "cloudfoundry_org_roles" {{.HclObjectName}} {
			{{- if .Type}}
				type = "{{.Type}}"
			{{- end -}}
			{{if .User}}
				user = "{{.User}}"
			{{- end -}}
			{{if .Org}}
				org = "{{.Org}}"
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
	return rrmp.HclType + ` "cloudfoundry_org_roles" ` + rrmp.HclObjectName + ` {}`
}

func TestOrgRolesDataSource_Configure(t *testing.T) {
	testOrgGuid := "261e5031-3e54-4b12-b316-94b3195b5f8e"
	t.Parallel()
	dataSourceName := "data.cloudfoundry_org_roles.ds"
	t.Run("happy path - read org role", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_org_roles")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclOrgRolesDataSource(&OrgRolesModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Org:           strtostrptr(testOrgGuid),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(dataSourceName, "org", testOrgGuid),
						resource.TestCheckResourceAttr(dataSourceName, "roles.#", "5")),
				},
				{
					Config: hclProvider(nil) + hclOrgRolesDataSource(&OrgRolesModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Org:           strtostrptr(testOrgGuid),
						Type:          strtostrptr("organization_manager"),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(dataSourceName, "org", testOrgGuid),
						resource.TestCheckResourceAttr(dataSourceName, "roles.#", "2")),
				},
			},
		})
	})
	t.Run("error path - role does not exist", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_org_roles_invalid")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclOrgRolesDataSource(&OrgRolesModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Type:          strtostrptr("organization_man"),
						Org:           strtostrptr(invalidOrgGUID),
					}),
					ExpectError: regexp.MustCompile(`Invalid Attribute Value Match`),
				},
				{
					Config: hclProvider(nil) + hclOrgRolesDataSource(&OrgRolesModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Org:           strtostrptr(invalidOrgGUID),
					}),
					ExpectError: regexp.MustCompile(`API Error Fetching Org`),
				},
			},
		})
	})
}
