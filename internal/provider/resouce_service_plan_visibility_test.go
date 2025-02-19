package provider

import (
	"bytes"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type ServicePlanVisibilityModel struct {
	HclType       string
	HclObjectName string
	Type          *string
	Organizations *string
	SpaceGUID     *string
	Id            *string
	CreatedAt     *string
	UpdatedAt     *string
}

func hclServicePlanVisibility(spvm *ServicePlanVisibilityModel) string {
	if spvm != nil {
		s := `
		{{.HclType}} "cloudfoundry_service_plan_visibility" {{.HclObjectName}} {
			{{- if .Type}}
				type = "{{.Type}}"
			{{- end -}}
			{{if .Id}}
				id = "{{.Id}}"
			{{- end -}}
			{{if .Organizations}}
				organizations = {{.Organizations}}
			{{- end -}}
			{{if .SpaceGUID}}
				space_guid = "{{.SpaceGUID}}"
			{{- end -}}
			{{if .CreatedAt}}
				created_at = "{{.CreatedAt}}"
			{{- end -}}
			{{if .UpdatedAt}}
				updated_at = "{{.UpdatedAt}}"
			{{- end }}
		}`
		tmpl, err := template.New("resource_service_plan_visibility").Parse(s)
		if err != nil {
			panic(err)
		}
		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, spvm)
		if err != nil {
			panic(err)
		}
		return buf.String()
	}
	return spvm.HclType + ` "cloudfoundry_service_plan_visibility" ` + spvm.HclObjectName + ` {}`
}

func TestServicePlanVisibility_Configure(t *testing.T) {
	var (
		resourceName    = "cloudfoundry_service_plan_visibility.rs"
		visibilityType  = "organization"
		organizationIDs = "[\"749899b9-c991-4890-97a5-de04c6a4745f\"]"
		spaceGUID       = "06ed9592-56ae-4dde-86f9-fc04f4d1bb9d	"
	)
	t.Parallel()
	t.Run("happy path - create/update/import service plan visibility", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_service_plan_visibility")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclServicePlanVisibility(&ServicePlanVisibilityModel{
						HclType:       hclObjectResource,
						HclObjectName: "rs",
						Type:          &visibilityType,
						Organizations: &organizationIDs,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestCheckResourceAttr(resourceName, "type", visibilityType),
						resource.TestCheckResourceAttr(resourceName, "organizations.#", "2"),
					),
				},
				{
					Config: hclProvider(nil) + hclServicePlanVisibility(&ServicePlanVisibilityModel{
						HclType:       hclObjectResource,
						HclObjectName: "rs",
						Type:          &visibilityType,
						SpaceGUID:     &spaceGUID,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestCheckResourceAttr(resourceName, "type", visibilityType),
						resource.TestCheckResourceAttr(resourceName, "space_guid", spaceGUID),
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
}
