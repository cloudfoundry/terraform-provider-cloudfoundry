package provider

import (
	"fmt"
	"strconv"
	"strings"

	"code.cloudfoundry.org/policy_client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type networkPoliciesType struct {
	Id       types.String         `tfsdk:"id"`
	Policies networkPoliciesSlice `tfsdk:"policies"`
}

type networkPoliciesSlice []networkPolicyType

type networkPolicyType struct {
	SourceApp      types.String `tfsdk:"source_app"`
	DestinationApp types.String `tfsdk:"destination_app"`
	Port           types.String `tfsdk:"port"`
	Protocol       types.String `tfsdk:"protocol"`
}

func (data *networkPoliciesType) mapToPolicyClientPolicies() ([]policy_client.Policy, diag.Diagnostics) {
	return data.Policies.mapToPolicyClientPolicies()
}

func (policies networkPoliciesSlice) mapToPolicyClientPolicies() ([]policy_client.Policy, diag.Diagnostics) {
	var diags diag.Diagnostics
	var mapped []policy_client.Policy

	for _, p := range policies {
		start, end, err := portRangeParse(p.Port.ValueString())
		if err != nil {
			diags.AddError("Error parsing port range", err.Error())
			return nil, diags
		}
		mapped = append(mapped, policy_client.Policy{
			Source: policy_client.Source{
				ID: p.SourceApp.ValueString(),
			},
			Destination: policy_client.Destination{
				ID:       p.DestinationApp.ValueString(),
				Protocol: p.Protocol.ValueString(),
				Ports: policy_client.Ports{
					Start: start,
					End:   end,
				},
			},
		})
	}
	return mapped, diags
}

func portRangeParse(portRange string) (start int, end int, err error) {
	portRangeSplit := strings.Split(portRange, "-")
	if len(portRangeSplit) > 2 {
		return 0, 0, fmt.Errorf("invalid range")
	}
	start, err = strconv.Atoi(portRangeSplit[0])
	if err != nil {
		return 0, 0, err
	}
	if len(portRangeSplit) == 1 {
		return start, start, nil
	}
	end, err = strconv.Atoi(portRangeSplit[1])
	if err != nil {
		return 0, 0, err
	}
	return start, end, nil
}

func mapPolicyClientPoliciesToNetworkPoliciesSlice(policies []policy_client.Policy) networkPoliciesSlice {
	var mapped networkPoliciesSlice

	for _, p := range policies {
		port := strconv.Itoa(p.Destination.Ports.Start)
		if p.Destination.Ports.Start != p.Destination.Ports.End {
			port = fmt.Sprintf("%d-%d", p.Destination.Ports.Start, p.Destination.Ports.End)
		}
		mapped = append(mapped, networkPolicyType{
			SourceApp:      types.StringValue(p.Source.ID),
			DestinationApp: types.StringValue(p.Destination.ID),
			Protocol:       types.StringValue(p.Destination.Protocol),
			Port:           types.StringValue(port),
		})
	}
	return mapped
}
