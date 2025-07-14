package provider

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type UsersDataSourceModelPtr struct {
	HclType       string
	HclObjectName string
	Space         *string
	Organization  *string
	Users         *string
}

func hclDataSourceUsers(usdsmp *UsersDataSourceModelPtr) string {
	if usdsmp != nil {
		s := `
		{{.HclType}} "cloudfoundry_users" {{.HclObjectName}} {
			{{- if .Space}}
				space = "{{.Space}}"
			{{- end -}}
			{{- if .Organization}}
				org = "{{.Organization}}"
			{{- end -}}	
			{{if .Users}}
				users = {{.Users}}
			{{- end }}
			}`
		tmpl, err := template.New("datasource_users").Parse(s)
		if err != nil {
			panic(err)
		}
		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, usdsmp)
		if err != nil {
			panic(err)
		}
		return buf.String()
	}
	return usdsmp.HclType + ` "cloudfoundry_users "` + usdsmp.HclObjectName + ` {}`
}

func TestUsersDataSource_Configure(t *testing.T) {
	t.Parallel()
	dataSourceName := "data.cloudfoundry_users.ds"
	t.Run("happy path - read users from org/space", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_users")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclDataSourceUsers(&UsersDataSourceModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Organization:  strtostrptr(testOrg2GUID),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(dataSourceName, "users.0.id", regexpValidUUID),
						resource.TestMatchResourceAttr(dataSourceName, "users.0.created_at", regexpValidRFC3999Format),
						resource.TestCheckResourceAttr(dataSourceName, "users.0.origin", "sap.ids"),
					),
				},
				{
					Config: hclProvider(nil) + hclDataSourceUsers(&UsersDataSourceModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Space:         strtostrptr(testSpace2GUID),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(dataSourceName, "users.0.id", regexpValidUUID),
						resource.TestMatchResourceAttr(dataSourceName, "users.0.created_at", regexpValidRFC3999Format),
						resource.TestCheckResourceAttr(dataSourceName, "users.0.origin", "sap.ids"),
					),
				},
			},
		})
	})
	t.Run("error path - space/org does not exist", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_users_invalid")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclDataSourceUsers(&UsersDataSourceModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Space:         strtostrptr(testOrg2GUID),
					}),
					ExpectError: regexp.MustCompile(`API Error Fetching Users`),
				},
				{
					Config: hclProvider(nil) + hclDataSourceUsers(&UsersDataSourceModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Organization:  strtostrptr(testSpace2GUID),
					}),
					ExpectError: regexp.MustCompile(`API Error Fetching Users`),
				},
			},
		})
	})
}
