package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/stretchr/testify/assert"
)

func TestProcessSchemaAttributes_NonStandardProcessType(t *testing.T) {
	// Create an instance of the appResource
	r := &appResource{}

	// Get the schema attributes
	attributes := r.ProcessSchemaAttributes()

	// Check that the "type" attribute exists
	typeAttr, exists := attributes["type"]
	assert.True(t, exists, "type attribute should exist")

	// Check that it's a StringAttribute
	stringAttr, ok := typeAttr.(schema.StringAttribute)
	assert.True(t, ok, "type attribute should be a StringAttribute")

	// Check that it's required
	assert.True(t, stringAttr.Required, "type attribute should be required")

	// Check that it has no validators (was previously restricted to web/worker)
	assert.Empty(t, stringAttr.Validators, "type attribute should not have validators restricting to web/worker")

	// Check the description contains information about accepting any string
	assert.Contains(t, stringAttr.MarkdownDescription, "Any string identifier",
		"description should indicate any string identifier is accepted")
}
