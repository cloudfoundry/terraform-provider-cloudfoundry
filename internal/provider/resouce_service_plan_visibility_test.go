package provider

import (
	"context"
	"testing"

	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/provider/managers"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestServicePlanVisibilityResource(t *testing.T) {
	mockClient := &client.Client{}
	session := &managers.Session{CFClient: mockClient}
	resource := NewServicePlanVisibilityResource().(*servicePlanVisibilityResource)
	resource.Configure(context.Background(), resource.ConfigureRequest{
		ProviderData: session,
	}, &resource.ConfigureResponse{})

	t.Run("Create", func(t *testing.T) {
		// Mock the Create API response
		mockClient.ServicePlansVisibility.ApplyFunc = func(ctx context.Context, guid string, visibility *client.ServicePlanVisibility) (*client.Job, error) {
			return &client.Job{GUID: "job-guid"}, nil
		}
		mockClient.ServicePlanVisibilities.GetFunc = func(ctx context.Context, guid string) (*client.ServicePlanVisibility, error) {
			return &client.ServicePlanVisibility{GUID: "visibility-guid"}, nil
		}

		req := resource.CreateRequest{
			Plan: types.ObjectValue(map[string]types.Value{
				"service_plan_guid": types.StringValue("service-plan-guid"),
				"organization_guid": types.StringValue("organization-guid"),
			}),
		}
		resp := &resource.CreateResponse{}
		resource.Create(context.Background(), req, resp)

		assert.False(t, resp.Diagnostics.HasError())
	})

	t.Run("Read", func(t *testing.T) {
		// Mock the Read API response
		mockClient.ServicePlanVisibilities.GetFunc = func(ctx context.Context, guid string) (*client.ServicePlanVisibility, error) {
			return &client.ServicePlanVisibility{GUID: "visibility-guid", ServicePlanGUID: "service-plan-guid", OrganizationGUID: "organization-guid"}, nil
		}

		req := resource.ReadRequest{
			State: types.ObjectValue(map[string]types.Value{
				"service_plan_guid": types.StringValue("service-plan-guid"),
			}),
		}
		resp := &resource.ReadResponse{}
		resource.Read(context.Background(), req, resp)

		assert.False(t, resp.Diagnostics.HasError())
	})

	t.Run("Update", func(t *testing.T) {
		// Mock the Update API response
		mockClient.ServicePlanVisibilities.UpdateFunc = func(ctx context.Context, guid string, visibility *client.ServicePlanVisibility) (*client.ServicePlanVisibility, error) {
			return &client.ServicePlanVisibility{GUID: "visibility-guid"}, nil
		}

		req := resource.UpdateRequest{
			Plan: types.ObjectValue(map[string]types.Value{
				"service_plan_guid": types.StringValue("service-plan-guid"),
				"organization_guid": types.StringValue("organization-guid"),
			}),
			State: types.ObjectValue(map[string]types.Value{
				"service_plan_guid": types.StringValue("service-plan-guid"),
			}),
		}
		resp := &resource.UpdateResponse{}
		resource.Update(context.Background(), req, resp)

		assert.False(t, resp.Diagnostics.HasError())
	})

	t.Run("Delete", func(t *testing.T) {
		// Mock the Delete API response
		mockClient.ServicePlanVisibilities.DeleteFunc = func(ctx context.Context, guid string) error {
			return nil
		}

		req := resource.DeleteRequest{
			State: types.ObjectValue(map[string]types.Value{
				"service_plan_guid": types.StringValue("service-plan-guid"),
			}),
		}
		resp := &resource.DeleteResponse{}
		resource.Delete(context.Background(), req, resp)

		assert.False(t, resp.Diagnostics.HasError())
	})
}
