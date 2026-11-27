package provider

import (
	"bytes"
	"net/http"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type EnvVarGroupModelPtr struct {
	HclType       string
	HclObjectName string
	Name          *string
	Variables     *string
	UpdatedAt     *string
}

func hclEnvVarGroup(emp *EnvVarGroupModelPtr) string {
	if emp != nil {
		s := `
		{{.HclType}} "cloudfoundry_env_var_group" {{.HclObjectName}} {
			{{- if .Name}}
				name = "{{.Name}}"
			{{- end -}}
			{{if .Variables}}
				variables = {{.Variables}}
			{{- end -}}
			{{if .UpdatedAt}}
				updated_at = "{{.UpdatedAt}}"
			{{- end }}
			}`
		tmpl, err := template.New("resource_env_var_group").Parse(s)
		if err != nil {
			panic(err)
		}
		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, emp)
		if err != nil {
			panic(err)
		}
		return buf.String()
	}
	return emp.HclType + ` "cloudfoundry_env_var_group" ` + emp.HclObjectName + ` {}`
}

func TestEnvVarGroupResource(t *testing.T) {
	t.Parallel()
	testRunningGroupName := "running"
	testStagingGroupName := "staging"
	testVarsCreate := `{"HTTP_PROXY" = "http://proxy.example.com:8080", "NO_PROXY" = "localhost"}`
	testVarsUpdate := `{"HTTP_PROXY" = "http://proxy.example.com:8080", "HTTPS_PROXY" = "https://proxy.example.com:8080"}`
	resourceName := "cloudfoundry_env_var_group.running"

	t.Run("happy path - running group create/update/import/delete", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_env_var_group_running")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclEnvVarGroup(&EnvVarGroupModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "running",
						Name:          &testRunningGroupName,
						Variables:     &testVarsCreate,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "name", "running"),
						resource.TestCheckResourceAttr(resourceName, "variables.HTTP_PROXY", "http://proxy.example.com:8080"),
						resource.TestCheckResourceAttr(resourceName, "variables.NO_PROXY", "localhost"),
						resource.TestMatchResourceAttr(resourceName, "updated_at", regexpValidRFC3999Format),
					),
				},
				{
					Config: hclProvider(nil) + hclEnvVarGroup(&EnvVarGroupModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "running",
						Name:          &testRunningGroupName,
						Variables:     &testVarsUpdate,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "name", "running"),
						resource.TestCheckResourceAttr(resourceName, "variables.HTTP_PROXY", "http://proxy.example.com:8080"),
						resource.TestCheckResourceAttr(resourceName, "variables.HTTPS_PROXY", "https://proxy.example.com:8080"),
						resource.TestCheckNoResourceAttr(resourceName, "variables.NO_PROXY"),
					),
				},
				{
					ResourceName:                         resourceName,
					ImportStateId:                        "running",
					ImportState:                          true,
					ImportStateVerify:                    true,
					ImportStateVerifyIgnore:              []string{"variables"},
					ImportStateVerifyIdentifierAttribute: "name",
				},
			},
		})
	})

	t.Run("happy path - staging group create/delete", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_env_var_group_staging")
		defer stopQuietly(rec)

		stagingResourceName := "cloudfoundry_env_var_group.staging"
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclEnvVarGroup(&EnvVarGroupModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "staging",
						Name:          &testStagingGroupName,
						Variables:     &testVarsCreate,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(stagingResourceName, "name", "staging"),
						resource.TestCheckResourceAttr(stagingResourceName, "variables.HTTP_PROXY", "http://proxy.example.com:8080"),
					),
				},
			},
		})
	})

	t.Run("error path - invalid group name", func(t *testing.T) {
		invalidName := "invalid"
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(http.DefaultClient),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclEnvVarGroup(&EnvVarGroupModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rs",
						Name:          &invalidName,
					}),
					ExpectError: regexp.MustCompile(`Attribute name value must be one of:`),
				},
			},
		})
	})

	_ = resourceName
}
