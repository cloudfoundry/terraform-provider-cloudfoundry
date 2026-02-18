package provider

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	cfv3resource "github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/assert"
)

type ServiceInstanceSharingResourceModelPtr struct {
	HclType         string
	HclObjectName   string
	ServiceInstance *string
	Spaces          *string
}

func hclResourceServiceInstanceSharing(model *ServiceInstanceSharingResourceModelPtr) string {
	if model != nil {
		s := `
		{{.HclType}} "cloudfoundry_service_instance_sharing" {{.HclObjectName}} {
			{{- if .ServiceInstance}}
				service_instance = "{{.ServiceInstance}}"
			{{- end -}}
			{{ if .Spaces}}
				spaces = {{.Spaces}}
			{{- end }}
		}`

		tmpl, err := template.New("resource_service_instance_sharing").Parse(s)
		if err != nil {
			panic(err)
		}
		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, model)
		if err != nil {
			panic(err)
		}
		return buf.String()
	}
	return model.HclType + ` "cloudfoundry_service_instance_sharing" ` + model.HclObjectName + ` {}`
}

func TestServiceInstanceSharingResource_Configure(t *testing.T) {
	var (
		testUserProvidedServiceInstanceGUID = "5e2976bb-332e-41e1-8be3-53baafea9296"
		testSpaces                          = `["02c0cc92-6ecc-44b1-b7b2-096ca19ee143", "121c3a95-0f82-45a6-8ff2-1920b2067edb"]`
		updatedSpaces                       = `["592c4741-9fb2-4b5d-8848-4f7db2d2c4c4", "02c0cc92-6ecc-44b1-b7b2-096ca19ee143"]`
	)
	t.Parallel()
	t.Run("happy path - create and update service instance sharing", func(t *testing.T) {
		// setup
		resourceName := "cloudfoundry_service_instance_sharing.rs"
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_service_instance_sharing")
		defer stopQuietly(rec)

		// actual test
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclResourceServiceInstanceSharing(&ServiceInstanceSharingResourceModelPtr{
						HclType:         hclObjectResource,
						HclObjectName:   "rs",
						ServiceInstance: new(testUserProvidedServiceInstanceGUID),
						Spaces:          &testSpaces,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "service_instance", regexpValidUUID),
						resource.TestCheckResourceAttr(resourceName, "spaces.#", "2"),
						resource.TestCheckTypeSetElemAttr(resourceName, "spaces.*", "121c3a95-0f82-45a6-8ff2-1920b2067edb"),
						resource.TestCheckTypeSetElemAttr(resourceName, "spaces.*", "02c0cc92-6ecc-44b1-b7b2-096ca19ee143"),
					),
				},
				{
					// Update - remove one space and add another
					Config: hclProvider(nil) + hclResourceServiceInstanceSharing(&ServiceInstanceSharingResourceModelPtr{
						HclType:         hclObjectResource,
						HclObjectName:   "rs",
						ServiceInstance: new(testUserProvidedServiceInstanceGUID),
						Spaces:          &updatedSpaces,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "service_instance", regexpValidUUID),
						resource.TestCheckResourceAttr(resourceName, "spaces.#", "2"),
						resource.TestCheckTypeSetElemAttr(resourceName, "spaces.*", "592c4741-9fb2-4b5d-8848-4f7db2d2c4c4"),
						resource.TestCheckTypeSetElemAttr(resourceName, "spaces.*", "02c0cc92-6ecc-44b1-b7b2-096ca19ee143"),
					),
				},
			},
		})
	})

	t.Run("error path - create instance sharing with missing space", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_service_instance_sharing_missing_space")

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclResourceServiceInstanceSharing(&ServiceInstanceSharingResourceModelPtr{
						HclType:         hclObjectResource,
						HclObjectName:   "rs",
						ServiceInstance: new(testUserProvidedServiceInstanceGUID),
						Spaces:          &testSpaces,
					}),
					ExpectError: regexp.MustCompile(`Error sharing service instance with space`),
				},
			},
		})
	})

	t.Run("error path - create instance sharing with missing service instance", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_service_instance_sharing_missing_service_instance")

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclResourceServiceInstanceSharing(&ServiceInstanceSharingResourceModelPtr{
						HclType:         hclObjectResource,
						HclObjectName:   "rs",
						ServiceInstance: new(testUserProvidedServiceInstanceGUID),
						Spaces:          &testSpaces,
					}),
					ExpectError: regexp.MustCompile(`Error sharing service instance with space`),
				},
			},
		})
	})
}

func TestMapSharedSpacesValuesToType(t *testing.T) {
	spaceGUID1 := "space-guid-1"
	spaceGUID2 := "space-guid-2"
	sharedSpaces := []attr.Value{types.StringValue(spaceGUID1), types.StringValue(spaceGUID2)}
	serviceInstance := "service-instance-guid-1"
	relationship := &cfv3resource.ServiceInstanceSharedSpaceRelationships{
		Data: []cfv3resource.Relationship{
			{GUID: spaceGUID1}, {GUID: spaceGUID2},
		},
	}
	spaces := types.SetValueMust(types.StringType, sharedSpaces)
	expected := ServiceInstanceSharingType{
		Id:              types.StringValue(serviceInstance),
		ServiceInstance: types.StringValue(serviceInstance),
		Spaces:          spaces,
	}

	result := mapSharedSpacesValuesToType(relationship, serviceInstance)

	assert.Equal(t, expected, result)
}

func TestServiceInstanceSharingResource_Import(t *testing.T) {
	var (
		testUserProvidedServiceInstanceGUID = "5e2976bb-332e-41e1-8be3-53baafea9296"
		testSpaces                          = `["02c0cc92-6ecc-44b1-b7b2-096ca19ee143", "121c3a95-0f82-45a6-8ff2-1920b2067edb"]`
	)
	t.Parallel()

	// setup
	resourceName := "cloudfoundry_service_instance_sharing.rs"
	cfg := getCFHomeConf()
	rec := cfg.SetupVCR(t, "fixtures/resource_service_instance_sharing")
	defer stopQuietly(rec)

	// actual test
	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
		Steps: []resource.TestStep{
			{
				Config: hclProvider(nil) + hclResourceServiceInstanceSharing(&ServiceInstanceSharingResourceModelPtr{
					HclType:         hclObjectResource,
					HclObjectName:   "rs",
					ServiceInstance: new(testUserProvidedServiceInstanceGUID),
					Spaces:          &testSpaces,
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "service_instance", regexpValidUUID),
					resource.TestMatchResourceAttr(resourceName, "spaces.0", regexpValidUUID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
