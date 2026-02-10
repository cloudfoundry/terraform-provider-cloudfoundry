package provider

import (
	"bytes"
	"fmt"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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
					ResourceName:      resourceName,
					ImportStateIdFunc: getIdForImport(resourceName),
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})

	// This test verifies that updating a space's allow_ssh attribute does not cause
	// route bindings to be replaced through a cascading dependency chain:
	// space -> service_instance + route -> route_binding. When the space is updated,
	// cloudfoundry_space.test.id becomes "known after apply", which flows into both
	// the service instance's space and the route's space, making their IDs also
	// "known after apply", which then flows into the route binding's service_instance
	// and route attributes. UseStateForUnknown at each level prevents the cascade
	// from triggering replacements.
	t.Run("happy path - route binding not replaced when space allow_ssh changes", func(t *testing.T) {
		resourceName := "cloudfoundry_service_route_binding.si_stability"
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_service_route_binding_space_allow_ssh_update")
		defer stopQuietly(rec)

		var bindingID string

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					// Step 1: Create the full dependency chain:
					// space -> service_instance + route -> route_binding
					Config: hclProvider(nil) + `
resource "cloudfoundry_space" "test" {
	name      = "test-space-route-binding-stability"
	org       = "` + testOrgGUID + `"
	allow_ssh = false
}
resource "cloudfoundry_service_instance" "test" {
	name         = "test-si-route-binding-stability"
	type         = "user-provided"
	space        = cloudfoundry_space.test.id
	route_service_url = "https://example.com"
}
resource "cloudfoundry_route" "test" {
	space  = cloudfoundry_space.test.id
	domain = "` + testDomainRouteGUID + `"
	host   = "route-binding-stability-test"
}
resource "cloudfoundry_service_route_binding" "si_stability" {
	service_instance = cloudfoundry_service_instance.test.id
	route            = cloudfoundry_route.test.id
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						func(s *terraform.State) error {
							rs, ok := s.RootModule().Resources[resourceName]
							if !ok {
								return fmt.Errorf("resource not found: %s", resourceName)
							}
							bindingID = rs.Primary.ID
							return nil
						},
					),
				},
				{
					// Step 2: Change allow_ssh on the space. This causes a cascade:
					// space updated -> space.id "known after apply" ->
					// service_instance.space + route.space "known after apply" ->
					// service_instance.id + route.id "known after apply" ->
					// route_binding.service_instance + route_binding.route
					// "known after apply". Without UseStateForUnknown at each
					// level, this cascade would trigger replacements.
					Config: hclProvider(nil) + `
resource "cloudfoundry_space" "test" {
	name      = "test-space-route-binding-stability"
	org       = "` + testOrgGUID + `"
	allow_ssh = true
}
resource "cloudfoundry_service_instance" "test" {
	name         = "test-si-route-binding-stability"
	type         = "user-provided"
	space        = cloudfoundry_space.test.id
	route_service_url = "https://example.com"
}
resource "cloudfoundry_route" "test" {
	space  = cloudfoundry_space.test.id
	domain = "` + testDomainRouteGUID + `"
	host   = "route-binding-stability-test"
}
resource "cloudfoundry_service_route_binding" "si_stability" {
	service_instance = cloudfoundry_service_instance.test.id
	route            = cloudfoundry_route.test.id
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						func(s *terraform.State) error {
							rs, ok := s.RootModule().Resources[resourceName]
							if !ok {
								return fmt.Errorf("resource not found: %s", resourceName)
							}
							if rs.Primary.ID != bindingID {
								return fmt.Errorf("route binding was unexpectedly replaced: old ID %s, new ID %s", bindingID, rs.Primary.ID)
							}
							return nil
						},
					),
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
