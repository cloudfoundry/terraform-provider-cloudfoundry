package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	cfv3operation "github.com/cloudfoundry/go-cfclient/v3/operation"
	cfv3resource "github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/samber/lo"
)

// Type AppType representing Schema Attribute from function Schema in go type from resource_appManifest.go file.
type AppType struct {
	Name                                  types.String       `tfsdk:"name"`
	Space                                 types.String       `tfsdk:"space_name"`
	Org                                   types.String       `tfsdk:"org_name"`
	EnableSSH                             types.Bool         `tfsdk:"enable_ssh"`
	Stack                                 types.String       `tfsdk:"stack"`
	Buildpacks                            types.List         `tfsdk:"buildpacks"`
	Path                                  types.String       `tfsdk:"path"`
	SourceCodeHash                        types.String       `tfsdk:"source_code_hash"`
	DockerImage                           types.String       `tfsdk:"docker_image"`
	DockerCredentials                     *DockerCredentials `tfsdk:"docker_credentials"`
	Strategy                              types.String       `tfsdk:"strategy"`
	ServiceBindings                       types.Set          `tfsdk:"service_bindings"`
	Routes                                types.Set          `tfsdk:"routes"`
	Stopped                               types.Bool         `tfsdk:"stopped"`
	Environment                           types.Map          `tfsdk:"environment"`
	HealthCheckInterval                   types.Int64        `tfsdk:"health_check_interval"`
	ReadinessHealthCheckType              types.String       `tfsdk:"readiness_health_check_type"`
	ReadinessHealthCheckHttpEndpoint      types.String       `tfsdk:"readiness_health_check_http_endpoint"`
	ReadinessHealthCheckInvocationTimeout types.Int64        `tfsdk:"readiness_health_check_invocation_timeout"`
	ReadinessHealthCheckInterval          types.Int64        `tfsdk:"readiness_health_check_interval"`
	LogRateLimitPerSecond                 types.String       `tfsdk:"log_rate_limit_per_second"`
	NoRoute                               types.Bool         `tfsdk:"no_route"`
	RandomRoute                           types.Bool         `tfsdk:"random_route"`
	AppDeployedRunningTimeout             types.Int64        `tfsdk:"app_deployed_running_timeout"`
	AppDeployedRunningCheckInterval       types.Int64        `tfsdk:"app_deployed_running_check_interval"`
	Processes                             []Process          `tfsdk:"processes"`
	Sidecars                              []Sidecar          `tfsdk:"sidecars"`
	ID                                    types.String       `tfsdk:"id"`
	CreatedAt                             types.String       `tfsdk:"created_at"`
	UpdatedAt                             types.String       `tfsdk:"updated_at"`
	Command                               types.String       `tfsdk:"command"`
	DiskQuota                             types.String       `tfsdk:"disk_quota"`
	HealthCheckHttpEndpoint               types.String       `tfsdk:"health_check_http_endpoint"`
	HealthCheckInvocationTimeout          types.Int64        `tfsdk:"health_check_invocation_timeout"`
	HealthCheckType                       types.String       `tfsdk:"health_check_type"`
	Instances                             types.Int64        `tfsdk:"instances"`
	Memory                                types.String       `tfsdk:"memory"`
	Timeout                               types.Int64        `tfsdk:"timeout"`
	Labels                                types.Map          `tfsdk:"labels"`
	Annotations                           types.Map          `tfsdk:"annotations"`
}

type DatasourceAppType struct {
	Name                                  types.String                 `tfsdk:"name"`
	Space                                 types.String                 `tfsdk:"space_name"`
	Org                                   types.String                 `tfsdk:"org_name"`
	EnableSSH                             types.Bool                   `tfsdk:"enable_ssh"`
	Stack                                 types.String                 `tfsdk:"stack"`
	Buildpacks                            types.List                   `tfsdk:"buildpacks"`
	DockerImage                           types.String                 `tfsdk:"docker_image"`
	DockerCredentials                     *DataSourceDockerCredentials `tfsdk:"docker_credentials"`
	ServiceBindings                       types.Set                    `tfsdk:"service_bindings"`
	Routes                                types.Set                    `tfsdk:"routes"`
	Environment                           types.Map                    `tfsdk:"environment"`
	HealthCheckInterval                   types.Int64                  `tfsdk:"health_check_interval"`
	ReadinessHealthCheckType              types.String                 `tfsdk:"readiness_health_check_type"`
	ReadinessHealthCheckHttpEndpoint      types.String                 `tfsdk:"readiness_health_check_http_endpoint"`
	ReadinessHealthCheckInvocationTimeout types.Int64                  `tfsdk:"readiness_health_check_invocation_timeout"`
	ReadinessHealthCheckInterval          types.Int64                  `tfsdk:"readiness_health_check_interval"`
	LogRateLimitPerSecond                 types.String                 `tfsdk:"log_rate_limit_per_second"`
	Processes                             []Process                    `tfsdk:"processes"`
	Sidecars                              []Sidecar                    `tfsdk:"sidecars"`
	ID                                    types.String                 `tfsdk:"id"`
	CreatedAt                             types.String                 `tfsdk:"created_at"`
	UpdatedAt                             types.String                 `tfsdk:"updated_at"`
	Command                               types.String                 `tfsdk:"command"`
	DiskQuota                             types.String                 `tfsdk:"disk_quota"`
	HealthCheckHttpEndpoint               types.String                 `tfsdk:"health_check_http_endpoint"`
	HealthCheckInvocationTimeout          types.Int64                  `tfsdk:"health_check_invocation_timeout"`
	HealthCheckType                       types.String                 `tfsdk:"health_check_type"`
	Instances                             types.Int64                  `tfsdk:"instances"`
	Memory                                types.String                 `tfsdk:"memory"`
	Timeout                               types.Int64                  `tfsdk:"timeout"`
	Labels                                types.Map                    `tfsdk:"labels"`
	Annotations                           types.Map                    `tfsdk:"annotations"`
}

type DatasourceAppsType struct {
	Space types.String        `tfsdk:"space"`
	Org   types.String        `tfsdk:"org"`
	Name  types.String        `tfsdk:"name"`
	Apps  []DatasourceAppType `tfsdk:"apps"`
}

func (a *DatasourceAppType) Expand() AppType {
	var expanded AppType
	copyFields(&expanded, a)
	return expanded
}

type Sidecar struct {
	Name         types.String `tfsdk:"name"`
	Command      types.String `tfsdk:"command"`
	ProcessTypes types.Set    `tfsdk:"process_types"`
	Memory       types.String `tfsdk:"memory"`
}
type Process struct {
	Type                                  types.String `tfsdk:"type"`
	Command                               types.String `tfsdk:"command"`
	DiskQuota                             types.String `tfsdk:"disk_quota"`
	HealthCheckHttpEndpoint               types.String `tfsdk:"health_check_http_endpoint"`
	HealthCheckInvocationTimeout          types.Int64  `tfsdk:"health_check_invocation_timeout"`
	HealthCheckType                       types.String `tfsdk:"health_check_type"`
	Instances                             types.Int64  `tfsdk:"instances"`
	Memory                                types.String `tfsdk:"memory"`
	Timeout                               types.Int64  `tfsdk:"timeout"`
	HealthCheckInterval                   types.Int64  `tfsdk:"health_check_interval"`
	ReadinessHealthCheckType              types.String `tfsdk:"readiness_health_check_type"`
	ReadinessHealthCheckHttpEndpoint      types.String `tfsdk:"readiness_health_check_http_endpoint"`
	ReadinessHealthCheckInvocationTimeout types.Int64  `tfsdk:"readiness_health_check_invocation_timeout"`
	ReadinessHealthCheckInterval          types.Int64  `tfsdk:"readiness_health_check_interval"`
	LogRateLimitPerSecond                 types.String `tfsdk:"log_rate_limit_per_second"`
}

type DockerCredentials struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

type DataSourceDockerCredentials struct {
	Username types.String `tfsdk:"username"`
}

type ServiceBinding struct {
	ServiceInstance types.String         `tfsdk:"service_instance"`
	Params          jsontypes.Normalized `tfsdk:"params"`
}

var serviceBindingObjType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"service_instance": types.StringType,
		"params":           jsontypes.NormalizedType{},
	},
}

type Route struct {
	Route    types.String `tfsdk:"route"`
	Protocol types.String `tfsdk:"protocol"`
}

var routeObjType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"route":    types.StringType,
		"protocol": types.StringType,
	},
}

// mapAppTypeToValues function maps AppType to cfv3resource manifest type.
func (appType *AppType) mapAppTypeToValues(ctx context.Context) (*cfv3operation.AppManifest, diag.Diagnostics) {
	var diags, tempDiags diag.Diagnostics
	appmanifest := cfv3operation.AppManifest{}
	appmanifest.Name = appType.Name.ValueString()
	if !appType.Stack.IsUnknown() {
		appmanifest.Stack = appType.Stack.ValueString()
	}
	if !appType.Buildpacks.IsUnknown() {
		var buildpacks []string
		tempDiags = appType.Buildpacks.ElementsAs(ctx, &buildpacks, false)
		diags = append(diags, tempDiags...)
		appmanifest.Buildpacks = buildpacks
	}
	if !appType.DockerImage.IsNull() {
		appManifestDocker := cfv3operation.AppManifestDocker{
			Image: appType.DockerImage.ValueString(),
		}
		if appType.DockerCredentials != nil {
			appManifestDocker.Username = appType.DockerCredentials.Username.ValueString()
			err := os.Setenv("CF_DOCKER_PASSWORD", appType.DockerCredentials.Password.ValueString())
			if err != nil {
				tempDiags.AddError("Error setting docker password", err.Error())
				diags = append(diags, tempDiags...)
			}
		}
		appmanifest.Docker = &appManifestDocker
	}
	if !appType.ServiceBindings.IsUnknown() {
		var services cfv3operation.AppManifestServices
		tfServiceBindings := []ServiceBinding{}
		tempDiags = appType.ServiceBindings.ElementsAs(ctx, &tfServiceBindings, false)
		diags = append(diags, tempDiags...)
		for _, service := range tfServiceBindings {
			serviceManifest := cfv3operation.AppManifestService{
				Name: service.ServiceInstance.ValueString(),
			}
			if !service.Params.IsNull() {
				var params json.RawMessage
				err := json.Unmarshal([]byte(service.Params.ValueString()), &params)
				if err != nil {
					tempDiags.AddError("Error unmarshalling service params", err.Error())
					diags = append(diags, tempDiags...)
				}
				var finalParams map[string]interface{}
				err = json.Unmarshal(params, &finalParams)
				if err != nil {
					tempDiags.AddError("Error unmarshalling service params", err.Error())
					diags = append(diags, tempDiags...)
				}
				serviceManifest.Parameters = finalParams
			}
			services = append(services, serviceManifest)
		}
		appmanifest.Services = &services
	}
	if !appType.Routes.IsUnknown() {
		var routes cfv3operation.AppManifestRoutes
		tfRoutes := []Route{}
		tempDiags = appType.Routes.ElementsAs(ctx, &tfRoutes, false)
		diags = append(diags, tempDiags...)
		for _, route := range tfRoutes {
			routeManifest := cfv3operation.AppManifestRoute{
				Route: route.Route.ValueString(),
			}
			if !route.Protocol.IsUnknown() {
				routeManifest.Protocol = cfv3operation.AppRouteProtocol(route.Protocol.ValueString())
			}
			routes = append(routes, routeManifest)
		}
		appmanifest.Routes = &routes
	}
	if !appType.Environment.IsNull() {
		var env map[string]string
		tempDiags = appType.Environment.ElementsAs(ctx, &env, false)
		diags = append(diags, tempDiags...)
		appmanifest.Env = env
	}
	if !appType.HealthCheckInterval.IsNull() {
		appmanifest.HealthCheckInterval = uint(appType.HealthCheckInterval.ValueInt64())
	}
	if !appType.ReadinessHealthCheckType.IsUnknown() {
		appmanifest.ReadinessHealthCheckType = appType.ReadinessHealthCheckType.ValueString()
	}
	if !appType.ReadinessHealthCheckHttpEndpoint.IsNull() {
		appmanifest.ReadinessHealthCheckHttpEndpoint = appType.ReadinessHealthCheckHttpEndpoint.ValueString()
	}
	if !appType.ReadinessHealthCheckInvocationTimeout.IsNull() {
		appmanifest.ReadinessHealthInvocationTimeout = uint(appType.ReadinessHealthCheckInvocationTimeout.ValueInt64())
	}
	if !appType.ReadinessHealthCheckInterval.IsNull() {
		appmanifest.ReadinessHealthCheckInterval = uint(appType.ReadinessHealthCheckInterval.ValueInt64())
	}
	if !appType.LogRateLimitPerSecond.IsUnknown() {
		appmanifest.LogRateLimitPerSecond = appType.LogRateLimitPerSecond.ValueString()
	}
	if !appType.NoRoute.IsNull() {
		appmanifest.NoRoute = appType.NoRoute.ValueBool()
	}
	if !appType.RandomRoute.IsNull() {
		appmanifest.RandomRoute = appType.RandomRoute.ValueBool()
	}
	if len(appType.Processes) != 0 {
		var processes cfv3operation.AppManifestProcesses
		for _, process := range appType.Processes {
			processManifest := cfv3operation.AppManifestProcess{}
			processManifest.Type = cfv3operation.AppProcessType(process.Type.ValueString())
			if !process.Command.IsNull() {
				processManifest.Command = process.Command.ValueString()
			}
			if !process.DiskQuota.IsUnknown() {
				processManifest.DiskQuota = process.DiskQuota.ValueString()
			}
			if !process.HealthCheckHttpEndpoint.IsNull() {
				processManifest.HealthCheckHTTPEndpoint = process.HealthCheckHttpEndpoint.ValueString()
			}
			if !process.HealthCheckInvocationTimeout.IsNull() {
				processManifest.HealthCheckInvocationTimeout = uint(process.HealthCheckInvocationTimeout.ValueInt64())
			}
			if !process.HealthCheckType.IsUnknown() {
				processManifest.HealthCheckType = cfv3operation.AppHealthCheckType(process.HealthCheckType.ValueString())
			}
			if !process.HealthCheckInterval.IsNull() {
				processManifest.HealthCheckInterval = uint(process.HealthCheckInterval.ValueInt64())
			}
			if !process.Instances.IsUnknown() {
				processManifest.Instances = uinttouintptr(uint(process.Instances.ValueInt64()))
			}
			if !process.Memory.IsUnknown() {
				processManifest.Memory = process.Memory.ValueString()
			}
			if !process.Timeout.IsNull() {
				processManifest.Timeout = uint(process.Timeout.ValueInt64())
			}
			if !process.ReadinessHealthCheckType.IsUnknown() {
				processManifest.ReadinessHealthCheckType = process.ReadinessHealthCheckType.ValueString()
			}
			if !process.ReadinessHealthCheckHttpEndpoint.IsNull() {
				processManifest.ReadinessHealthCheckHttpEndpoint = process.ReadinessHealthCheckHttpEndpoint.ValueString()
			}
			if !process.ReadinessHealthCheckInvocationTimeout.IsNull() {
				processManifest.ReadinessHealthInvocationTimeout = uint(process.ReadinessHealthCheckInvocationTimeout.ValueInt64())
			}
			if !process.ReadinessHealthCheckInterval.IsNull() {
				processManifest.ReadinessHealthCheckInterval = uint(process.ReadinessHealthCheckInterval.ValueInt64())
			}
			if !appType.LogRateLimitPerSecond.IsUnknown() {
				processManifest.LogRateLimitPerSecond = process.LogRateLimitPerSecond.ValueString()
			}
			processes = append(processes, processManifest)
		}
		appmanifest.Processes = &processes
	}
	if len(appType.Sidecars) != 0 {
		var sidecars cfv3operation.AppManifestSideCars
		for _, sidecar := range appType.Sidecars {
			sidecarManifest := cfv3operation.AppManifestSideCar{
				Name: sidecar.Name.ValueString(),
			}
			if !sidecar.Command.IsNull() {
				sidecarManifest.Command = sidecar.Command.ValueString()
			}
			if !sidecar.ProcessTypes.IsNull() {
				var processTypes []string
				tempDiags = sidecar.ProcessTypes.ElementsAs(ctx, &processTypes, false)
				diags = append(diags, tempDiags...)
				sidecarManifest.ProcessTypes = processTypes
			}
			if !sidecar.Memory.IsNull() {
				sidecarManifest.Memory = sidecar.Memory.ValueString()
			}
			sidecars = append(sidecars, sidecarManifest)
		}
		appmanifest.Sidecars = &sidecars
	}
	if !appType.Command.IsNull() {
		appmanifest.Command = appType.Command.ValueString()
	}
	if !appType.DiskQuota.IsUnknown() {
		appmanifest.DiskQuota = appType.DiskQuota.ValueString()
	}
	if !appType.HealthCheckHttpEndpoint.IsUnknown() {
		appmanifest.HealthCheckHTTPEndpoint = appType.HealthCheckHttpEndpoint.ValueString()
	}
	if !appType.HealthCheckInvocationTimeout.IsNull() {
		appmanifest.HealthCheckInvocationTimeout = uint(appType.HealthCheckInvocationTimeout.ValueInt64())
	}
	if !appType.HealthCheckType.IsUnknown() {
		appmanifest.HealthCheckType = cfv3operation.AppHealthCheckType(appType.HealthCheckType.ValueString())
	}
	if !appType.Instances.IsUnknown() {
		appmanifest.Instances = uinttouintptr(uint(appType.Instances.ValueInt64()))
	}
	if !appType.Memory.IsUnknown() {
		appmanifest.Memory = appType.Memory.ValueString()
	}
	if !appType.Timeout.IsNull() {
		appmanifest.Timeout = uint(appType.Timeout.ValueInt64())
	}
	appmanifest.Metadata = cfv3resource.NewMetadata()
	if !appType.Labels.IsNull() {
		tempDiags = appType.Labels.ElementsAs(ctx, &appmanifest.Metadata.Labels, false)
		diags = append(diags, tempDiags...)
	}
	if !appType.Annotations.IsNull() {
		tempDiags = appType.Annotations.ElementsAs(ctx, &appmanifest.Metadata.Annotations, false)
		diags = append(diags, tempDiags...)
	}
	return &appmanifest, diags
}

// mapAppValuesToType function maps cfv3resource manifest type to AppType
/*
	reqPlanType is required here to identify whether attributes like "health-check-interval", "readiness-health-check-interval"
	are present as part of app spec or not, since cf api controller converts them to be part of process spec internally
*/
func mapAppValuesToType(ctx context.Context, appManifest *cfv3operation.AppManifest, app *cfv3resource.App, reqPlanType *AppType, sshResp *cfv3resource.AppFeature) (AppType, diag.Diagnostics) {
	var diags, tempDiags diag.Diagnostics
	var appType AppType
	appType.Name = types.StringValue(appManifest.Name)
	appType.Stack = types.StringValue(appManifest.Stack)
	if len(appManifest.Buildpacks) != 0 {
		appType.Buildpacks, tempDiags = types.ListValueFrom(ctx, types.StringType, appManifest.Buildpacks)
		diags = append(diags, tempDiags...)
	} else {
		appType.Buildpacks = types.ListNull(types.StringType)
	}
	if appManifest.Docker != nil {
		appType.DockerImage = types.StringValue(appManifest.Docker.Image)
		if appManifest.Docker.Username != "" {
			appType.DockerCredentials = &DockerCredentials{}
			appType.DockerCredentials.Username = types.StringValue(appManifest.Docker.Username)
			appType.DockerCredentials.Password = reqPlanType.DockerCredentials.Password
		}
	}
	if appManifest.Services != nil {
		var serviceBindings []ServiceBinding
		for _, service := range *appManifest.Services {
			var sb ServiceBinding
			sb.ServiceInstance = types.StringValue(service.Name)
			if service.Parameters != nil {
				param, err := json.Marshal(service.Parameters)
				if err != nil {
					tempDiags.AddError("Error marshalling service parameters", err.Error())
					diags = append(diags, tempDiags...)
				}
				sb.Params = jsontypes.NewNormalizedValue(string(param))
				diags = append(diags, tempDiags...)
			} else {
				sb.Params = jsontypes.NewNormalizedNull()
				if reqPlanType != nil {
					tfServiceBindings := []ServiceBinding{}
					tempDiags = reqPlanType.ServiceBindings.ElementsAs(ctx, &tfServiceBindings, false)
					diags = append(diags, tempDiags...)
					binding, found := lo.Find(tfServiceBindings, func(binding ServiceBinding) bool {
						return sb.ServiceInstance.Equal(binding.ServiceInstance)
					})
					if found {
						sb.Params = binding.Params
					}
				}
			}
			serviceBindings = append(serviceBindings, sb)
		}
		appType.ServiceBindings, tempDiags = types.SetValueFrom(ctx, serviceBindingObjType, serviceBindings)
		diags = append(diags, tempDiags...)
	} else {
		appType.ServiceBindings = types.SetNull(serviceBindingObjType)
	}
	if appManifest.Routes != nil {
		var routes []Route
		for _, route := range *appManifest.Routes {
			var r Route
			r.Route = types.StringValue(route.Route)
			if route.Protocol != "" {
				r.Protocol = types.StringValue(string(route.Protocol))
			}
			routes = append(routes, r)
		}
		appType.Routes, diags = types.SetValueFrom(ctx, routeObjType, routes)
	} else {
		appType.Routes = types.SetNull(routeObjType)
	}
	if appManifest.Env != nil {
		appType.Environment, tempDiags = types.MapValueFrom(ctx, types.StringType, appManifest.Env)
		diags = append(diags, tempDiags...)
	} else {
		appType.Environment = types.MapNull(types.StringType)
	}
	if appManifest.Processes != nil {
		var planProcessesByType map[string]*Process
		if reqPlanType != nil && len(reqPlanType.Processes) > 0 {
			planProcessesByType = make(map[string]*Process, len(reqPlanType.Processes))
			for i := range reqPlanType.Processes {
				proc := &reqPlanType.Processes[i]
				if proc.Type.IsNull() || proc.Type.IsUnknown() {
					continue
				}
				planProcessesByType[proc.Type.ValueString()] = proc
			}
		}
		var processes []Process
		for _, process := range *appManifest.Processes {
			plannedProcess := (*Process)(nil)
			if len(planProcessesByType) > 0 {
				plannedProcess = planProcessesByType[string(process.Type)]
			}
			// reqPlanType will be set only for resources else nil
			// we also check if processes were not set in request but we have in response then map to app spec level else map to process spec level
			if reqPlanType != nil && len(reqPlanType.Processes) == 0 {
				if process.Command != "" {
					appType.Command = types.StringValue(process.Command)
				}
				if process.DiskQuota != "" {
					if reqPlanType != nil && !reqPlanType.DiskQuota.IsNull() && !reqPlanType.DiskQuota.IsUnknown() {
						result, err := getDesiredType(process.DiskQuota, reqPlanType.DiskQuota.ValueString())
						if err != nil {
							tempDiags.AddError("Error converting disk quota", err.Error())
							diags = append(diags, tempDiags...)
						}
						appType.DiskQuota = types.StringValue(result)
					} else {
						appType.DiskQuota = types.StringValue(process.DiskQuota)
					}
				}
				if process.HealthCheckType != "" {
					appType.HealthCheckType = types.StringValue(string(process.HealthCheckType))
				}
				if process.HealthCheckHTTPEndpoint != "" {
					appType.HealthCheckHttpEndpoint = types.StringValue(process.HealthCheckHTTPEndpoint)
				}
				if process.HealthCheckInvocationTimeout != 0 {
					appType.HealthCheckInvocationTimeout = types.Int64Value(int64(process.HealthCheckInvocationTimeout))
				}
				appType.Instances = types.Int64Value(int64(*process.Instances))
				if process.Memory != "" {
					if reqPlanType != nil && !reqPlanType.Memory.IsNull() && !reqPlanType.Memory.IsUnknown() {
						result, err := getDesiredType(process.Memory, reqPlanType.Memory.ValueString())
						if err != nil {
							tempDiags.AddError("Error converting memory", err.Error())
							diags = append(diags, tempDiags...)
						}
						appType.Memory = types.StringValue(result)
					} else {
						appType.Memory = types.StringValue(process.Memory)
					}
				}
				if process.Timeout != 0 {
					appType.Timeout = types.Int64Value(int64(process.Timeout))
				}
				if process.HealthCheckInterval != 0 {
					appType.HealthCheckInterval = types.Int64Value(int64(process.HealthCheckInterval))
				}
				if process.ReadinessHealthCheckType != "" {
					appType.ReadinessHealthCheckType = types.StringValue(process.ReadinessHealthCheckType)
				}
				if process.ReadinessHealthCheckHttpEndpoint != "" {
					appType.ReadinessHealthCheckHttpEndpoint = types.StringValue(process.ReadinessHealthCheckHttpEndpoint)
				}
				if process.ReadinessHealthInvocationTimeout != 0 {
					appType.ReadinessHealthCheckInvocationTimeout = types.Int64Value(int64(process.HealthCheckInvocationTimeout))
				}
				if process.ReadinessHealthCheckInterval != 0 {
					appType.ReadinessHealthCheckInterval = types.Int64Value(int64(process.ReadinessHealthCheckInterval))
				}
				if process.LogRateLimitPerSecond != "" {
					if reqPlanType != nil && !reqPlanType.LogRateLimitPerSecond.IsNull() && !reqPlanType.LogRateLimitPerSecond.IsUnknown() {
						result, err := getDesiredType(process.LogRateLimitPerSecond, reqPlanType.LogRateLimitPerSecond.ValueString())
						if err != nil {
							tempDiags.AddError("Error converting log_rate_limit", err.Error())
							diags = append(diags, tempDiags...)
						}
						appType.LogRateLimitPerSecond = types.StringValue(result)
					} else {
						appType.LogRateLimitPerSecond = types.StringValue(process.LogRateLimitPerSecond)
					}
				}
			} else {
				var p Process
				p.Type = types.StringValue(string(process.Type))
				if process.Command != "" {
					p.Command = types.StringValue(process.Command)
				}
				if process.DiskQuota != "" {
					if plannedProcess != nil && !plannedProcess.DiskQuota.IsNull() && !plannedProcess.DiskQuota.IsUnknown() {
						result, err := getDesiredType(process.DiskQuota, plannedProcess.DiskQuota.ValueString())
						if err != nil {
							tempDiags.AddError("Error converting disk quota", err.Error())
							diags = append(diags, tempDiags...)
						}
						p.DiskQuota = types.StringValue(result)
					} else {
						p.DiskQuota = types.StringValue(process.DiskQuota)
					}
				}
				if process.HealthCheckHTTPEndpoint != "" {
					p.HealthCheckHttpEndpoint = types.StringValue(process.HealthCheckHTTPEndpoint)
				}
				if process.HealthCheckInvocationTimeout != 0 {
					p.HealthCheckInvocationTimeout = types.Int64Value(int64(process.HealthCheckInvocationTimeout))
				}
				if process.HealthCheckType != "" {
					p.HealthCheckType = types.StringValue(string(process.HealthCheckType))
				}
				p.Instances = types.Int64Value(int64(*process.Instances))
				if process.Memory != "" {
					if plannedProcess != nil && !plannedProcess.Memory.IsNull() && !plannedProcess.Memory.IsUnknown() {
						result, err := getDesiredType(process.Memory, plannedProcess.Memory.ValueString())
						if err != nil {
							tempDiags.AddError("Error converting memory", err.Error())
							diags = append(diags, tempDiags...)
						}
						p.Memory = types.StringValue(result)
					} else {
						p.Memory = types.StringValue(process.Memory)
					}
				}
				if process.Timeout != 0 {
					p.Timeout = types.Int64Value(int64(process.Timeout))
				}
				if process.HealthCheckInterval != 0 {
					p.HealthCheckInterval = types.Int64Value(int64(process.HealthCheckInterval))
				}
				if process.ReadinessHealthCheckType != "" {
					p.ReadinessHealthCheckType = types.StringValue(process.ReadinessHealthCheckType)
				}
				if process.ReadinessHealthCheckHttpEndpoint != "" {
					p.ReadinessHealthCheckHttpEndpoint = types.StringValue(process.ReadinessHealthCheckHttpEndpoint)
				}
				if process.ReadinessHealthInvocationTimeout != 0 {
					p.ReadinessHealthCheckInvocationTimeout = types.Int64Value(int64(process.HealthCheckInvocationTimeout))
				}
				if process.ReadinessHealthCheckInterval != 0 {
					p.ReadinessHealthCheckInterval = types.Int64Value(int64(process.ReadinessHealthCheckInterval))
				}
				if process.LogRateLimitPerSecond != "" {
					if plannedProcess != nil && !plannedProcess.LogRateLimitPerSecond.IsNull() && !plannedProcess.LogRateLimitPerSecond.IsUnknown() {
						result, err := getDesiredType(process.LogRateLimitPerSecond, plannedProcess.LogRateLimitPerSecond.ValueString())
						if err != nil {
							tempDiags.AddError("Error converting log_rate_limit", err.Error())
							diags = append(diags, tempDiags...)
						}
						p.LogRateLimitPerSecond = types.StringValue(result)
					} else {
						p.LogRateLimitPerSecond = types.StringValue(process.LogRateLimitPerSecond)
					}
				}
				processes = append(processes, p)
			}
		}
		appType.Processes = processes
	}
	if appManifest.Sidecars != nil {
		var sidecars []Sidecar
		for i, sidecar := range *appManifest.Sidecars {
			var s Sidecar
			s.Name = types.StringValue(sidecar.Name)
			if sidecar.Command != "" {
				s.Command = types.StringValue(sidecar.Command)
			}
			if len(sidecar.ProcessTypes) != 0 {
				s.ProcessTypes, tempDiags = types.SetValueFrom(ctx, types.StringType, sidecar.ProcessTypes)
				diags = append(diags, tempDiags...)
			} else {
				s.ProcessTypes = types.SetNull(types.StringType)
			}
			if sidecar.Memory != "" {
				if reqPlanType != nil && !reqPlanType.Sidecars[i].Memory.IsUnknown() {
					result, err := getDesiredType(sidecar.Memory, reqPlanType.Sidecars[i].Memory.ValueString())
					if err != nil {
						tempDiags.AddError("Error converting memory", err.Error())
						diags = append(diags, tempDiags...)
					}
					s.Memory = types.StringValue(result)
				} else {
					s.Memory = types.StringValue(sidecar.Memory)
				}
			}
			sidecars = append(sidecars, s)
		}
		appType.Sidecars = sidecars
	}
	appType.ID = types.StringValue(app.GUID)
	appType.CreatedAt = types.StringValue(app.CreatedAt.Format(time.RFC3339))
	appType.UpdatedAt = types.StringValue(app.UpdatedAt.Format(time.RFC3339))

	appType.Labels, tempDiags = mapMetadataValueToType(ctx, app.Metadata.Labels)
	diags = append(diags, tempDiags...)
	appType.Annotations, tempDiags = mapMetadataValueToType(ctx, app.Metadata.Annotations)
	diags = append(diags, tempDiags...)
	appType.EnableSSH = types.BoolValue(sshResp.Enabled)
	return appType, diags
}

func (target *AppType) CopyConfigAttributes(source *AppType) {
	target.Space = source.Space
	target.Org = source.Org
	target.Path = source.Path
	target.Strategy = source.Strategy
	target.SourceCodeHash = source.SourceCodeHash
	target.RandomRoute = source.RandomRoute
	target.NoRoute = source.NoRoute
	target.AppDeployedRunningTimeout = source.AppDeployedRunningTimeout
	target.AppDeployedRunningCheckInterval = source.AppDeployedRunningCheckInterval
}

func getDesiredType(actual string, desired string) (string, error) {
	// log-rate-limit-per-second accepts -1 & 0 as valid values
	// For more info https://v3-apidocs.cloudfoundry.org/version/3.159.0/index.html#the-manifest-schema
	if actual == "-1" || actual == "0" {
		return actual, nil
	}
	actualValue, _, err := splitValueAndUnit(actual)
	if err != nil {
		return "", err
	}
	// Considering actual to be ideal & convert desired type(From Plan) to ideal type
	calculatedValue, _, err := convertToDesiredType(desired, actual)
	// We take floor value of it just like cf controller does
	calculatedValue = math.Floor(calculatedValue)

	if calculatedValue == actualValue {
		return desired, err
	} else {
		return actual, err
	}
}

func convertToDesiredType(actual string, desired string) (float64, string, error) {
	// Find actual unit from string
	actualValue, actualUnit, err := splitValueAndUnit(actual)
	if err != nil {
		return 0, "", err
	}
	_, desiredUnit, err := splitValueAndUnit(desired)
	if err != nil {
		return 0, "", err
	}
	m := map[string]float64{
		"B":  0,
		"K":  1,
		"KB": 1,
		"M":  2,
		"MB": 2,
		"G":  3,
		"GB": 3,
		"T":  4,
		"TB": 4,
	}
	dist := m[actualUnit] - m[desiredUnit]
	res := actualValue * math.Pow(1024, dist)
	return res, desiredUnit, nil
}

// function to split the string into value and unit.
func splitValueAndUnit(value string) (float64, string, error) {
	var unit string
	var val float64
	_, err := fmt.Sscanf(value, "%f%s", &val, &unit)
	unit = strings.ToUpper(unit)
	if err != nil {
		return 0, "", err
	}
	return val, unit, nil
}

// Prepares the env for cfclient updation from existing and planned tfstate envs.
func setEnvForUpdate(ctx context.Context, existingEnvs basetypes.MapValue, plannedEnvs basetypes.MapValue) (map[string]*string, diag.Diagnostics) {

	var (
		diagnostics                 diag.Diagnostics
		oldEnvs, newEnvs, finalEnvs map[string]*string
	)

	finalEnvs = make(map[string]*string)

	diagnostics.Append(existingEnvs.ElementsAs(ctx, &oldEnvs, false)...)
	diagnostics.Append(plannedEnvs.ElementsAs(ctx, &newEnvs, false)...)

	for key := range oldEnvs {
		finalEnvs[key] = nil
	}

	for key, value := range newEnvs {
		finalEnvs[key] = value
	}

	return finalEnvs, diagnostics
}

func mapAppDatasourceValuesToType(ctx context.Context, appManifest *cfv3operation.AppManifest, app *cfv3resource.App, sshResp *cfv3resource.AppFeature) (DatasourceAppType, diag.Diagnostics) {
	var diags, tempDiags diag.Diagnostics
	var appType DatasourceAppType
	appType.Name = types.StringValue(appManifest.Name)
	appType.Stack = types.StringValue(appManifest.Stack)
	if len(appManifest.Buildpacks) != 0 {
		appType.Buildpacks, tempDiags = types.ListValueFrom(ctx, types.StringType, appManifest.Buildpacks)
		diags = append(diags, tempDiags...)
	} else {
		appType.Buildpacks = types.ListNull(types.StringType)
	}
	if appManifest.Docker != nil {
		appType.DockerImage = types.StringValue(appManifest.Docker.Image)
		if appManifest.Docker.Username != "" {
			appType.DockerCredentials = &DataSourceDockerCredentials{}
			appType.DockerCredentials.Username = types.StringValue(appManifest.Docker.Username)
		}
	}
	if appManifest.Services != nil {
		var serviceBindings []ServiceBinding
		for _, service := range *appManifest.Services {
			var sb ServiceBinding
			sb.ServiceInstance = types.StringValue(service.Name)
			if service.Parameters != nil {
				param, err := json.Marshal(service.Parameters)
				if err != nil {
					tempDiags.AddError("Error marshalling service parameters", err.Error())
					diags = append(diags, tempDiags...)
				}
				sb.Params = jsontypes.NewNormalizedValue(string(param))
				diags = append(diags, tempDiags...)
			} else {
				sb.Params = jsontypes.NewNormalizedNull()
			}
			serviceBindings = append(serviceBindings, sb)
		}
		appType.ServiceBindings, tempDiags = types.SetValueFrom(ctx, serviceBindingObjType, serviceBindings)
		diags = append(diags, tempDiags...)
	} else {
		appType.ServiceBindings = types.SetNull(serviceBindingObjType)
	}
	if appManifest.Routes != nil {
		var routes []Route
		for _, route := range *appManifest.Routes {
			var r Route
			r.Route = types.StringValue(route.Route)
			if route.Protocol != "" {
				r.Protocol = types.StringValue(string(route.Protocol))
			}
			routes = append(routes, r)
		}
		appType.Routes, diags = types.SetValueFrom(ctx, routeObjType, routes)
	} else {
		appType.Routes = types.SetNull(routeObjType)
	}
	if appManifest.Env != nil {
		appType.Environment, tempDiags = types.MapValueFrom(ctx, types.StringType, appManifest.Env)
		diags = append(diags, tempDiags...)
	} else {
		appType.Environment = types.MapNull(types.StringType)
	}
	if appManifest.Processes != nil {
		var processes []Process
		for _, process := range *appManifest.Processes {
			var p Process
			p.Type = types.StringValue(string(process.Type))
			if process.Command != "" {
				p.Command = types.StringValue(process.Command)
			}
			if process.DiskQuota != "" {
				p.DiskQuota = types.StringValue(process.DiskQuota)
			}
			if process.HealthCheckHTTPEndpoint != "" {
				p.HealthCheckHttpEndpoint = types.StringValue(process.HealthCheckHTTPEndpoint)
			}
			if process.HealthCheckInvocationTimeout != 0 {
				p.HealthCheckInvocationTimeout = types.Int64Value(int64(process.HealthCheckInvocationTimeout))
			}
			if process.HealthCheckType != "" {
				p.HealthCheckType = types.StringValue(string(process.HealthCheckType))
			}
			p.Instances = types.Int64Value(int64(*process.Instances))
			if process.Memory != "" {
				p.Memory = types.StringValue(process.Memory)
			}
			if process.Timeout != 0 {
				p.Timeout = types.Int64Value(int64(process.Timeout))
			}
			if process.HealthCheckInterval != 0 {
				p.HealthCheckInterval = types.Int64Value(int64(process.HealthCheckInterval))
			}
			if process.ReadinessHealthCheckType != "" {
				p.ReadinessHealthCheckType = types.StringValue(process.ReadinessHealthCheckType)
			}
			if process.ReadinessHealthCheckHttpEndpoint != "" {
				p.ReadinessHealthCheckHttpEndpoint = types.StringValue(process.ReadinessHealthCheckHttpEndpoint)
			}
			if process.ReadinessHealthInvocationTimeout != 0 {
				p.ReadinessHealthCheckInvocationTimeout = types.Int64Value(int64(process.HealthCheckInvocationTimeout))
			}
			if process.ReadinessHealthCheckInterval != 0 {
				p.ReadinessHealthCheckInterval = types.Int64Value(int64(process.ReadinessHealthCheckInterval))
			}
			if process.LogRateLimitPerSecond != "" {
				p.LogRateLimitPerSecond = types.StringValue(process.LogRateLimitPerSecond)
			}
			processes = append(processes, p)
		}
		appType.Processes = processes
	}
	if appManifest.Sidecars != nil {
		var sidecars []Sidecar
		for _, sidecar := range *appManifest.Sidecars {
			var s Sidecar
			s.Name = types.StringValue(sidecar.Name)
			if sidecar.Command != "" {
				s.Command = types.StringValue(sidecar.Command)
			}
			if len(sidecar.ProcessTypes) != 0 {
				s.ProcessTypes, tempDiags = types.SetValueFrom(ctx, types.StringType, sidecar.ProcessTypes)
				diags = append(diags, tempDiags...)
			} else {
				s.ProcessTypes = types.SetNull(types.StringType)
			}
			if sidecar.Memory != "" {
				s.Memory = types.StringValue(sidecar.Memory)
			}
			sidecars = append(sidecars, s)
		}
		appType.Sidecars = sidecars
	}
	appType.ID = types.StringValue(app.GUID)
	appType.CreatedAt = types.StringValue(app.CreatedAt.Format(time.RFC3339))
	appType.UpdatedAt = types.StringValue(app.UpdatedAt.Format(time.RFC3339))

	appType.Labels, tempDiags = mapMetadataValueToType(ctx, app.Metadata.Labels)
	diags = append(diags, tempDiags...)
	appType.Annotations, tempDiags = mapMetadataValueToType(ctx, app.Metadata.Annotations)
	diags = append(diags, tempDiags...)
	appType.EnableSSH = types.BoolValue(sshResp.Enabled)
	return appType, diags
}
