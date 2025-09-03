package provider

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	cfconfig "github.com/cloudfoundry/go-cfclient/v3/config"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider/managers"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &CloudFoundryProvider{}

type CloudFoundryProvider struct {
	version    string
	httpClient *http.Client
}

type CloudFoundryProviderModel struct {
	Endpoint          types.String `tfsdk:"api_url"`
	User              types.String `tfsdk:"user"`
	Password          types.String `tfsdk:"password"`
	CFClientID        types.String `tfsdk:"cf_client_id"`
	CFClientSecret    types.String `tfsdk:"cf_client_secret"`
	SkipSslValidation types.Bool   `tfsdk:"skip_ssl_validation"`
	Origin            types.String `tfsdk:"origin"`
	AccessToken       types.String `tfsdk:"access_token"`
	RefreshToken      types.String `tfsdk:"refresh_token"`
	AssertionToken    types.String `tfsdk:"assertion_token"`
}

func (p *CloudFoundryProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "cloudfoundry"
	resp.Version = p.version
}

func (p *CloudFoundryProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The Cloud Foundry Terraform plugin is an integration that allows users to leverage Terraform, an infrastructure as code tool, to define and provision infrastructure resources within the Cloud Foundry platform.",
		Attributes: map[string]schema.Attribute{
			"api_url": schema.StringAttribute{
				MarkdownDescription: "Specific URL representing the entry point for communication between the client and a Cloud Foundry instance.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"user": schema.StringAttribute{
				MarkdownDescription: "A unique identifier associated with an individual or entity for authentication & authorization purposes.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "A confidential alphanumeric code associated with a user account on the Cloud Foundry platform, requires user to authenticate.",
				Optional:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"cf_client_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Unique identifier for a client application used in authentication and authorization processes",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"cf_client_secret": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "A confidential string used by a client application for secure authentication and authorization, requires cf_client_id to authenticate",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"skip_ssl_validation": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Allows the client to disregard SSL certificate validation when connecting to the Cloud Foundry API",
			},
			"origin": schema.StringAttribute{
				MarkdownDescription: "Indicates the identity provider to be used for login",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"access_token": schema.StringAttribute{
				MarkdownDescription: "OAuth token to authenticate with Cloud Foundry",
				Optional:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"refresh_token": schema.StringAttribute{
				MarkdownDescription: "Token to refresh the access token, requires access_token",
				Optional:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"assertion_token": schema.StringAttribute{
				MarkdownDescription: "OAuth JWT assertion token. Used for OAuth 2.0 JWT Bearer Assertion Grant flow to authenticate with Cloud Foundry. Typically used with a custom origin.",
				Optional:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		},
	}
}
func addGenericAttributeError(resp *provider.ConfigureResponse, status string, pathRoot string, commonName string, envName string) {
	resp.Diagnostics.AddAttributeError(
		path.Root(pathRoot),
		fmt.Sprintf("%s field %s", status, pathRoot),
		fmt.Sprintf("The provider cannot create the Cloud Foundry API client as there is an unknown configuration value for the Cloud Foundry %s. "+
			"Either target apply the source of the value first, set the value statically in the configuration, or use the %s environment variable, ensure value is not empty. ", commonName, envName),
	)
}
func addTypeCastAttributeError(resp *provider.ConfigureResponse, expectedType string, pathRoot string, commonName string, envName string) {
	resp.Diagnostics.AddAttributeError(
		path.Root(pathRoot),
		fmt.Sprintf("Expected %s in field %s", expectedType, pathRoot),
		fmt.Sprintf("The provider cannot create the Cloud Foundry API client as there is an invalid configuration value for the Cloud Foundry %s. "+
			"Ensure %s is of type %s ", commonName, envName, expectedType),
	)
}
func checkConfigUnknown(config *CloudFoundryProviderModel, resp *provider.ConfigureResponse) {
	_, cfconfigerr := cfconfig.NewFromCFHome()

	anyParamExists := !config.User.IsUnknown() || !config.Password.IsUnknown() || !config.CFClientID.IsUnknown() || !config.CFClientSecret.IsUnknown() || !config.AccessToken.IsUnknown() || !config.RefreshToken.IsUnknown() || !config.AssertionToken.IsUnknown()

	/*
		There can be 3 cases of error:
		1. If endpoint is unknown and any other parameter is set
		2. If endpoint is set and all other parameter is unknown
		3. If all parameters are unknown and CF config is not correctly set
	*/
	if (config.Endpoint.IsUnknown() && anyParamExists) || (!config.Endpoint.IsUnknown() && !anyParamExists) || (!anyParamExists && cfconfigerr != nil) {
		resp.Diagnostics.AddError(
			"Unable to create CF Client due to missing values",
			"Either user/password or client_id/client_secret or access_token must be set with api_url or CF config must exist in path (default ~/.cf/config.json)",
		)
	}
	if !config.Endpoint.IsUnknown() {
		switch {
		case config.User.IsUnknown() && !config.Password.IsUnknown():
			addGenericAttributeError(resp, "Unknown", "user", "Username", "CF_USER")
		case !config.User.IsUnknown() && config.Password.IsUnknown():
			addGenericAttributeError(resp, "Unknown", "password", "Password", "CF_PASSWORD")
		case config.CFClientID.IsUnknown() && !config.CFClientSecret.IsUnknown():
			addGenericAttributeError(resp, "Unknown", "cf_client_id", "CF Client ID", "CF_CLIENT_ID")
		case !config.CFClientID.IsUnknown() && config.CFClientSecret.IsUnknown():
			addGenericAttributeError(resp, "Unknown", "cf_client_secret", "CF Client Secret", "CF_CLIENT_SECRET")
		}
	}
}

func checkConfig(resp *provider.ConfigureResponse, config managers.CloudFoundryProviderConfig) {
	_, cfconfigerr := cfconfig.NewFromCFHome()

	anyParamExists := config.User != "" || config.Password != "" || config.CFClientID != "" || config.CFClientSecret != "" || config.AccessToken != "" || config.RefreshToken != "" || config.AssertionToken != ""
	// There can be 3 cases of error:
	// 1. If endpoint is empty and any other parameter is set
	// 2. If endpoint is set and all other parameter is empty
	// 3. If all parameters are empty and CF config is not correctly set
	if (config.Endpoint == "" && anyParamExists) || (config.Endpoint != "" && !anyParamExists) || (!anyParamExists && cfconfigerr != nil) {
		resp.Diagnostics.AddError(
			"Unable to create CF Client due to missing values",
			"Either user/password or client_id/client_secret or access_token must be set with api_url or CF config must exist in path (default ~/.cf/config.json)",
		)
	}

	if config.Endpoint != "" {
		switch {
		case config.User == "" && config.Password != "":
			addGenericAttributeError(resp, "Missing", "user", "Username", "CF_USER")
		case config.User != "" && config.Password == "":
			addGenericAttributeError(resp, "Missing", "password", "Password", "CF_PASSWORD")
		case config.CFClientID == "" && config.CFClientSecret != "":
			addGenericAttributeError(resp, "Missing", "cf_client_id", "Client ID", "CF_CLIENT_ID")
		case config.CFClientID != "" && config.CFClientSecret == "":
			addGenericAttributeError(resp, "Missing", "cf_client_secret", " Client Secret", "CF_CLIENT_SECRET")
		}
	}
}

func getAndSetProviderValues(config *CloudFoundryProviderModel, resp *provider.ConfigureResponse) *managers.CloudFoundryProviderConfig {
	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	c := managers.CloudFoundryProviderConfig{
		Endpoint:       os.Getenv("CF_API_URL"),
		User:           os.Getenv("CF_USER"),
		Password:       os.Getenv("CF_PASSWORD"),
		CFClientID:     os.Getenv("CF_CLIENT_ID"),
		CFClientSecret: os.Getenv("CF_CLIENT_SECRET"),
		Origin:         os.Getenv("CF_ORIGIN"),
		AccessToken:    os.Getenv("CF_ACCESS_TOKEN"),
		RefreshToken:   os.Getenv("CF_REFRESH_TOKEN"),
		AssertionToken: os.Getenv("CF_ASSERTION_TOKEN"),
	}

	var err error
	if os.Getenv("CF_SKIP_SSL_VALIDATION") != "" {
		c.SkipSslValidation, err = strconv.ParseBool(os.Getenv("CF_SKIP_SSL_VALIDATION"))
		if err != nil {
			addTypeCastAttributeError(resp, "Boolean", "skip_ssl_validation", "Skip SSL Validation", "CF_SKIP_SSL_VALIDATION")
			return nil
		}
	}
	if !config.Endpoint.IsNull() {
		c.Endpoint = config.Endpoint.ValueString()
	}
	if !config.User.IsNull() {
		c.User = config.User.ValueString()
	}
	if !config.Password.IsNull() {
		c.Password = config.Password.ValueString()
	}
	if !config.CFClientID.IsNull() {
		c.CFClientID = config.CFClientID.ValueString()
	}
	if !config.CFClientSecret.IsNull() {
		c.CFClientSecret = config.CFClientSecret.ValueString()
	}
	if !config.Origin.IsNull() {
		c.Origin = config.Origin.ValueString()
	}
	if !config.AccessToken.IsNull() {
		c.AccessToken = config.AccessToken.ValueString()
	}
	if !config.RefreshToken.IsNull() {
		c.RefreshToken = config.RefreshToken.ValueString()
	}
	if !config.AssertionToken.IsNull() {
		c.AssertionToken = config.AssertionToken.ValueString()
	}

	checkConfig(resp, c)
	if resp.Diagnostics.HasError() {
		return nil
	}
	if !config.SkipSslValidation.IsNull() {
		c.SkipSslValidation = config.SkipSslValidation.ValueBool()
	}
	c.Endpoint = strings.TrimSuffix(c.Endpoint, "/")

	return &c
}
func (p *CloudFoundryProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config CloudFoundryProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	checkConfigUnknown(&config, resp)

	if resp.Diagnostics.HasError() {
		return
	}

	cloudFoundryProviderConfig := getAndSetProviderValues(&config, resp)
	if resp.Diagnostics.HasError() {
		return
	}
	session, err := cloudFoundryProviderConfig.NewSession(p.httpClient, req)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create CF Client",
			"Client creation failed with error "+err.Error(),
		)
	}

	// Make the Cloud Foundry session available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = session
	resp.ResourceData = session
}

func (p *CloudFoundryProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewOrgResource,
		NewOrgQuotaResource,
		NewSpaceResource,
		NewUserResource,
		NewSpaceQuotaResource,
		NewSpaceRoleResource,
		NewOrgeRoleResource,
		NewServiceInstanceResource,
		NewServiceInstanceSharingResource,
		NewSecurityGroupResource,
		NewRouteResource,
		NewDomainResource,
		NewAppResource,
		NewServiceCredentialBindingResource,
		NewMtaResource,
		NewIsolationSegmentResource,
		NewIsolationSegmentEntitlementResource,
		NewServiceRouteBindingResource,
		NewBuildpackResource,
		NewServiceBrokerResource,
		NewUserGroupsResource,
		NewSecurityGroupSpacesResource,
		NewCFUserResource,
		NewServicePlanVisibilityResource,
		NewNetworkPolicyResource,
	}
}

func (p *CloudFoundryProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewOrgDataSource,
		NewOrgQuotaDataSource,
		NewSpaceDataSource,
		NewUserDataSource,
		NewSpaceQuotaDataSource,
		NewSpaceRoleDataSource,
		NewOrgRoleDataSource,
		NewUsersDataSource,
		NewServiceInstanceDataSource,
		NewSecurityGroupDataSource,
		NewRouteDataSource,
		NewDomainDataSource,
		NewAppDataSource,
		NewServiceCredentialBindingsDataSource,
		NewServiceCredentialBindingDataSource,
		NewServiceCredentialBindingDetailsDataSource,
		NewMtasDataSource,
		NewMtaDataSource,
		NewIsolationSegmentDataSource,
		NewIsolationSegmentEntitlementDataSource,
		NewStackDataSource,
		NewRemoteMtarHashDataSource,
		NewSpacesDataSource,
		NewServicePlansDataSource,
		NewServicePlanDataSource,
		NewOrgsDataSource,
		NewServiceInstancesDataSource,
		NewOrgRolesDataSource,
		NewSpaceQuotasDataSource,
		NewAppsDataSource,
		NewSpaceRolesDataSource,
		NewDomainsDataSource,
		NewRoutesDataSource,
		NewServiceBrokerDataSource,
		NewServiceRouteBindingsDataSource,
		NewServiceBrokersDataSource,
		NewServiceRouteBindingDataSource,
		NewBuildpacksDataSource,
		NewIsolationSegmentsDataSource,
		NewOrgQuotasDataSource,
		NewSecurityGroupsDataSource,
		NewStacksDataSource,
	}
}

func New(version string, httpClient *http.Client) func() provider.Provider {
	return func() provider.Provider {
		return &CloudFoundryProvider{
			version:    version,
			httpClient: httpClient,
		}
	}
}
