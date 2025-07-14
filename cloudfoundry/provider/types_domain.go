package provider

import (
	"context"
	"time"

	"github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type domainType struct {
	Name               types.String `tfsdk:"name"`
	Id                 types.String `tfsdk:"id"`
	Internal           types.Bool   `tfsdk:"internal"`
	RouterGroup        types.String `tfsdk:"router_group"`
	Org                types.String `tfsdk:"org"`
	SharedOrgs         types.Set    `tfsdk:"shared_orgs"`
	SupportedProtocols types.Set    `tfsdk:"supported_protocols"`
	Labels             types.Map    `tfsdk:"labels"`
	Annotations        types.Map    `tfsdk:"annotations"`
	CreatedAt          types.String `tfsdk:"created_at"`
	UpdatedAt          types.String `tfsdk:"updated_at"`
}

type domainsDatasourceType struct {
	Org     types.String `tfsdk:"org"`
	Domains []domainType `tfsdk:"domains"`
}

// Sets the terraform struct values from the domain resource returned by the cf-client.
func mapDomainValuesToType(ctx context.Context, domain *resource.Domain) (domainType, diag.Diagnostics) {

	domainType := domainType{
		Name:      types.StringValue(domain.Name),
		Id:        types.StringValue(domain.GUID),
		CreatedAt: types.StringValue(domain.CreatedAt.Format(time.RFC3339)),
		UpdatedAt: types.StringValue(domain.UpdatedAt.Format(time.RFC3339)),
		Internal:  types.BoolValue(domain.Internal),
	}
	if domain.RouterGroup != nil {
		domainType.RouterGroup = types.StringValue(domain.RouterGroup.GUID)
	}
	if domain.Relationships.Organization.Data != nil {
		domainType.Org = types.StringValue(domain.Relationships.Organization.Data.GUID)
	}

	var diags, diagnostics diag.Diagnostics
	domainType.Labels, diags = mapMetadataValueToType(ctx, domain.Metadata.Labels)
	diagnostics.Append(diags...)
	domainType.Annotations, diags = mapMetadataValueToType(ctx, domain.Metadata.Annotations)
	diagnostics.Append(diags...)

	domainType.SharedOrgs, diags = setRelationshipToTFSet(domain.Relationships.SharedOrganizations.Data)
	diagnostics.Append(diags...)
	domainType.SupportedProtocols, diags = types.SetValueFrom(ctx, types.StringType, domain.SupportedProtocols)
	diagnostics.Append(diags...)

	return domainType, diagnostics
}

// Sets the domain resource values for creation with cf-client from the terraform struct values.
func (data *domainType) mapCreateDomainTypeToValues(ctx context.Context) (resource.DomainCreate, diag.Diagnostics) {

	createDomain := resource.NewDomainCreate(data.Name.ValueString())

	var diagnostics, diags diag.Diagnostics
	createDomain.Metadata = resource.NewMetadata()
	diags = data.Labels.ElementsAs(ctx, &createDomain.Metadata.Labels, false)
	diagnostics.Append(diags...)
	diags = data.Annotations.ElementsAs(ctx, &createDomain.Metadata.Annotations, false)
	diagnostics.Append(diags...)

	createDomain.Internal = data.Internal.ValueBoolPointer()

	if !data.RouterGroup.IsNull() {
		createDomain.RouterGroup = &resource.Relationship{GUID: data.RouterGroup.ValueString()}
	}

	if !data.Org.IsNull() {
		createDomain.Relationships = &resource.DomainRelationships{
			Organization: &resource.ToOneRelationship{
				Data: &resource.Relationship{
					GUID: data.Org.ValueString(),
				},
			},
		}
	}
	if !data.SharedOrgs.IsNull() {
		var domainSharedOrgs []string
		diags = data.SharedOrgs.ElementsAs(ctx, &domainSharedOrgs, false)
		diagnostics.Append(diags...)
		createDomain.Relationships.SharedOrganizations = resource.NewToManyRelationships(domainSharedOrgs)
	}
	return *createDomain, diagnostics
}

// Sets the domain resource values for updation with cf-client from the terraform struct values.
func (plan *domainType) mapUpdateDomainTypeToValues(ctx context.Context, state domainType) (resource.DomainUpdate, diag.Diagnostics) {

	updateDomain := &resource.DomainUpdate{}

	var diagnostics diag.Diagnostics
	updateDomain.Metadata, diagnostics = setClientMetadataForUpdate(ctx, state.Labels, state.Annotations, plan.Labels, plan.Annotations)

	return *updateDomain, diagnostics
}

func mapDomainsValuesToType(ctx context.Context, domains []*resource.Domain) ([]domainType, diag.Diagnostics) {

	var diagnostics diag.Diagnostics
	domainsList := []domainType{}
	for _, domain := range domains {
		domainValue, diags := mapDomainValuesToType(ctx, domain)
		diagnostics.Append(diags...)
		domainsList = append(domainsList, domainValue)
	}

	return domainsList, diagnostics
}
