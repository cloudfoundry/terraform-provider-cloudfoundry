package provider

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
)

type OrgRoleResourceModelPtr struct {
	HclType       string
	HclObjectName string
	UserName      *string
	Origin        *string
	Type          *string
	User          *string
	Id            *string
	Organization  *string
	CreatedAt     *string
	UpdatedAt     *string
}

func hclOrgRoleResource(rrmp *OrgRoleResourceModelPtr) string {
	if rrmp != nil {
		s := `
		{{.HclType}} "cloudfoundry_org_role" {{.HclObjectName}} {
			{{- if .Type}}
				type = "{{.Type}}"
			{{- end -}}
			{{if .Id}}
				id = "{{.Id}}"
			{{- end -}}
			{{if .User}}
				user = "{{.User}}"
			{{- end -}}
			{{if .Organization}}
				org = "{{.Organization}}"
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
		tmpl, err := template.New("resource_org_role").Parse(s)
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
	return rrmp.HclType + ` "cloudfoundry_org_role "` + rrmp.HclObjectName + ` {}`
}

func TestOrgRoleResource_Configure(t *testing.T) {
	var (
		// in canary -> PerformanceTeamBLR -> tf-space-1
		testUserName  = "debaditya.ray@sap.com"
		origin        = "sap.ids"
		testUserName2 = "6c05946f-f32b-46f3-8ce5-a9c2020c2aa5"
		testOrgGUID   = "b4da43cd-2055-4d4d-ae6e-4066ce53a8b9"
	)
	t.Parallel()
	t.Run("happy path - create org role", func(t *testing.T) {
		resourceName := "cloudfoundry_org_role.rs"
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_org_role")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclOrgRoleResource(&OrgRoleResourceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rs",
						Type:          strtostrptr("organization_user"),
						User:          strtostrptr(testUser2GUID),
						Organization:  strtostrptr(testOrg2GUID),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestCheckResourceAttr(resourceName, "user", testUser2GUID),
						resource.TestCheckResourceAttr(resourceName, "org", testOrg2GUID),
					),
				},
				{
					ResourceName:      resourceName,
					ImportStateIdFunc: getIdForImport(resourceName),
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})

	t.Run("happy path - create org role with identity", func(t *testing.T) {
		resourceName := "cloudfoundry_org_role.rs"
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_org_role_import_identity")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclOrgRoleResource(&OrgRoleResourceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rs",
						Type:          strtostrptr("organization_user"),
						User:          strtostrptr(testUserName2),
						Organization:  strtostrptr(testOrgGUID),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestCheckResourceAttr(resourceName, "user", testUserName2),
						resource.TestCheckResourceAttr(resourceName, "org", testOrgGUID),
					),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectIdentity("cloudfoundry_org_role.rs", map[string]knownvalue.Check{
							"role_guid": knownvalue.NotNull(),
						}),
					},
				},
				{
					ResourceName:    resourceName,
					ImportState:     true,
					ImportStateKind: resource.ImportBlockWithResourceIdentity,
				},
			},
		})
	})

	t.Run("error path - create role with existing id", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_org_role_invalid")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclOrgRoleResource(&OrgRoleResourceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rsi",
						Type:          strtostrptr("organization_manager"),
						UserName:      strtostrptr(testUserName),
						Origin:        strtostrptr(origin),
						Organization:  strtostrptr(testOrg2GUID),
					}),
					ExpectError: regexp.MustCompile(`API Error Registering Role`),
				},
			},
		})
	})

}
