package provider

import (
	"context"
	"fmt"

	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	cfv3operation "github.com/cloudfoundry/go-cfclient/v3/operation"
	"github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/provider/managers"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/validation"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"gopkg.in/yaml.v2"
)

var _ datasource.DataSource = &appsDataSource{}
var _ datasource.DataSourceWithConfigure = &appsDataSource{}

func NewAppsDataSource() datasource.DataSource {
	return &appsDataSource{}
}

type appsDataSource struct {
	cfClient *cfv3client.Client
}

func (d *appsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_apps"
}

func (d *appsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches information on Cloud Foundry applications present in a space.",

		Attributes: map[string]schema.Attribute{
			"org": schema.StringAttribute{
				MarkdownDescription: "The GUID of the org where the applications are present",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(path.Expressions{
						path.MatchRoot("space"),
						path.MatchRoot("org"),
					}...),
					validation.ValidUUID(),
				},
			},
			"space": schema.StringAttribute{
				MarkdownDescription: "The GUID of the space where the applications are present",
				Optional:            true,
				Validators: []validator.String{
					validation.ValidUUID(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the application to filter by",
				Optional:            true,
			},
			"apps": schema.ListNestedAttribute{
				MarkdownDescription: "The list of apps",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: datasourceAppsSchema(),
				},
			},
		},
	}
}

func datasourceAppsSchema() map[string]schema.Attribute {
	pSchema := map[string]schema.Attribute{
		"name": schema.StringAttribute{
			MarkdownDescription: "The name of the application to look up",
			Computed:            true,
		},
		"space_name": schema.StringAttribute{
			MarkdownDescription: "The name of the space to look up",
			Computed:            true,
		},
		"org_name": schema.StringAttribute{
			MarkdownDescription: "The name of the associated Cloud Foundry organization to look up",
			Computed:            true,
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
						MarkdownDescription: "List of processes to associate sidecar with.",
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
	}
	for k, v := range datasourceProcessAppCommonSchema() {
		if _, ok := pSchema[k]; !ok {
			pSchema[k] = v
		}
	}
	return pSchema
}

func (d *appsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *appsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var (
		data DatasourceAppsType
		org  *resource.Organization
		err  error
	)
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	appListOptions := cfv3client.NewAppListOptions()

	if !data.Org.IsNull() {
		org, err = d.cfClient.Organizations.Get(ctx, data.Org.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"API Error Fetching Organization",
				"Could not get details of the Organization with ID "+data.Org.ValueString()+" : "+err.Error(),
			)
			return
		}
		appListOptions.OrganizationGUIDs = cfv3client.Filter{
			Values: []string{
				org.GUID,
			},
		}
	}
	if !data.Space.IsNull() {
		_, org, err = d.cfClient.Spaces.GetIncludeOrganization(ctx, data.Space.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"API Error Fetching Space",
				"Could not get space with ID "+data.Space.ValueString()+" : "+err.Error(),
			)
			return
		}
		appListOptions.SpaceGUIDs = cfv3client.Filter{
			Values: []string{
				data.Space.ValueString(),
			},
		}
	}

	if !data.Name.IsNull() {
		appListOptions.Names = cfv3client.Filter{
			Values: []string{
				data.Name.ValueString(),
			},
		}
	}

	apps, err := d.cfClient.Applications.ListAll(ctx, appListOptions)
	if err != nil {
		resp.Diagnostics.AddError("API Error Fetching Apps", err.Error())
		return
	}

	appsList := []DatasourceAppType{}
	for _, app := range apps {

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
		datasourceAppTypeResp.Org = types.StringValue(org.Name)
		datasourceAppTypeResp.Space = types.StringValue(app.Relationships.Space.Data.GUID)
		appsList = append(appsList, datasourceAppTypeResp)
	}

	data.Apps = appsList
	tflog.Trace(ctx, "read an apps data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
