package provider

import (
	"context"
	"fmt"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	cfv3operation "github.com/cloudfoundry/go-cfclient/v3/operation"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider/managers"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"gopkg.in/yaml.v2"
)

var _ datasource.DataSource = &appDataSource{}
var _ datasource.DataSourceWithConfigure = &appDataSource{}

func NewAppDataSource() datasource.DataSource {
	return &appDataSource{}
}

type appDataSource struct {
	cfClient *cfv3client.Client
}

func (d *appDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app"
}

func (d *appDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Gets information on a Cloud Foundry application.",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the application to look up",
				Required:            true,
			},
			"space_name": schema.StringAttribute{
				MarkdownDescription: "The name of the space to look up",
				Required:            true,
			},
			"org_name": schema.StringAttribute{
				MarkdownDescription: "The name of the associated Cloud Foundry organization to look up",
				Required:            true,
			},
			"enable_ssh": schema.BoolAttribute{
				MarkdownDescription: "Whether SSH access is enabled or disabled on an app level.",
				Computed:            true,
			},
			"stack": schema.StringAttribute{
				MarkdownDescription: "The name of the stack the application will be deployed to.",
				Computed:            true,
			},
			"buildpacks": schema.ListAttribute{
				MarkdownDescription: "Multiple buildpacks used to stage the application.",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"docker_image": schema.StringAttribute{
				MarkdownDescription: "The URL to the docker image with tag",
				Computed:            true,
			},
			"docker_credentials": schema.SingleNestedAttribute{
				MarkdownDescription: "Defines login credentials for private docker repositories",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"username": schema.StringAttribute{
						MarkdownDescription: "The username for the private docker repository.",
						Computed:            true,
						Sensitive:           true,
					},
				},
			},
			"app_lifecycle": schema.StringAttribute{
				MarkdownDescription: "The lifecycle type for the app which can be one of 'buildpack', 'docker', 'cnb'.",
				Computed:            true,
			},
			"service_bindings": schema.SetNestedAttribute{
				MarkdownDescription: "Service instances bound to the application.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"service_instance": schema.StringAttribute{
							MarkdownDescription: "The service instance name.",
							Computed:            true,
						},
						"params": schema.StringAttribute{
							CustomType:          jsontypes.NormalizedType{},
							MarkdownDescription: "A json object to represent the parameters for the service instance.",
							Computed:            true,
						},
					},
				},
			},
			"routes": schema.SetNestedAttribute{
				MarkdownDescription: "The routes to map to the application to control its ingress traffic.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"route": schema.StringAttribute{
							MarkdownDescription: "The fully qualified domain name which will be bound to app",
							Computed:            true,
						},
						"protocol": schema.StringAttribute{
							MarkdownDescription: "The protocol used for the route. Valid values are http2, http1, and tcp.",
							Computed:            true,
						},
					},
				},
			},
			"environment": schema.MapAttribute{
				MarkdownDescription: "Key/value pairs of custom environment variables in your app. Does not include any system or service variables.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"processes": schema.SetNestedAttribute{
				MarkdownDescription: "List of configurations for individual process types.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: datasourceProcessSchemaAttributes(),
				},
			},
			"sidecars": schema.SetNestedAttribute{
				MarkdownDescription: "The attribute specifies additional processes to run in the same container as your app",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Sidecar name. The identifier for the sidecars to be configured.",
							Computed:            true,
						},
						"command": schema.StringAttribute{
							MarkdownDescription: "The command used to start the sidecar.",
							Computed:            true,
						},
						"process_types": schema.SetAttribute{
							MarkdownDescription: "List of processes to associate sidecar with. Any string identifier (e.g., web, worker, scheduler).",
							ElementType:         types.StringType,
							Computed:            true,
						},
						"memory": schema.StringAttribute{
							MarkdownDescription: "The memory limit for the sidecar.",
							Computed:            true,
						},
					},
				},
			},
			labelsKey:      datasourceLabelsSchema(),
			annotationsKey: datasourceAnnotationsSchema(),
			idKey:          guidSchema(),
			createdAtKey:   createdAtSchema(),
			updatedAtKey:   updatedAtSchema(),
		},
	}
	for k, v := range datasourceProcessAppCommonSchema() {
		if _, ok := resp.Schema.Attributes[k]; !ok {
			resp.Schema.Attributes[k] = v
		}
	}
}
func datasourceProcessSchemaAttributes() map[string]schema.Attribute {
	pSchema := map[string]schema.Attribute{
		"type": schema.StringAttribute{
			MarkdownDescription: "The process type. Any string identifier (e.g., web, worker, scheduler).",
			Computed:            true,
		},
	}
	for k, v := range datasourceProcessAppCommonSchema() {
		if _, ok := pSchema[k]; !ok {
			pSchema[k] = v
		}
	}
	return pSchema
}
func datasourceProcessAppCommonSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"command": schema.StringAttribute{
			MarkdownDescription: "A custom start command for the application. This overrides the start command provided by the buildpack.",
			Computed:            true,
		},
		"disk_quota": schema.StringAttribute{
			MarkdownDescription: "The disk space to be allocated for each application instance.",
			Computed:            true,
		},
		"health_check_http_endpoint": schema.StringAttribute{
			MarkdownDescription: "The endpoint for the http health check type.",
			Computed:            true,
		},
		"health_check_invocation_timeout": schema.Int64Attribute{
			MarkdownDescription: "The timeout in seconds for the health check requests for http and port health checks.",
			Computed:            true,
		},
		"health_check_type": schema.StringAttribute{
			MarkdownDescription: "The health check type which can be one of 'port', 'process', 'http'.",
			Computed:            true,
		},
		"health_check_interval": schema.Int64Attribute{
			MarkdownDescription: "The interval in seconds between health checks.",
			Computed:            true,
		},
		"readiness_health_check_type": schema.StringAttribute{
			MarkdownDescription: "The readiness health check type which can be one of 'port', 'process', 'http'.",
			Computed:            true,
		},
		"readiness_health_check_http_endpoint": schema.StringAttribute{
			MarkdownDescription: "The endpoint for the http readiness health check type.",
			Computed:            true,
		},
		"readiness_health_check_invocation_timeout": schema.Int64Attribute{
			MarkdownDescription: "The timeout in seconds for the readiness health check requests for http and port health checks.",
			Computed:            true,
		},
		"readiness_health_check_interval": schema.Int64Attribute{
			MarkdownDescription: "The interval in seconds between readiness health checks.",
			Computed:            true,
		},
		"log_rate_limit_per_second": schema.StringAttribute{
			MarkdownDescription: "The attribute specifies the log rate limit for all instances of an app.",
			Computed:            true,
		},
		"instances": schema.Int64Attribute{
			MarkdownDescription: "The number of app instances started.",
			Computed:            true,
		},
		"memory": schema.StringAttribute{
			MarkdownDescription: "The memory limit for each application instance.",
			Computed:            true,
		},
		"timeout": schema.Int64Attribute{
			MarkdownDescription: "Time in seconds at which the health-check will report failure.",
			Computed:            true,
		},
	}
}
func (d *appDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	session, ok := req.ProviderData.(*managers.Session)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *managers.Session, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.cfClient = session.CFClient
}

func (d *appDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var datasourceAppType DatasourceAppType
	diags := req.Config.Get(ctx, &datasourceAppType)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	org, err := d.cfClient.Organizations.Single(ctx, &cfv3client.OrganizationListOptions{
		Names: cfv3client.Filter{
			Values: []string{datasourceAppType.Org.ValueString()},
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Error finding given org", err.Error())
		return
	}
	space, err := d.cfClient.Spaces.Single(ctx, &cfv3client.SpaceListOptions{
		Names: cfv3client.Filter{
			Values: []string{datasourceAppType.Space.ValueString()},
		},
		OrganizationGUIDs: cfv3client.Filter{
			Values: []string{org.GUID},
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Error finding given space", err.Error())
		return
	}
	app, err := d.cfClient.Applications.First(ctx, &cfv3client.AppListOptions{
		Names: cfv3client.Filter{
			Values: []string{datasourceAppType.Name.ValueString()},
		},
		OrganizationGUIDs: cfv3client.Filter{
			Values: []string{org.GUID},
		},
		SpaceGUIDs: cfv3client.Filter{
			Values: []string{space.GUID},
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Error finding given app", err.Error())
		return
	}
	appRaw, err := d.cfClient.Manifests.Generate(ctx, app.GUID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading app", err.Error())
		return
	}
	sshResp, err := d.cfClient.AppFeatures.GetSSH(ctx, app.GUID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading app feature", err.Error())
		return
	}
	var appManifest cfv3operation.Manifest
	err = yaml.Unmarshal([]byte(appRaw), &appManifest)
	if err != nil {
		resp.Diagnostics.AddError("Error unmarshalling app", err.Error())
		return
	}
	atResp, diags := mapAppValuesToType(ctx, appManifest.Applications[0], app, nil, sshResp)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	datasourceAppTypeResp := atResp.Reduce()
	datasourceAppTypeResp.Org = datasourceAppType.Org
	datasourceAppTypeResp.Space = datasourceAppType.Space

	tflog.Trace(ctx, "read an app data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &datasourceAppTypeResp)...)
}
