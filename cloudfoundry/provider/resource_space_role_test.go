package provider

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type SpaceRoleResourceModelPtr struct {
	HclType       string
	HclObjectName string
	UserName      *string
	Origin        *string
	Type          *string
	User          *string
	Space         *string
	Id            *string
	CreatedAt     *string
	UpdatedAt     *string
}

func hclSpaceRoleResource(rrmp *SpaceRoleResourceModelPtr) string {
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
			{{if .UserName}}
				username = "{{.UserName}}"
			{{- end -}}
			{{if .Origin}}
				origin = "{{.Origin}}"
			{{- end -}}
			{{if .CreatedAt}}
				created_at = "{{.CreatedAt}}"
			{{- end -}}
			{{if .UpdatedAt}}
				updated_at = "{{.UpdatedAt}}"
			{{- end }}
			}`
		tmpl, err := template.New("resource_space_role").Parse(s)
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
	return rrmp.HclType + ` "cloudfoundry_space_role "` + rrmp.HclObjectName + ` {}`
}

func TestSpaceRoleResource_Configure(t *testing.T) {
	var (
		// in canary -> PerformanceTeamBLR -> tf-space-1
		testSpaceGUID = "02c0cc92-6ecc-44b1-b7b2-096ca19ee143"
		testUserName  = "kesavan.s@sap.com"
		origin        = "sap.ids"
	)
	t.Parallel()
	t.Run("happy path - create space role", func(t *testing.T) {
		resourceName := "cloudfoundry_space_role.rf"
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_space_role")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclSpaceRoleResource(&SpaceRoleResourceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rf",
						Type:          strtostrptr("space_auditor"),
						UserName:      strtostrptr(testUserName),
						Origin:        strtostrptr(origin),
						Space:         strtostrptr(testSpaceGUID),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestMatchResourceAttr(resourceName, "user", regexpValidUUID),
						resource.TestCheckResourceAttr(resourceName, "space", testSpaceGUID),
					),
				},
			},
		})
	})

	t.Run("error path - create role with existing id", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_space_role_invalid")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclSpaceRoleResource(&SpaceRoleResourceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rsi",
						Type:          strtostrptr("space_manager"),
						User:          strtostrptr(testUser2GUID),
						Space:         strtostrptr(testSpaceGUID),
					}),
					ExpectError: regexp.MustCompile(`API Error Registering Role`),
				},
			},
		})
	})

}
