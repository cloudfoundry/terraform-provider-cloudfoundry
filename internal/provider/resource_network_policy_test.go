package provider

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type NetworkPolicyPtr struct {
	HclObjectName    string
	SourceAppId      string
	DestinationAppId string
	PortStr          string
}

func hclNetworkPolicy(npp *NetworkPolicyPtr) string {
	if npp != nil {
		s := `
		resource "cloudfoundry_network_policy" "{{.HclObjectName}}" {
			policies = [
				{
					source_app = "{{.SourceAppId}}"
					destination_app = "{{.DestinationAppId}}"
					port = "{{.PortStr}}"
				}
			]
		}`
		tmpl, err := template.New("resource_network_policy").Parse(s)
		if err != nil {
			panic(err)
		}
		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, npp)
		if err != nil {
			panic(err)
		}
		return buf.String()
	}
	return `resource "cloudfoundry_network_policy" "np" {}`
}

func TestNetworkPolicyResource_Configure(t *testing.T) {
	t.Parallel()

	t.Run("happy path - create/read/update/delete policy", func(t *testing.T) {
		resourceName := "cloudfoundry_network_policy.np"
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_network_policy_crud")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclNetworkPolicy(&NetworkPolicyPtr{
						HclObjectName:    "np",
						SourceAppId:      "a4bf5d3c-b9ac-4d6b-bc36-edb82e9cbda1",
						DestinationAppId: "8888f08b-f5c9-4e89-8f6b-95e0c2e5c7f0",
						PortStr:          "61443",
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestCheckResourceAttr(resourceName, "policies.0.protocol", "tcp"),
						resource.TestCheckResourceAttr(resourceName, "policies.0.port", "61443"),
					),
				},
			},
		})
	})

	t.Run("error path - invalid port range", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_network_policy_invalid_source")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclNetworkPolicy(&NetworkPolicyPtr{
						HclObjectName:    "np_invalid_source",
						SourceAppId:      "a4bf5d3c-b9ac-4d6b-bc36-edb82e9cbda1",
						DestinationAppId: "8888f08b-f5c9-4e89-8f6b-95e0c2e5c7f0",
						PortStr:          "8090-8089",
					}),
					ExpectError: regexp.MustCompile(`API Error Creating Policies`),
				},
			},
		})
	})
}
