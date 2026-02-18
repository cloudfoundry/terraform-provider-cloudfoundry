package provider

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type ServiceBrokerModelPtr struct {
	HclType       string
	HclObjectName string
	Name          *string
	Url           *string
	Space         *string
	Username      *string
	Password      *string
	Id            *string
	Labels        *string
	Annotations   *string
	CreatedAt     *string
	UpdatedAt     *string
}

func hclServiceBroker(sbmp *ServiceBrokerModelPtr) string {
	if sbmp != nil {
		s := `
		{{.HclType}} "cloudfoundry_service_broker" {{.HclObjectName}} {
			{{- if .Name}}
				name = "{{.Name}}"
			{{- end -}}
			{{if .Id}}
				id = "{{.Id}}"
			{{- end -}}
			{{if .Space}}
				space = "{{.Space}}"
			{{- end -}}
			{{if .Url}}
				url = "{{.Url}}"
			{{- end -}}
			{{if .Username}}
				username = "{{.Username}}"
			{{- end -}}
			{{if .Password}}
				password = "{{.Password}}"
			{{- end -}}
			{{if .Labels}}
				labels = {{.Labels}}
			{{- end -}}
			{{if .Annotations}}
				annotations = {{.Annotations}}
			{{- end -}}
			{{if .CreatedAt}}
				created_at = "{{.CreatedAt}}"
			{{- end -}}
			{{if .UpdatedAt}}
				updated_at = "{{.UpdatedAt}}"
			{{- end }}
			}`
		tmpl, err := template.New("resource_service_broker").Parse(s)
		if err != nil {
			panic(err)
		}
		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, sbmp)
		if err != nil {
			panic(err)
		}
		return buf.String()
	}
	return sbmp.HclType + ` "cloudfoundry_service_broker" ` + sbmp.HclObjectName + ` {}`
}

func TestServiceBroker_Configure(t *testing.T) {
	var (
		resourceName           = "cloudfoundry_service_broker.rs"
		brokerName             = "broker"
		brokerName2            = "brokerName2"
		brokerNameUpdated      = "broker-2"
		brokerURL              = "https://sample-broker.cert.cfapps.stagingazure.hanavlab.ondemand.com"
		brokerUsername         = "admin"
		brokerPassword         = "hi"
		brokerSpace            = "0925b3c7-7544-4700-b71b-191b3c348e5c"
		spaceBrokerName        = "space-broker"
		spaceBrokerNameUpdated = "space-broker-2"
		existingBrokerName     = "hi"
	)
	t.Parallel()
	t.Run("happy path - create/update/import service broker", func(t *testing.T) {

		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_service_broker")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclServiceBroker(&ServiceBrokerModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rs",
						Name:          &brokerName,
						Url:           &brokerURL,
						Username:      &brokerUsername,
						Password:      &brokerPassword,
						Labels:        &testCreateLabel,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestCheckResourceAttr(resourceName, "name", brokerName),
						resource.TestCheckResourceAttr(resourceName, "url", brokerURL),
						resource.TestCheckResourceAttr(resourceName, "username", brokerUsername),
						resource.TestCheckResourceAttr(resourceName, "labels.purpose", "testing"),
					),
				},
				{
					Config: hclProvider(nil) + hclServiceBroker(&ServiceBrokerModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rs",
						Name:          &brokerNameUpdated,
						Url:           &brokerURL,
						Username:      &brokerUsername,
						Password:      &brokerPassword,
						Labels:        &testUpdateLabel,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestCheckResourceAttr(resourceName, "name", brokerNameUpdated),
						resource.TestCheckResourceAttr(resourceName, "url", brokerURL),
						resource.TestCheckResourceAttr(resourceName, "username", brokerUsername),
						resource.TestCheckResourceAttr(resourceName, "labels.purpose", "production"),
					),
				},
				{
					ResourceName:            resourceName,
					ImportStateIdFunc:       getIdForImport(resourceName),
					ImportStateVerifyIgnore: []string{"username", "password"},
					ImportState:             true,
					ImportStateVerify:       true,
				},
			},
		})
	})

	t.Run("happy path - create/update/import space-scoped service broker", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_service_broker_space")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclServiceBroker(&ServiceBrokerModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rs",
						Name:          &spaceBrokerName,
						Url:           &brokerURL,
						Username:      &brokerUsername,
						Password:      &brokerPassword,
						Labels:        &testCreateLabel,
						Space:         &brokerSpace,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestCheckResourceAttr(resourceName, "name", spaceBrokerName),
						resource.TestCheckResourceAttr(resourceName, "url", brokerURL),
						resource.TestCheckResourceAttr(resourceName, "username", brokerUsername),
						resource.TestCheckResourceAttr(resourceName, "space", brokerSpace),
						resource.TestCheckResourceAttr(resourceName, "labels.purpose", "testing"),
					),
				},
				{
					Config: hclProvider(nil) + hclServiceBroker(&ServiceBrokerModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rs",
						Name:          &spaceBrokerNameUpdated,
						Url:           &brokerURL,
						Username:      &brokerUsername,
						Password:      &brokerPassword,
						Labels:        &testUpdateLabel,
						Space:         &brokerSpace,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestCheckResourceAttr(resourceName, "name", spaceBrokerNameUpdated),
						resource.TestCheckResourceAttr(resourceName, "url", brokerURL),
						resource.TestCheckResourceAttr(resourceName, "username", brokerUsername),
						resource.TestCheckResourceAttr(resourceName, "space", brokerSpace),
						resource.TestCheckResourceAttr(resourceName, "labels.purpose", "production"),
					),
				},
				{
					ResourceName:            resourceName,
					ImportStateIdFunc:       getIdForImport(resourceName),
					ImportStateVerifyIgnore: []string{"username", "password"},
					ImportState:             true,
					ImportStateVerify:       true,
				},
			},
		})
	})
	t.Run("error path - create/update service broker with existing name", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_service_broker_space_invalid")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclServiceBroker(&ServiceBrokerModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rs",
						Name:          &existingBrokerName,
						Url:           &brokerURL,
						Username:      &brokerUsername,
						Password:      &brokerPassword,
						Labels:        &testCreateLabel,
					}),
					ExpectError: regexp.MustCompile(`API Error in creating Service Broker`),
				},
				{
					Config: hclProvider(nil) + hclServiceBroker(&ServiceBrokerModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rs",
						Name:          &brokerName2,
						Url:           &brokerURL,
						Username:      &brokerUsername,
						Password:      new("hello"),
						Labels:        &testCreateLabel,
					}),
					ExpectError: regexp.MustCompile(`Unable to verify service broker creation`),
				},
				{
					Config: hclProvider(nil) + hclServiceBroker(&ServiceBrokerModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rs",
						Name:          &brokerName2,
						Url:           &brokerURL,
						Username:      &brokerUsername,
						Password:      &brokerPassword,
						Labels:        &testCreateLabel,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestCheckResourceAttr(resourceName, "name", brokerName2),
						resource.TestCheckResourceAttr(resourceName, "url", brokerURL),
						resource.TestCheckResourceAttr(resourceName, "username", brokerUsername),
						resource.TestCheckResourceAttr(resourceName, "labels.purpose", "testing"),
					),
				},
				{
					Config: hclProvider(nil) + hclServiceBroker(&ServiceBrokerModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rs",
						Name:          &existingBrokerName,
						Url:           &brokerURL,
						Username:      &brokerUsername,
						Password:      &brokerPassword,
						Labels:        &testCreateLabel,
					}),
					ExpectError: regexp.MustCompile(`API Error Updating Service Broker`),
				},
				{
					Config: hclProvider(nil) + hclServiceBroker(&ServiceBrokerModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rs",
						Name:          &brokerName2,
						Url:           &brokerURL,
						Username:      &brokerUsername,
						Password:      new("hello"),
						Labels:        &testCreateLabel,
					}),
					ExpectError: regexp.MustCompile(`Unable to verify service broker update`),
				},
			},
		})
	})

}
