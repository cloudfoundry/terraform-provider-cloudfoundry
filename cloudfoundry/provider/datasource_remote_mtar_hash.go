package provider

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"slices"

	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider/managers"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &RemoteMtarHashDataSource{}
	_ datasource.DataSourceWithConfigure = &RemoteMtarHashDataSource{}
)

// Instantiates a remote file hash data source.
func NewRemoteMtarHashDataSource() datasource.DataSource {
	return &RemoteMtarHashDataSource{}
}

// Contains reference to the mta client to be used for making the API calls.
type RemoteMtarHashDataSource struct {
	cfClient *client.Client
}

func (d *RemoteMtarHashDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_remote_mtar_hash"
}

func (d *RemoteMtarHashDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	session, _ := req.ProviderData.(*managers.Session)
	d.cfClient = session.CFClient
}

func (d *RemoteMtarHashDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Gets the SHA-256 sum of a MTAR file hosted on a remote URL.",

		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				MarkdownDescription: "The URL where the file is hosted",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The SHA sum of the file",
				Computed:            true,
			},
		},
	}
}

func (d *RemoteMtarHashDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RemoteFileHashDataSourceType
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := data.Url.ValueString()

	urlCheckReq, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create HEAD request",
			fmt.Sprintf("Error: %v", err),
		)
		return
	}
	urlCheckResp, err := d.cfClient.ExecuteRequest(urlCheckReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid URL",
			fmt.Sprintf("Could not reach URL: %v", err),
		)
		return
	}
	defer urlCheckResp.Body.Close()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create GET request",
			fmt.Sprintf("Error: %v", err),
		)
		return
	}
	httpResp, err := d.cfClient.ExecuteRequest(httpReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to download file",
			fmt.Sprintf("Error: %v", err),
		)
		return
	}
	defer httpResp.Body.Close()

	allowedContentTypes := []string{
		"application/octet-stream",
	}

	contentType := httpResp.Header.Get("Content-Type")
	if !slices.Contains(allowedContentTypes, contentType) {
		resp.Diagnostics.AddError(
			"Invalid content type",
			fmt.Sprintf("The content-type of the response should be a file type. Received %s", contentType),
		)
		return
	}

	hash := sha256.New()
	if _, err := io.Copy(hash, httpResp.Body); err != nil {
		resp.Diagnostics.AddError(
			"Failed to calculate checksum",
			fmt.Sprintf("Error: %v", err),
		)
		return
	}

	data.Id = types.StringValue(hex.EncodeToString(hash.Sum(nil)))

	tflog.Trace(ctx, "read a remote file hash datasource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
