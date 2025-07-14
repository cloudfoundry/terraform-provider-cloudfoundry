package provider

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type UserResourceModelPtr struct {
	HclType       string
	HclObjectName string
	UserName      *string
	Password      *string
	GivenName     *string
	FamilyName    *string
	Origin        *string
	Groups        *string
	Email         *string
	Id            *string
	Labels        *string
	Annotations   *string
	CreatedAt     *string
	UpdatedAt     *string
}

func hclResourceUser(urmp *UserResourceModelPtr) string {
	if urmp != nil {
		s := `
		{{.HclType}} "cloudfoundry_user" {{.HclObjectName}} {
			{{- if .UserName}}
				username = "{{.UserName}}"
			{{- end -}}
			{{if .Id}}
				id = "{{.Id}}"
			{{- end -}}
			{{if .Password}}
				password = "{{.Password}}"
			{{- end -}}
			{{if .Origin}}
				origin = "{{.Origin}}"
			{{- end -}}
			{{if .GivenName}}
				given_name = "{{.GivenName}}"
			{{- end -}}
			{{if .FamilyName}}
				family_name = "{{.FamilyName}}"
			{{- end -}}
			{{if .Groups}}
				groups = "{{.Groups}}"
			{{- end -}}
			{{if .Email}}
				email = "{{.Email}}"
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
	return urmp.HclType + ` "cloudfoundry_user" ` + urmp.HclObjectName + ` {}`
}

func TestUserResource_Configure(t *testing.T) {
	t.Parallel()
	var (
		createUsername   = "tf-test"
		createEmail      = "tf-test@example.com"
		createPassword   = "tf-test"
		familyName       = "tf-family"
		givenName        = "tf-given"
		updateUsername   = "tf-test2"
		updateEmail      = "tf-test-updated@example.com"
		resourceName     = "cloudfoundry_user.us"
		createUsername2  = "tf-test3"
		createPassword2  = "tf-test3"
		existingUsername = "test"
		testInvalidLabel = `{"purpose@!": "testing", landscape: "test"}`
	)
	t.Run("happy path - create/update/import/delete user", func(t *testing.T) {

		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_user_crud")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclResourceUser(&UserResourceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "us",
						UserName:      &createUsername,
						Password:      &createPassword,
						GivenName:     &givenName,
						FamilyName:    &familyName,
						Email:         &createEmail,
						Labels:        &testCreateLabel,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckNoResourceAttr(resourceName, "annotations"),
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestCheckResourceAttr(resourceName, "username", createUsername),
						resource.TestCheckResourceAttr(resourceName, "given_name", givenName),
						resource.TestCheckResourceAttr(resourceName, "family_name", familyName),
						resource.TestCheckResourceAttr(resourceName, "email", createEmail),
						resource.TestCheckResourceAttr(resourceName, "labels.purpose", "testing"),
					),
				},
				{
					Config: hclProvider(nil) + hclResourceUser(&UserResourceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "us",
						UserName:      &updateUsername,
						Password:      &createPassword,
						Email:         &updateEmail,
						Labels:        &testUpdateLabel,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestCheckResourceAttr(resourceName, "username", updateUsername),
						resource.TestCheckResourceAttr(resourceName, "email", updateEmail),
						resource.TestCheckResourceAttr(resourceName, "labels.purpose", "production"),
						resource.TestCheckResourceAttr(resourceName, "labels.%", "2"),
					),
				},
				{
					ResourceName:            resourceName,
					ImportStateIdFunc:       getIdForImport(resourceName),
					ImportStateVerifyIgnore: []string{"password"},
					ImportState:             true,
					ImportStateVerify:       true,
				},
			},
		})
	})
	t.Run("error path - invalid create/update scenarios", func(t *testing.T) {

		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_user_crud_invalid")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclResourceUser(&UserResourceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "us",
					}),
					ExpectError: regexp.MustCompile(`Missing required argument`),
				},
				{
					Config: hclProvider(nil) + hclResourceUser(&UserResourceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "us",
						UserName:      &createUsername2,
					}),
					ExpectError: regexp.MustCompile(`API Error Creating User in Origin`),
				},
				{
					Config: hclProvider(nil) + hclResourceUser(&UserResourceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "us",
						UserName:      &createUsername2,
						Password:      &createPassword2,
						Labels:        &testInvalidLabel,
					}),
					ExpectError: regexp.MustCompile(`API Error Creating CF User`),
				},
				{
					Config: hclProvider(nil) + hclResourceUser(&UserResourceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "us",
						UserName:      &createUsername2,
						Password:      &createPassword2,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestCheckResourceAttr(resourceName, "username", createUsername2),
						resource.TestCheckResourceAttr(resourceName, "password", createPassword2),
					),
				},
				{
					Config: hclProvider(nil) + hclResourceUser(&UserResourceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "us",
					}),
					ExpectError: regexp.MustCompile(`Missing required argument`),
				},
				{
					Config: hclProvider(nil) + hclResourceUser(&UserResourceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "us",
						UserName:      &createUsername2,
						Password:      &createPassword,
					}),
					ExpectError: regexp.MustCompile(`API Error Updating Password of User`),
				},
				{
					Config: hclProvider(nil) + hclResourceUser(&UserResourceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "us",
						UserName:      &existingUsername,
						Password:      &createPassword2,
					}),
					ExpectError: regexp.MustCompile(`API Error Updating User in Origin`),
				},
				{
					Config: hclProvider(nil) + hclResourceUser(&UserResourceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "us",
						UserName:      &createUsername2,
						Password:      &createPassword2,
						Labels:        &testInvalidLabel,
					}),
					ExpectError: regexp.MustCompile(`API Error Updating CF User`),
				},
			},
		})
	})
}
