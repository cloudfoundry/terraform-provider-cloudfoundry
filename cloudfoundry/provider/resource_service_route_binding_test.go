package provider

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type ServiceRouteBindingModelPtr struct {
	HclType         string
	HclObjectName   string
	Id              *string
	Labels          *string
	Annotations     *string
	Parameters      *string
	RouteServiceURL *string
	LastOperation   *string
	CreatedAt       *string
	UpdatedAt       *string
	Route           *string
	ServiceInstance *string
}

func hclServiceRouteBinding(sip *ServiceRouteBindingModelPtr) string {
	if sip != nil {
		s := `
		{{.HclType}} "cloudfoundry_service_route_binding" {{.HclObjectName}} {
			{{- if .Id}}
				id = "{{.Id}}"
			{{- end -}}
			{{if .LastOperation}}
				last_operation = {{.LastOperation}}
			{{- end -}}
			{{if .CreatedAt}}
				created_at = {{.CreatedAt}}
			{{- end -}}
			{{if .UpdatedAt}}
				updated_at = {{.UpdatedAt}}
			{{- end -}}
			{{if .Route}}
				route = "{{.Route}}"
			{{- end -}}
			{{if .ServiceInstance}}
				service_instance = "{{.ServiceInstance}}"
			{{- end -}}
			{{if .Labels}}
				labels = {{.Labels}}
			{{- end -}}
			{{if .Annotations}}
				annotations = {{.Annotations}}
			{{- end -}}
			{{if .Parameters}}
				parameters = <<EOT
				{{.Parameters}}
				EOT
			{{- end -}}
			{{if .RouteServiceURL}}
				route_service_url = "{{.RouteServiceURL}}"
			{{- end }}
		}`
		tmpl, err := template.New("service_route_binding").Parse(s)
		if err != nil {
			panic(err)
		}
		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, sip)
		if err != nil {
			panic(err)
		}
		return buf.String()
	}
	return sip.HclType + ` "cloudfoundry_service_route_binding" ` + sip.HclObjectName + ` {}`
}

func TestResourceServiceRouteBinding(t *testing.T) {
	var (
		// in canary -> PerformanceTeamBLR -> tf-space-1
		resourceName           = "cloudfoundry_service_route_binding.si"
		testRouteGUID          = "3966c2fb-d84d-462d-82a5-a81cf7cdab20"
		testRouteGUID2         = "490d6825-5d8f-4dd2-b332-1e8ea6ae5158"
		testServiceUPSGuid     = "3a8588f9-f846-444f-ab9e-48282f06449b"
		testServiceManagedGuid = "a92e1186-b229-4711-b233-a8726879dad6"
	)
	t.Parallel()
	t.Run("happy path - create route binding with a user provided instance", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_service_route_binding_user_provided")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclServiceRouteBinding(&ServiceRouteBindingModelPtr{
						HclType:         hclObjectResource,
						HclObjectName:   "si",
						Route:           &testRouteGUID,
						ServiceInstance: &testServiceUPSGuid,
						Labels:          &testCreateLabel,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "route", testRouteGUID),
						resource.TestCheckResourceAttr(resourceName, "service_instance", testServiceUPSGuid),
						resource.TestCheckResourceAttr(resourceName, "labels.purpose", "testing"),
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestMatchResourceAttr(resourceName, "updated_at", regexpValidRFC3999Format),
					),
				},
				{
					Config: hclProvider(nil) + hclServiceRouteBinding(&ServiceRouteBindingModelPtr{
						HclType:         hclObjectResource,
						HclObjectName:   "si",
						Route:           &testRouteGUID,
						ServiceInstance: &testServiceUPSGuid,
						Labels:          &testUpdateLabel,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "route", testRouteGUID),
						resource.TestCheckResourceAttr(resourceName, "service_instance", testServiceUPSGuid),
						resource.TestCheckResourceAttr(resourceName, "labels.purpose", "production"),
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestMatchResourceAttr(resourceName, "updated_at", regexpValidRFC3999Format),
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
	t.Run("happy path - create route binding with a managed instance", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_service_route_binding_managed")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclServiceRouteBinding(&ServiceRouteBindingModelPtr{
						HclType:         hclObjectResource,
						HclObjectName:   "si",
						Route:           &testRouteGUID2,
						ServiceInstance: &testServiceManagedGuid,
						Labels:          &testCreateLabel,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "route", testRouteGUID2),
						resource.TestCheckResourceAttr(resourceName, "service_instance", testServiceManagedGuid),
						resource.TestCheckResourceAttr(resourceName, "labels.purpose", "testing"),
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestMatchResourceAttr(resourceName, "updated_at", regexpValidRFC3999Format),
					),
				},
				{
					Config: hclProvider(nil) + hclServiceRouteBinding(&ServiceRouteBindingModelPtr{
						HclType:         hclObjectResource,
						HclObjectName:   "si",
						Route:           &testRouteGUID2,
						ServiceInstance: &testServiceManagedGuid,
						Labels:          &testUpdateLabel,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "route", testRouteGUID2),
						resource.TestCheckResourceAttr(resourceName, "service_instance", testServiceManagedGuid),
						resource.TestCheckResourceAttr(resourceName, "labels.purpose", "production"),
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestMatchResourceAttr(resourceName, "updated_at", regexpValidRFC3999Format),
					),
				},
				{
					ResourceName:    resourceName,
					ImportState:     true,
					ImportStateKind: resource.ImportBlockWithResourceIdentity,
				},
			},
		})
	})

	t.Run("error path - create route binding with invalid instance", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_service_route_binding_invalid")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclServiceRouteBinding(&ServiceRouteBindingModelPtr{
						HclType:         hclObjectResource,
						HclObjectName:   "si",
						Route:           &testRouteGUID,
						ServiceInstance: &invalidOrgGUID,
						Labels:          &testCreateLabel,
					}),
					ExpectError: regexp.MustCompile(`API Error in creating service Route Binding`),
				},
			},
		})
	})
}
