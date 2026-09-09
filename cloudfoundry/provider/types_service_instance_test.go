package provider

import (
	"context"
	"testing"
	"time"

	"github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
)

// Reproduces https://github.com/cloudfoundry/terraform-provider-cloudfoundry/issues/590:
// the Cloud Controller API omits maintenance_info (json tag "omitempty") for some managed
// service instances, and the resource mapper dereferenced it unconditionally, crashing the
// provider with a nil pointer dereference during terraform apply.
func TestMapResourceServiceInstanceValuesToType_NilMaintenanceInfo(t *testing.T) {
	managedInstance := &resource.ServiceInstance{
		Name: "tf-test-rds",
		Type: managedSerivceInstance,
		Relationships: resource.ServiceInstanceRelationships{
			Space:       &resource.ToOneRelationship{Data: &resource.Relationship{GUID: "space-guid"}},
			ServicePlan: &resource.ToOneRelationship{Data: &resource.Relationship{GUID: "plan-guid"}},
		},
		Metadata:        &resource.Metadata{},
		MaintenanceInfo: nil,
		Resource: resource.Resource{
			GUID:      "instance-guid",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	result, diags := mapResourceServiceInstanceValuesToType(context.Background(), managedInstance, jsontypes.NewNormalizedNull())

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if !result.MaintenanceInfo.IsNull() {
		t.Errorf("expected MaintenanceInfo to be null when the API omits it, got %v", result.MaintenanceInfo)
	}
}
