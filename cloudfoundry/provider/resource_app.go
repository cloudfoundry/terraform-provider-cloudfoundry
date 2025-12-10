package provider

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	logclient "code.cloudfoundry.org/go-log-cache/v3"
	cfv3client "github.com/cloudfoundry/go-cfclient/v3/client"
	cfv3operation "github.com/cloudfoundry/go-cfclient/v3/operation"
	cfv3resource "github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider/managers"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gopkg.in/yaml.v2"
)

var (
	_ resource.Resource                = &appResource{}
	_ resource.ResourceWithConfigure   = &appResource{}
	_ resource.ResourceWithImportState = &appResource{}
)

func NewAppResource() resource.Resource {
	return &appResource{}
}

type appResource struct {
	cfClient *cfv3client.Client
}

func (r *appResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app"
}

func (r *appResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provides a Cloud Foundry resource to manage applications.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the application.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"space_name": schema.StringAttribute{
				MarkdownDescription: "The name of the associated Cloud Foundry space.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"org_name": schema.StringAttribute{
				MarkdownDescription: "The name of the associated Cloud Foundry organization.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enable_ssh": schema.BoolAttribute{
				MarkdownDescription: "Whether to enable or disable SSH access on an app level.",
				Optional:            true,
				Computed:            true,
			},
			"stack": schema.StringAttribute{
				MarkdownDescription: "The base operating system and file system that your application will execute in. Please refer to the [docs](https://v3-apidocs.cloudfoundry.org/version/3.155.0/index.html#stacks) for more information",
				Optional:            true,
				Computed:            true,
			},
			"buildpacks": schema.ListAttribute{
				MarkdownDescription: "Multiple buildpacks used to stage the application.",
				ElementType:         types.StringType,
				Computed:            true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				Optional: true,
			},
			"path": schema.StringAttribute{
				MarkdownDescription: "The path to the zip file for the application.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("docker_image"), path.MatchRoot("path")),
				},
			},
			"source_code_hash": schema.StringAttribute{
				MarkdownDescription: "Used to trigger updates. Must be set to a base64-encoded SHA256 hash of the path specified.",
				Optional:            true,
			},
			"docker_image": schema.StringAttribute{
				MarkdownDescription: "The URL to the docker image with tag e.g registry.example.com:5000/user/repository/tag or docker image name from the public repo e.g. redis:4.0",
				Optional:            true,
			},
			"docker_credentials": schema.SingleNestedAttribute{
				MarkdownDescription: "Defines login credentials for private docker repositories",
				Optional:            true,
				Validators: []validator.Object{
					objectvalidator.AlsoRequires(path.MatchRoot("docker_image")),
				},
				Attributes: map[string]schema.Attribute{
					"username": schema.StringAttribute{
						MarkdownDescription: "The username for the private docker repository.",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.LengthAtLeast(1),
						},
						Sensitive: true,
					},
					"password": schema.StringAttribute{
						MarkdownDescription: "The password for the private docker repository.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.LengthAtLeast(1),
						},
						Sensitive: true,
					},
				},
			},
			"strategy": schema.StringAttribute{
				MarkdownDescription: "The deployment strategy to use when deploying the application. Valid values are 'none', 'rolling', and 'blue-green', defaults to 'none'.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("none", "rolling", "blue-green"),
				},
			},
			"service_bindings": schema.SetNestedAttribute{
				MarkdownDescription: "Service instances to bind to the application.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.AlsoRequires(path.MatchRoot("service_bindings").AtAnySetValue().AtName("service_instance")),
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
					setplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"service_instance": schema.StringAttribute{
							MarkdownDescription: "The service instance name.",
							Optional:            true,
							Computed:            true,
						},
						"params": schema.StringAttribute{
							CustomType:          jsontypes.NormalizedType{},
							MarkdownDescription: "A json object to send to the service broker during service binding.",
							Optional:            true,
							Computed:            true,
							Default:             stringdefault.StaticString("{}"),
						},
					},
				},
			},
			"routes": schema.SetNestedAttribute{
				MarkdownDescription: "The routes to map to the application to control its ingress traffic.",
				Optional:            true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.AlsoRequires(path.MatchRoot("routes").AtAnySetValue().AtName("route")),
				},
				Computed: true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplaceIfConfigured(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"route": schema.StringAttribute{
							MarkdownDescription: "The fully route qualified domain name which will be bound to app",
							Optional:            true,
							Computed:            true,
						},
						"protocol": schema.StringAttribute{
							MarkdownDescription: "The protocol to use for the route. Valid values are http2, http1, and tcp.",
							Optional:            true,
							Computed:            true,
							Validators: []validator.String{
								stringvalidator.OneOf("http2", "http1", "tcp"),
							},
						},
					},
				},
			},
			"stopped": schema.BoolAttribute{
				MarkdownDescription: "Whether the application is started or stopped after creation. By default, this value is false, meaning the application will be started automatically after creation.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"environment": schema.MapAttribute{
				MarkdownDescription: "Key/value pairs of custom environment variables to set in your app. Does not include any system or service variables.",
				Optional:            true,
				Validators: []validator.Map{
					mapvalidator.SizeAtLeast(1),
				},
				ElementType: types.StringType,
			},
			"no_route": schema.BoolAttribute{
				MarkdownDescription: "The attribute with a value of true to prevent a route from being created for your app.",
				Optional:            true,
				Validators: []validator.Bool{
					boolvalidator.ConflictsWith(path.MatchRoot("routes")),
					boolvalidator.ConflictsWith(path.MatchRoot("random_route")),
				},
			},
			"random_route": schema.BoolAttribute{
				MarkdownDescription: "The random-route attribute to generate a unique route and avoid name collisions.",
				Optional:            true,
				Validators: []validator.Bool{
					boolvalidator.ConflictsWith(path.MatchRoot("routes")),
					boolvalidator.ConflictsWith(path.MatchRoot("no_route")),
				},
			},
			"app_deployed_running_timeout": schema.Int64Attribute{
				MarkdownDescription: "Timeout in minutes to wait for app to be running after updating deployment with 'blue-green' strategy. The default is 5 minutes. Min value is 1 minute. Used only when strategy is set to 'blue-green'.",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"app_deployed_running_check_interval": schema.Int64Attribute{
				MarkdownDescription: "The interval in seconds between checks to see if the app is running after updating deployment with 'blue-green' strategy. The default is 5 seconds. Min value is 1 second, max value is 30 seconds. Used only when strategy is set to 'blue-green'.",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
					int64validator.AtMost(30),
				},
			},
			"processes": schema.SetNestedAttribute{
				MarkdownDescription: "List of configurations for individual process types.",
				Optional:            true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.ConflictsWith(path.MatchRoot("command")),
					setvalidator.ConflictsWith(path.MatchRoot("disk_quota")),
					setvalidator.ConflictsWith(path.MatchRoot("health_check_http_endpoint")),
					setvalidator.ConflictsWith(path.MatchRoot("health_check_invocation_timeout")),
					setvalidator.ConflictsWith(path.MatchRoot("health_check_type")),
					setvalidator.ConflictsWith(path.MatchRoot("health_check_interval")),
					setvalidator.ConflictsWith(path.MatchRoot("readiness_health_check_type")),
					setvalidator.ConflictsWith(path.MatchRoot("readiness_health_check_http_endpoint")),
					setvalidator.ConflictsWith(path.MatchRoot("readiness_health_check_invocation_timeout")),
					setvalidator.ConflictsWith(path.MatchRoot("readiness_health_check_interval")),
					setvalidator.ConflictsWith(path.MatchRoot("log_rate_limit_per_second")),
					setvalidator.ConflictsWith(path.MatchRoot("instances")),
					setvalidator.ConflictsWith(path.MatchRoot("memory")),
					setvalidator.ConflictsWith(path.MatchRoot("timeout")),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: r.ProcessSchemaAttributes(),
				},
			},
			"sidecars": schema.SetNestedAttribute{
				MarkdownDescription: "The attribute specifies additional processes to run in the same container as your app",
				Optional:            true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Sidecar name. The identifier for the sidecars to be configured.",
							Required:            true,
						},
						"command": schema.StringAttribute{
							MarkdownDescription: "The command used to start the sidecar.",
							Optional:            true,
						},
						"process_types": schema.SetAttribute{
							MarkdownDescription: "List of processes to associate sidecar with.",
							ElementType:         types.StringType,
							Optional:            true,
							Validators: []validator.Set{
								setvalidator.SizeAtLeast(1),
							},
						},
						"memory": schema.StringAttribute{
							MarkdownDescription: "The memory limit for the sidecar.",
							Optional:            true,
						},
					},
				},
			},
			labelsKey:      resourceLabelsSchema(),
			annotationsKey: resourceAnnotationsSchema(),
			idKey: schema.StringAttribute{
				MarkdownDescription: "The GUID of the object.",
				Computed:            true,
			},
			createdAtKey: schema.StringAttribute{
				MarkdownDescription: "The date and time when the resource was created in [RFC3339](https://www.ietf.org/rfc/rfc3339.txt) format.",
				Computed:            true,
			},
			updatedAtKey: updatedAtSchema(),
		},
	}
	for k, v := range r.ProcessAppCommonSchema() {
		if _, ok := resp.Schema.Attributes[k]; !ok {
			resp.Schema.Attributes[k] = v
		}
	}
}

func (r *appResource) ProcessSchemaAttributes() map[string]schema.Attribute {
	pSchema := map[string]schema.Attribute{
		"type": schema.StringAttribute{
			MarkdownDescription: "The process type. Any string identifier is accepted (e.g., web, worker, scheduler).",
			Required:            true,
		},
	}
	for k, v := range r.ProcessAppCommonSchema() {
		if _, ok := pSchema[k]; !ok {
			pSchema[k] = v
		}
	}
	return pSchema
}
func (r *appResource) ProcessAppCommonSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"command": schema.StringAttribute{
			MarkdownDescription: "A custom start command for the application. This overrides the start command provided by the buildpack.",
			Optional:            true,
		},
		"disk_quota": schema.StringAttribute{
			MarkdownDescription: "The disk space to be allocated for each application instance.",
			Optional:            true,
			Computed:            true,
		},
		"health_check_http_endpoint": schema.StringAttribute{
			MarkdownDescription: "The endpoint for the http health check type.",
			Optional:            true,
		},
		"health_check_invocation_timeout": schema.Int64Attribute{
			MarkdownDescription: "The timeout in seconds for the health check requests for http and port health checks.",
			Optional:            true,
		},
		"health_check_type": schema.StringAttribute{
			MarkdownDescription: "The health check type which can be one of 'port', 'process', 'http'.",
			Optional:            true,
			Computed:            true,
			Validators: []validator.String{
				stringvalidator.OneOf("port", "process", "http"),
			},
		},
		"health_check_interval": schema.Int64Attribute{
			MarkdownDescription: "The interval in seconds between health checks.",
			Optional:            true,
		},
		"readiness_health_check_type": schema.StringAttribute{
			MarkdownDescription: "The readiness health check type which can be one of 'port', 'process', 'http'.",
			Optional:            true,
			Computed:            true,
			Validators: []validator.String{
				stringvalidator.OneOf("port", "process", "http"),
			},
		},
		"readiness_health_check_http_endpoint": schema.StringAttribute{
			MarkdownDescription: "The endpoint for the http readiness health check type.",
			Optional:            true,
		},
		"readiness_health_check_invocation_timeout": schema.Int64Attribute{
			MarkdownDescription: "The timeout in seconds for the readiness health check requests for http and port health checks.",
			Optional:            true,
		},
		"readiness_health_check_interval": schema.Int64Attribute{
			MarkdownDescription: "The interval in seconds between readiness health checks.",
			Optional:            true,
		},
		"log_rate_limit_per_second": schema.StringAttribute{
			MarkdownDescription: "The attribute specifies the log rate limit for all instances of an app.",
			Computed:            true,
			Optional:            true,
		},
		"instances": schema.Int64Attribute{
			MarkdownDescription: "The number of app instances that you want to start. Defaults to 1.",
			Optional:            true,
			Computed:            true,
		},
		"memory": schema.StringAttribute{
			MarkdownDescription: "The memory limit for each application instance. If not provided, value is computed and retreived from Cloud Foundry.",
			Optional:            true,
			Computed:            true,
		},
		"timeout": schema.Int64Attribute{
			MarkdownDescription: "Time in seconds at which the health-check will report failure.",
			Optional:            true,
		},
	}
}
func (r *appResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	session, ok := req.ProviderData.(*managers.Session)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *managers.Session, got: %T. Please report this issue to the provider developers", req.ProviderData),
		)
		return
	}
	r.cfClient = session.CFClient
}

func (r *appResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	r.upsert(ctx, &req.Plan, nil, &resp.State, &resp.Diagnostics)
}

func (r *appResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var appType AppType
	diags := req.State.Get(ctx, &appType)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	appResp, err := r.cfClient.Applications.Get(ctx, appType.ID.ValueString())
	if err != nil {
		handleReadErrors(ctx, resp, err, "app", appType.ID.ValueString())
		return
	}
	appRaw, err := r.cfClient.Manifests.Generate(ctx, appType.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading app", err.Error())
		return
	}
	sshResp, err := r.cfClient.AppFeatures.GetSSH(ctx, appType.ID.ValueString())
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
	space, org, err := r.cfClient.Spaces.GetIncludeOrganization(ctx, appResp.Relationships.Space.Data.GUID)
	if err != nil {
		resp.Diagnostics.AddError("Error fetching space and org details : ", err.Error())
		return
	}
	plan, diags := mapAppValuesToType(ctx, appManifest.Applications[0], appResp, &appType, sshResp)
	resp.Diagnostics.Append(diags...)
	plan.CopyConfigAttributes(&appType)
	plan.Space = types.StringValue(space.Name)
	plan.Org = types.StringValue(org.Name)
	if plan.Stopped.IsNull() || plan.Stopped.IsUnknown() {
		plan.Stopped = appType.Stopped
	}
	resp.State.Set(ctx, &plan)
}

func (r *appResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.upsert(ctx, &req.Plan, &req.State, &resp.State, &resp.Diagnostics)
}
func (r *appResource) upsert(ctx context.Context, reqPlan *tfsdk.Plan, reqState *tfsdk.State, respState *tfsdk.State, respDiags *diag.Diagnostics) {
	var (
		desiredState, previousState AppType
		envs                        map[string]*string
	)
	envs = make(map[string]*string)
	diags := reqPlan.Get(ctx, &desiredState)
	respDiags.Append(diags...)
	if respDiags.HasError() {
		return
	}
	appManifestValue, diags := desiredState.mapAppTypeToValues(ctx)
	respDiags.Append(diags...)
	if respDiags.HasError() {
		return
	}
	//Required to set instances value for rolling strategy app update for go-cfclient. So setting it to cf-default of 1
	if appManifestValue.Instances == nil {
		appManifestValue.Instances = uinttouintptr(1)
	}
	if reqState != nil {
		diags = reqState.Get(ctx, &previousState)
		respDiags.Append(diags...)
		appManifestValue.Metadata, diags = setClientMetadataForUpdate(ctx, previousState.Labels, previousState.Annotations, desiredState.Labels, desiredState.Annotations)
		respDiags.Append(diags...)
		envs, diags = setEnvForUpdate(ctx, previousState.Environment, desiredState.Environment)
		respDiags.Append(diags...)
	}

	curTime := time.Now()
	appResp, err := r.push(desiredState, appManifestValue, ctx)

	if err != nil {

		errString := getAppLogTrace(ctx, r, desiredState, curTime)
		respDiags.AddError("Error pushing app", err.Error()+"\n"+strings.Join(errString, ""))
		return
	}

	_, err = r.cfClient.Applications.SetEnvironmentVariables(ctx, appResp.GUID, envs)
	if err != nil {
		respDiags.AddError("Error setting environment variables", err.Error())
		return
	}

	sshResp, err := r.cfClient.AppFeatures.UpdateSSH(ctx, appResp.GUID, desiredState.EnableSSH.ValueBool())
	if err != nil {
		respDiags.AddError("Error setting space feature", err.Error())
		return
	}

	manifestRespRaw, err := r.cfClient.Manifests.Generate(ctx, appResp.GUID)
	if err != nil {
		respDiags.AddError("Error generating manifest", err.Error())
	}
	var manifest *cfv3operation.Manifest
	err = yaml.Unmarshal([]byte(manifestRespRaw), &manifest)
	if err != nil {
		respDiags.AddError("Error unmarshalling manifest", err.Error())
	}
	plan, diags := mapAppValuesToType(ctx, manifest.Applications[0], appResp, &desiredState, sshResp)
	respDiags.Append(diags...)
	plan.CopyConfigAttributes(&desiredState)
	plan.Stopped = desiredState.Stopped
	respDiags.Append(respState.Set(ctx, &plan)...)
}
func (r *appResource) push(appType AppType, appManifestValue *cfv3operation.AppManifest, ctx context.Context) (*cfv3resource.App, error) {
	var file *os.File
	var err error
	if !appType.Path.IsNull() {
		file, err = os.Open(appType.Path.ValueString())
		if err != nil {
			return nil, err
		}
	}
	manifestOp := cfv3operation.NewAppPushOperation(r.cfClient, appType.Org.ValueString(), appType.Space.ValueString())
	if !appType.Strategy.IsNull() {
		switch appType.Strategy.ValueString() {
		case "rolling":
			manifestOp.WithStrategy(cfv3operation.StrategyRolling)
		case "blue-green":
			timeout, checkInterval := r.getBlueGreenDeploymentStrategyOptions(appType)
			manifestOp.WithBlueGreenStrategy(timeout, checkInterval)
		default:
			manifestOp.WithStrategy(cfv3operation.StrategyNone)
		}
	}
	if appType.Stopped.ValueBool() {
		manifestOp.WithNoStart(true)
	}
	appResp, err := manifestOp.Push(ctx, appManifestValue, file)
	if err != nil {
		return nil, err
	}
	return appResp, nil
}

func (r *appResource) getBlueGreenDeploymentStrategyOptions(appType AppType) (uint, uint) {
	timeout := uint(0)
	if !appType.AppDeployedRunningTimeout.IsNull() {
		timeout = uint(appType.AppDeployedRunningTimeout.ValueInt64())
	}
	checkInterval := uint(0)
	if !appType.AppDeployedRunningCheckInterval.IsNull() {
		checkInterval = uint(appType.AppDeployedRunningCheckInterval.ValueInt64())
	}
	return timeout, checkInterval
}

func (r *appResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var appType AppType
	diags := req.State.Get(ctx, &appType)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	jobID, err := r.cfClient.Applications.Delete(ctx, appType.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete application",
			fmt.Sprintf("Application deletion verification failed %s with %s", appType.Name.ValueString(), err.Error()),
		)
		return
	}
	err = pollJob(ctx, *r.cfClient, jobID, defaultTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to verify org quota deletion",
			"Application deletion verification failed for "+appType.ID.ValueString()+": "+err.Error(),
		)
		return
	}
}

func (r *appResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func getAppLogTrace(ctx context.Context, r *appResource, desiredState AppType, curTime time.Time) []string {

	EnableCFAppLogTrace := os.Getenv("ENABLE_CF_APP_LOG_TRACE")
	var errString []string
	var logCacheAddr string
	if EnableCFAppLogTrace == "true" {

		org, _ := r.cfClient.Organizations.Single(ctx, &cfv3client.OrganizationListOptions{
			Names: cfv3client.Filter{
				Values: []string{desiredState.Org.ValueString()},
			},
		})
		space, _ := r.cfClient.Spaces.Single(ctx, &cfv3client.SpaceListOptions{
			Names: cfv3client.Filter{
				Values: []string{desiredState.Space.ValueString()},
			},
			OrganizationGUIDs: cfv3client.Filter{
				Values: []string{org.GUID},
			},
		})
		app, _ := r.cfClient.Applications.First(ctx, &cfv3client.AppListOptions{
			Names: cfv3client.Filter{
				Values: []string{desiredState.Name.ValueString()},
			},
			OrganizationGUIDs: cfv3client.Filter{
				Values: []string{org.GUID},
			},
			SpaceGUIDs: cfv3client.Filter{
				Values: []string{space.GUID},
			},
		})

		apiURL := app.Links.Self().Href
		parsedURL, err := url.Parse(apiURL)
		if err != nil {
			errString = append(errString, "Error in getting cf log: "+err.Error())
			return errString
		}
		host := parsedURL.Host
		if strings.HasPrefix(host, "api.") {
			logCacheAddr = strings.Replace(host, "api.", "log-cache.", 1)
			logCacheAddr = "https://" + logCacheAddr
		}

		log_client := logclient.NewClient(logCacheAddr, logclient.WithHTTPClient(r.cfClient.HTTPAuthClient()))
		es, err := log_client.Read(ctx, app.GUID, curTime)
		if err != nil {
			errString = append(errString, "Error in getting cf log: "+err.Error())
			return errString
		}
		for _, e := range es {
			logMessage := e.GetLog()
			if logMessage.GetPayload() != nil && logMessage.GetType().String() == "ERR" {
				payload := string(logMessage.GetPayload())
				errString = append(errString, payload+"\n")
			}
		}
	}
	return errString
}
