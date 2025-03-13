package provider

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"regexp"
	"testing"
	"text/template"

	cfconfig "github.com/cloudfoundry/go-cfclient/v3/config"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	testingResource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/assert"

	"gopkg.in/dnaeon/go-vcr.v3/recorder"
)

type CloudFoundryProviderConfigPtr struct {
	Endpoint          *string
	User              *string
	Password          *string
	CFClientID        *string
	CFClientSecret    *string
	SkipSslValidation *bool
	Origin            *string
	AccessToken       *string
	RefreshToken      *string
}

var redactedTestUser = CloudFoundryProviderConfigPtr{
	Endpoint:       strtostrptr("https://api.x.x.x.x.com"),
	User:           strtostrptr("xx"),
	Password:       strtostrptr("xxxx"),
	CFClientID:     strtostrptr("xx"),
	CFClientSecret: strtostrptr("xxxx"),
	AccessToken:    strtostrptr("bearer eyJhbGciOiJSUzI1NiIsImprdSI6Imh0dHBzOi8vdWFhLngueC54LnguY29tL3Rva2VuX2tleXMiLCJraWQiOiJrZXktMSIsInR5cCI6IkpXVCJ9.eyJqdGkiOiI0YmYyMmRhYjNiYmU0NTg1OTUwM2Y0MWExZmRkZmFmOCIsImNsaWVudF9hdXRoX21ldGhvZCI6Im5vbmUiLCJzdWIiOiIxZjI2M2UwMC05YTA3LTQyZDgtYmU3MS1iMThiZTZkMTBiNDUiLCJzY29wZSI6WyJvcGVuaWQiLCJ1YWEudXNlciIsImNsb3VkX2NvbnRyb2xsZXIucmVhZCIsInBhc3N3b3JkLndyaXRlIiwiY2xvdWRfY29udHJvbGxlci53cml0ZSJdLCJjbGllbnRfaWQiOiJjZiIsImNpZCI6ImNmIiwiYXpwIjoiY2YiLCJncmFudF90eXBlIjoicGFzc3dvcmQiLCJ1c2VyX2lkIjoiMWYyNjNlMDAtOWEwNy00MmQ4LWJlNzEtYjE4YmU2ZDEwYjQ1Iiwib3JpZ2luIjoic2FwLmlkcyIsInVzZXJfbmFtZSI6InRlc3RfdXNlckB4LmNvbSIsImVtYWlsIjoidGVzdF91c2VyQHguY29tIiwiYXV0aF90aW1lIjoxNzA1MDYwOTM2LCJyZXZfc2lnIjoiZGUxMDU3ZDEiLCJpYXQiOjE3MDUwNjA5MzYsImV4cCI6MjcwNTA2MjEzNiwiaXNzIjoiaHR0cHM6Ly91YWEueC54LngueC5jb20vb2F1dGgvdG9rZW4iLCJ6aWQiOiJ1YWEiLCJhdWQiOlsiY2xvdWRfY29udHJvbGxlciIsInBhc3N3b3JkIiwiY2YiLCJ1YWEiLCJvcGVuaWQiXX0.DADNqcmHbP8R0Dp3pMZVE7OeD5eTmcwh5dyFKFpryGEl3QqXKd1Af3raTFnJe1SRi66qjkvpdLub31Fh3LDdkAPAoFYshvwxozCdEinGYEx-qlW1Ttt6qyk_0y3CjKDExv43F8CpHwqD41A57IOAbz14revnb6tbW9pA_dBxhF9sYdXJvhPOnGUDKgv5SIYNUyt0_ekEaHNMVHp__4dnaCw7qdMkJ7Y7Pn4ES3KJqc88Ed9PzRJw0WQzwvHlJbQyCtpBXFx_ZzIEFNjcXo9p-YbezEKVypKlREs59h-HzpbhLwjW9_MzuY3wFveYT4FLsF-U0s0KeQq83E8J_zWRhw"),
	RefreshToken:   strtostrptr("xxxx"),
}

func hclProvider(cfConfig *CloudFoundryProviderConfigPtr) string {
	if cfConfig != nil {
		s := `
			provider "cloudfoundry" {
			{{- if .Endpoint}}
				api_url  = "{{.Endpoint}}"
			{{- end -}}
			{{if .User}}
				user = "{{.User}}"
			{{- end -}}
			{{if .Password}}
				password = "{{.Password}}"
			{{- end -}}
			{{if .CFClientID}}
				cf_client_id = "{{.CFClientID}}"
			{{- end -}}
			{{if .CFClientSecret}}
				cf_client_secret = "{{.CFClientSecret}}"
			{{- end -}}
			{{if .SkipSslValidation}}
				skip_ssl_validation = "{{.SkipSslValidation}}"
			{{- end -}}
			{{if .Origin}}
				origin = "{{.Origin}}"
			{{- end -}}
			{{if .AccessToken}}
				access_token = "{{.AccessToken}}"
			{{- end -}}
			{{if .RefreshToken}}
				refresh_token = "{{.RefreshToken}}"
			{{- end }}
			}`
		tmpl, err := template.New("provider").Parse(s)
		if err != nil {
			panic(err)
		}
		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, cfConfig)
		if err != nil {
			panic(err)
		}
		return buf.String()
	}
	return `provider "cloudfoundry" {}`
}
func hclProviderWithDataSource(cfConfig *CloudFoundryProviderConfigPtr) string {
	s := `
	data "cloudfoundry_org" "org" {
		name = "PerformanceTeamBLR"
	}`
	return hclProvider(cfConfig) + s
}

func TestCloudFoundryProvider_Configure(t *testing.T) {
	t.Parallel()
	t.Run("error path - user login with missing user/password data", func(t *testing.T) {

		testingResource.Test(t, testingResource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(http.DefaultClient),
			Steps: []testingResource.TestStep{
				{
					Config: hclProviderWithDataSource(&CloudFoundryProviderConfigPtr{
						Endpoint: redactedTestUser.Endpoint,
						Password: redactedTestUser.Password,
					}),
					ExpectError: regexp.MustCompile(`Error: Missing field user`),
				},
				{
					Config: hclProviderWithDataSource(&CloudFoundryProviderConfigPtr{
						Endpoint: redactedTestUser.Endpoint,
						User:     redactedTestUser.User,
					}),
					ExpectError: regexp.MustCompile(`Error: Missing field password`),
				},
			},
		})
	})
	t.Run("error path - user login with missing clientid/clientsecret data", func(t *testing.T) {

		testingResource.Test(t, testingResource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(http.DefaultClient),
			Steps: []testingResource.TestStep{
				{
					Config: hclProviderWithDataSource(&CloudFoundryProviderConfigPtr{
						Endpoint:   redactedTestUser.Endpoint,
						CFClientID: redactedTestUser.CFClientID,
					}),
					ExpectError: regexp.MustCompile(`Error: Missing field cf_client_secret`),
				},
				{
					Config: hclProviderWithDataSource(&CloudFoundryProviderConfigPtr{
						Endpoint:       redactedTestUser.Endpoint,
						CFClientSecret: redactedTestUser.CFClientSecret,
					}),
					ExpectError: regexp.MustCompile(`Error: Missing field cf_client_id`),
				},
				{
					Config: hclProviderWithDataSource(&CloudFoundryProviderConfigPtr{
						Endpoint: redactedTestUser.Endpoint,
					}),
					ExpectError: regexp.MustCompile(`Error: Unable to create CF Client due to missing values`),
				},
			},
		})
	})
	t.Run("user login with valid user/pass data", func(t *testing.T) {
		endpoint := strtostrptr(os.Getenv("TEST_CF_API_URL"))
		user := strtostrptr(os.Getenv("TEST_CF_USER"))
		password := strtostrptr(os.Getenv("TEST_CF_PASSWORD"))
		if *endpoint == "" || *user == "" || *password == "" {
			t.Logf("\nATTENTION: Using redacted user credentions since endpoint, username & password not set as env \n Make sure you are not triggering a recording else test will fail")
			endpoint = redactedTestUser.Endpoint
			user = redactedTestUser.User
			password = redactedTestUser.Password
		}
		cfg := CloudFoundryProviderConfigPtr{
			Endpoint: endpoint,
			User:     user,
			Password: password,
		}
		recUserPass := cfg.SetupVCR(t, "fixtures/provider.user_pwd")
		defer stopQuietly(recUserPass)

		testingResource.Test(t, testingResource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(recUserPass.GetDefaultClient()),
			Steps: []testingResource.TestStep{
				{
					Config: hclProvider(&cfg) + `
					data "cloudfoundry_org" "org" {
						name = "PerformanceTeamBLR"
					}`,
				},
			},
		})
	})
	t.Run("user login with valid access token", func(t *testing.T) {
		endpoint := strtostrptr(os.Getenv("TEST_CF_API_URL"))
		accessToken := strtostrptr(os.Getenv("TEST_CF_ACCESS_TOKEN"))
		refreshToken := strtostrptr(os.Getenv("TEST_CF_REFRESH_TOKEN"))
		if *endpoint == "" || *accessToken == "" || *refreshToken == "" {
			t.Logf("\nATTENTION: Using redacted user credentions since endpoint, username & password not set as env \n Make sure you are not triggering a recording else test will fail")
			endpoint = redactedTestUser.Endpoint
			accessToken = redactedTestUser.AccessToken
			refreshToken = redactedTestUser.RefreshToken
		}
		cfg := CloudFoundryProviderConfigPtr{
			Endpoint:     endpoint,
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		}

		recUserPass := cfg.SetupVCR(t, "fixtures/provider.access_token")
		defer stopQuietly(recUserPass)

		testingResource.Test(t, testingResource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(recUserPass.GetDefaultClient()),
			Steps: []testingResource.TestStep{
				{
					Config: hclProvider(&cfg) + `
				data "cloudfoundry_org" "org" {
					name = "PerformanceTeamBLR"
				}`,
				},
			},
		})
	})
	t.Run("user login with valid home directory", func(t *testing.T) {
		cfg := getCFHomeConf()
		recHomeDir := cfg.SetupVCR(t, "fixtures/provider.home_dir")
		defer stopQuietly(recHomeDir)

		testingResource.Test(t, testingResource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(recHomeDir.GetDefaultClient()),
			Steps: []testingResource.TestStep{
				{
					Config: hclProviderWithDataSource(nil),
				},
			},
		})
	})
}

func getProviders(httpClient *http.Client) map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"cloudfoundry": providerserver.NewProtocol6WithError(New("test", httpClient)()),
	}
}
func getCFHomeConf() *CloudFoundryProviderConfigPtr {
	cfConf, err := cfconfig.NewFromCFHome()
	if err != nil {
		return &CloudFoundryProviderConfigPtr{
			Endpoint: strtostrptr("https://api.x.x.x.x.com"),
		}
	}
	apiEndpointURL := cfConf.ApiURL("")
	cfg := CloudFoundryProviderConfigPtr{
		Endpoint: &apiEndpointURL,
	}
	return &cfg
}
func stopQuietly(rec *recorder.Recorder) {
	if err := rec.Stop(); err != nil {
		panic(err)
	}
}
func TestCloudFoundryProvider_HasResources(t *testing.T) {
	expectedResources := []string{
		"cloudfoundry_org",
		"cloudfoundry_org_quota",
		"cloudfoundry_space",
		"cloudfoundry_user",
		"cloudfoundry_space_quota",
		"cloudfoundry_space_role",
		"cloudfoundry_org_role",
		"cloudfoundry_security_group",
		"cloudfoundry_service_instance",
		"cloudfoundry_service_instance_sharing",
		"cloudfoundry_route",
		"cloudfoundry_domain",
		"cloudfoundry_app",
		"cloudfoundry_service_credential_binding",
		"cloudfoundry_mta",
		"cloudfoundry_isolation_segment",
		"cloudfoundry_isolation_segment_entitlement",
		"cloudfoundry_service_route_binding",
		"cloudfoundry_buildpack",
		"cloudfoundry_service_broker",
		"cloudfoundry_user_groups",
		"cloudfoundry_security_group_space_bindings",
	}

	ctx := context.Background()
	registeredResources := []string{}

	for _, resourceFunc := range New("test", &http.Client{})().Resources(ctx) {
		var resp resource.MetadataResponse

		resourceFunc().Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "cloudfoundry"}, &resp)

		registeredResources = append(registeredResources, resp.TypeName)
	}

	assert.ElementsMatch(t, expectedResources, registeredResources)
}

func TestProvider_HasDataSources(t *testing.T) {
	expectedDataSources := []string{
		"cloudfoundry_org",
		"cloudfoundry_space",
		"cloudfoundry_org_quota",
		"cloudfoundry_user",
		"cloudfoundry_space_quota",
		"cloudfoundry_space_role",
		"cloudfoundry_org_role",
		"cloudfoundry_users",
		"cloudfoundry_security_group",
		"cloudfoundry_service_instance",
		"cloudfoundry_route",
		"cloudfoundry_domain",
		"cloudfoundry_app",
		"cloudfoundry_service_credential_binding",
		"cloudfoundry_mtas",
		"cloudfoundry_mta",
		"cloudfoundry_isolation_segment",
		"cloudfoundry_isolation_segment_entitlement",
		"cloudfoundry_stack",
		"cloudfoundry_remote_mtar_hash",
		"cloudfoundry_spaces",
		"cloudfoundry_service_plans",
		"cloudfoundry_orgs",
		"cloudfoundry_service_instances",
		"cloudfoundry_org_roles",
		"cloudfoundry_space_quotas",
		"cloudfoundry_apps",
		"cloudfoundry_space_roles",
		"cloudfoundry_domains",
		"cloudfoundry_routes",
		"cloudfoundry_service_broker",
		"cloudfoundry_service_route_bindings",
		"cloudfoundry_service_brokers",
		"cloudfoundry_service_route_binding",
		"cloudfoundry_buildpacks",
		"cloudfoundry_isolation_segments",
		"cloudfoundry_org_quotas",
		"cloudfoundry_security_groups",
		"cloudfoundry_stacks",
	}

	ctx := context.Background()
	registeredDataSources := []string{}

	for _, resourceFunc := range New("test", &http.Client{})().DataSources(ctx) {
		var resp datasource.MetadataResponse

		resourceFunc().Metadata(ctx, datasource.MetadataRequest{ProviderTypeName: "cloudfoundry"}, &resp)

		registeredDataSources = append(registeredDataSources, resp.TypeName)
	}

	assert.ElementsMatch(t, expectedDataSources, registeredDataSources)
}
