package provider

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type UserCFResourceModelPtr struct {
	HclType          string
	HclObjectName    string
	UserName         *string
	Origin           *string
	PresentationName *string
	Id               *string
	Labels           *string
	Annotations      *string
	CreatedAt        *string
	UpdatedAt        *string
}

func hclResourceUserCF(urmp *UserCFResourceModelPtr) string {
	if urmp != nil {
		s := `
		{{.HclType}} "cloudfoundry_user_cf" {{.HclObjectName}} {
			{{- if .UserName}}
				username = "{{.UserName}}"
			{{- end -}}
			{{if .Id}}
				id = "{{.Id}}"
			{{- end -}}
			{{if .Origin}}
				origin = "{{.Origin}}"
			{{- end -}}
			{{if .PresentationName}}
				presentation_name = "{{.PresentationName}}"
			{{- end -}}
			{{if .CreatedAt}}
				created_at = "{{.CreatedAt}}"
			{{- end -}}
			{{if .UpdatedAt}}
				updated_at = "{{.UpdatedAt}}"
			{{- end -}}
			{{if .Labels}}
				labels = {{.Labels}}
			{{- end -}}
			{{if .Annotations}}
				annotations = {{.Annotations}}
			{{- end }}
			}`
		tmpl, err := template.New("resource_user").Parse(s)
		if err != nil {
			panic(err)
		}
		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, urmp)
		if err != nil {
			panic(err)
		}
		return buf.String()
	}
	return urmp.HclType + ` "cloudfoundry_user_cf" ` + urmp.HclObjectName + ` {}`
}

func TestCFUserResource_Configure(t *testing.T) {
	t.Parallel()
	var (
		createUsername   = "tf-test@example.com"
		origin           = "sap.ids"
		updateUsername   = "tf-test-updated@example.com"
		guidUser         = "1234567898765432"
		resourceName     = "cloudfoundry_user_cf.us"
		existingUsername = "test"
		testInvalidLabel = `{"purpose@!": "testing", landscape: "test"}`
	)
	t.Run("happy path - create/update/import/delete user", func(t *testing.T) {

		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_user_cf_crud")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclResourceUserCF(&UserCFResourceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "us",
						UserName:      &createUsername,
						Origin:        &origin,
						Labels:        &testCreateLabel,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckNoResourceAttr(resourceName, "annotations"),
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestCheckResourceAttr(resourceName, "username", createUsername),
						resource.TestCheckResourceAttr(resourceName, "presentation_name", createUsername),
						resource.TestCheckResourceAttr(resourceName, "origin", origin),
						resource.TestCheckResourceAttr(resourceName, "labels.purpose", "testing"),
					),
				},
				{
					Config: hclProvider(nil) + hclResourceUserCF(&UserCFResourceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "us",
						UserName:      &updateUsername,
						Origin:        &origin,
						Labels:        &testCreateLabel,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckNoResourceAttr(resourceName, "annotations"),
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestCheckResourceAttr(resourceName, "username", updateUsername),
						resource.TestCheckResourceAttr(resourceName, "presentation_name", updateUsername),
						resource.TestCheckResourceAttr(resourceName, "origin", origin),
						resource.TestCheckResourceAttr(resourceName, "labels.purpose", "testing"),
					),
				},
				{
					Config: hclProvider(nil) + hclResourceUserCF(&UserCFResourceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "us",
						Id:            &guidUser,
						Labels:        &testCreateLabel,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckNoResourceAttr(resourceName, "annotations"),
						resource.TestCheckResourceAttr(resourceName, "id", guidUser),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestCheckResourceAttr(resourceName, "presentation_name", guidUser),
						resource.TestCheckResourceAttr(resourceName, "labels.purpose", "testing"),
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
	t.Run("error path - invalid create/update scenarios", func(t *testing.T) {

		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_user_cf_crud_invalid")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclResourceUserCF(&UserCFResourceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "us",
					}),
					ExpectError: regexp.MustCompile(`Error: Invalid Attribute Combination`),
				},
				{
					Config: hclProvider(nil) + hclResourceUserCF(&UserCFResourceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "us",
						UserName:      &createUsername,
					}),
					ExpectError: regexp.MustCompile(`Error: API Error Creating CF User`),
				},
				{
					Config: hclProvider(nil) + hclResourceUserCF(&UserCFResourceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "us",
						UserName:      &existingUsername,
						Labels:        &testInvalidLabel,
					}),
					ExpectError: regexp.MustCompile(`API Error Creating CF User`),
				},
			},
		})
	})
}
