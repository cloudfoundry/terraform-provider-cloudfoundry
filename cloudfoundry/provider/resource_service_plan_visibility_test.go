package provider

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type ServicePlanVisibilityModelPtr struct {
	HclType       string
	HclObjectName string
	ServicePlan   *string
	Organizations *string
	Type          *string
}

func hclServicePlanVisibility(spv *ServicePlanVisibilityModelPtr) string {
	if spv != nil {
		hclTemplate := `
		{{.HclType}} "cloudfoundry_service_plan_visibility" {{.HclObjectName}} {
			{{- if .ServicePlan}}
			service_plan = "{{.ServicePlan}}"
			{{- end -}}
			{{- if .Organizations}}
			organizations = [{{.Organizations}}]
			{{- end -}}
			{{- if .Type}}
			type = "{{.Type}}"
			{{- end }}
		}`
		tmpl, err := template.New("servicePlanVisibility").Parse(hclTemplate)
		if err != nil {
			panic(err)
		}
		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, spv)
		if err != nil {
			panic(err)
		}
		return buf.String()
	}
	return spv.HclType + ` "cloudfoundry_service_plan_visibility" ` + spv.HclObjectName + ` {}`
}

func TestServicePlanVisibility_Configure(t *testing.T) {
	var (
		testServicePlanGUID                 = "f37176d7-39eb-4e80-a3c0-328dfe36902c"
		testOrganizations                   = `"3533be5d-272f-42fe-bf70-fc4b108c2043"`
		testServicePlanVisibilityType       = "organization"
		testServicePlanVisibilityTypePublic = "public"
		invalidServicePlanGUID              = "invalid-02c0cc92-6ecc-44b1-b7b2-096ca19ee143"
	)

	t.Parallel()

	t.Run("happy path - create service plan visibility", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_service_plan_visibility_create")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclServicePlanVisibility(&ServicePlanVisibilityModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "spv",
						ServicePlan:   new(testServicePlanGUID),
						Organizations: new(testOrganizations),
						Type:          new(testServicePlanVisibilityType),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("cloudfoundry_service_plan_visibility.spv", "service_plan", testServicePlanGUID),
						resource.TestCheckResourceAttr("cloudfoundry_service_plan_visibility.spv", "organizations.0", "3533be5d-272f-42fe-bf70-fc4b108c2043"),
						resource.TestCheckResourceAttr("cloudfoundry_service_plan_visibility.spv", "type", testServicePlanVisibilityType),
					),
				},
				{
					Config: hclProvider(nil) + hclServicePlanVisibility(&ServicePlanVisibilityModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "spv",
						ServicePlan:   new(testServicePlanGUID),
						Type:          new(testServicePlanVisibilityTypePublic),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("cloudfoundry_service_plan_visibility.spv", "service_plan", testServicePlanGUID),
						resource.TestCheckResourceAttr("cloudfoundry_service_plan_visibility.spv", "type", "public"),
					),
				},
			},
		})
	})

	t.Run("error path - invalid service plan", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_service_plan_visibility_invalid_service_plan")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclServicePlanVisibility(&ServicePlanVisibilityModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "spv_invalid",
						ServicePlan:   new(invalidServicePlanGUID),
						Organizations: new(testOrganizations),
						Type:          new(testServicePlanVisibilityType),
					}),
					ExpectError: regexp.MustCompile("Error: API Error Creating Service Plan Visibility"),
				},
			},
		})
	})

	t.Run("error path - invalid type", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_service_plan_visibility_invalid_type")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclServicePlanVisibility(&ServicePlanVisibilityModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "spv_invalid_type",
						ServicePlan:   new(testServicePlanGUID),
						Type:          new("invalid-type"),
					}),
					ExpectError: regexp.MustCompile("Error: Invalid Attribute Value Match"),
				},
			},
		})
	})
}
