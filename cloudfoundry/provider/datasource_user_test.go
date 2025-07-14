package provider

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type UserDataSourceModelPtr struct {
	HclType       string
	HclObjectName string
	Name          *string
	Users         *string
}

func hclDataSourceUser(udsmp *UserDataSourceModelPtr) string {
	if udsmp != nil {
		s := `
		{{.HclType}} "cloudfoundry_user" {{.HclObjectName}} {
			{{- if .Name}}
				name = "{{.Name}}"
			{{- end -}}
			{{if .Users}}
				users = {{.Users}}
			{{- end }}
			}`
		tmpl, err := template.New("datasource_user").Parse(s)
		if err != nil {
			panic(err)
		}
		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, udsmp)
		if err != nil {
			panic(err)
		}
		return buf.String()
	}
	return udsmp.HclType + ` "cloudfoundry_user "` + udsmp.HclObjectName + ` {}`
}

func TestUserDataSource_Configure(t *testing.T) {
	t.Parallel()
	dataSourceName := "data.cloudfoundry_user.ds"
	t.Run("happy path - read user", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_user")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclDataSourceUser(&UserDataSourceModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Name:          strtostrptr(testUser),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(dataSourceName, "users.#", "1"),
						resource.TestMatchResourceAttr(dataSourceName, "users.0.id", regexpValidUUID),
						resource.TestCheckResourceAttr(dataSourceName, "users.0.presentation_name", testUser),
						resource.TestCheckResourceAttr(dataSourceName, "users.0.labels.enviroment", "staging"),
					),
				},
			},
		})
	})
	t.Run("error path - user does not exist", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_user_invalid")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclDataSourceUser(&UserDataSourceModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Name:          strtostrptr(testUser + "x"),
					}),
					ExpectError: regexp.MustCompile(`Unable to find user data in list`),
				},
			},
		})
	})
}
