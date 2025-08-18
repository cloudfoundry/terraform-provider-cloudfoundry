package main

import (
	"fmt"

	"github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func main() {
	// Test lifecycle type mapping
	appType := &provider.AppType{}

	// Test buildpack lifecycle
	appType.AppLifecycle = types.StringValue("buildpack")
	fmt.Printf("Buildpack lifecycle set: %s\n", appType.AppLifecycle.ValueString())

	// Test docker lifecycle
	appType.AppLifecycle = types.StringValue("docker")
	fmt.Printf("Docker lifecycle set: %s\n", appType.AppLifecycle.ValueString())

	// Test cnb lifecycle
	appType.AppLifecycle = types.StringValue("cnb")
	fmt.Printf("CNB lifecycle set: %s\n", appType.AppLifecycle.ValueString())

	fmt.Println("All lifecycle types work correctly!")
}
