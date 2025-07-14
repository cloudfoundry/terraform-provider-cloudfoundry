package provider

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type SecurityGroupSpacesModelPtr struct {
	HclType       string
	HclObjectName string
	SecurityGroup *string
	RunningSpaces *string
	StagingSpaces *string
}

func hclSecurityGroupSpaces(sgsmp *SecurityGroupSpacesModelPtr) string {
	if sgsmp != nil {
		s := `
		{{.HclType}} "cloudfoundry_security_group_space_bindings" {{.HclObjectName}} {
			{{if .SecurityGroup}}
				security_group = "{{.SecurityGroup}}"
			{{- end -}}
			{{if .RunningSpaces}}
				running_spaces = {{.RunningSpaces}}
			{{- end -}}
			{{if .StagingSpaces}}
				staging_spaces = {{.StagingSpaces}}
			{{- end }}
			}`
		tmpl, err := template.New("resource_security_group_space_bindings").Parse(s)
		if err != nil {
			panic(err)
		}
		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, sgsmp)
		if err != nil {
			panic(err)
		}
		return buf.String()
	}
	return sgsmp.HclType + ` "cloudfoundry_security_group_space_bindings" ` + sgsmp.HclObjectName + ` {}`
}

func TestSecurityGroupSpacesResourceConfigure(t *testing.T) {
	var (
		resourceName      = "cloudfoundry_security_group_space_bindings.ds"
		securityGroup     = "56eedab7-cb97-469b-a3e9-89521827c039"
		runningSpacesBind = `["121c3a95-0f82-45a6-8ff2-1920b2067edb","02c0cc92-6ecc-44b1-b7b2-096ca19ee143"]`
		stagingSpacesBind = `["02c0cc92-6ecc-44b1-b7b2-096ca19ee143"]`
	)
	t.Parallel()
	t.Run("happy path - bind/unbind security ", func(t *testing.T) {

		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_security_group_spaces_crud")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclSecurityGroupSpaces(&SecurityGroupSpacesModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "ds",
						SecurityGroup: &securityGroup,
						RunningSpaces: &runningSpacesBind,
						StagingSpaces: &stagingSpacesBind,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "security_group", securityGroup),
						resource.TestCheckResourceAttr(resourceName, "running_spaces.#", "2"),
						resource.TestCheckResourceAttr(resourceName, "staging_spaces.#", "1"),
					),
				},
				{
					Config: hclProvider(nil) + hclSecurityGroupSpaces(&SecurityGroupSpacesModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "ds",
						SecurityGroup: &securityGroup,
						RunningSpaces: &stagingSpacesBind,
						StagingSpaces: &runningSpacesBind,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "security_group", securityGroup),
						resource.TestCheckResourceAttr(resourceName, "running_spaces.#", "1"),
						resource.TestCheckResourceAttr(resourceName, "staging_spaces.#", "2"),
					),
				},
			},
		})
	})
	t.Run("error path - invalid security group/space", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_security_group_spaces_invalid_security_group")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclSecurityGroupSpaces(&SecurityGroupSpacesModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "ds_invalid_name",
						SecurityGroup: &invalidOrgGUID,
						RunningSpaces: &runningSpacesBind,
						StagingSpaces: &stagingSpacesBind,
					}),
					ExpectError: regexp.MustCompile(`API Error Binding Running Security Group`),
				},
			},
		})
	})
	t.Run("error path - invalid space", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_security_group_spaces_invalid_space")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclSecurityGroupSpaces(&SecurityGroupSpacesModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "ds_invalid_name",
						SecurityGroup: &securityGroup,
						RunningSpaces: &stagingSpaces,
						StagingSpaces: &runningSpaces,
					}),
					ExpectError: regexp.MustCompile(`API Error Binding Running Security Group`),
				},
			},
		})
	})

}
