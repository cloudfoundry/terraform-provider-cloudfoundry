package managers

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cloudfoundry/go-cfclient/v3/client"
	config "github.com/cloudfoundry/go-cfclient/v3/config"
	"github.com/cloudfoundry/terraform-provider-cloudfoundry/internal/version"
	"github.com/hashicorp/terraform-plugin-framework/provider"
)

type CloudFoundryProviderConfig struct {
	Endpoint          string
	User              string
	Password          string
	CFClientID        string
	CFClientSecret    string
	SkipSslValidation bool
	Origin            string
	AccessToken       string
	RefreshToken      string
	AssertionToken    string
}

type Session struct {
	CFClient *client.Client
}

func (c *CloudFoundryProviderConfig) NewSession(httpClient *http.Client, req provider.ConfigureRequest) (*Session, error) {
	var cfg *config.Config
	var err error
	var opts []config.Option
	var finalAgent string

	// Setting a higher value than default of 30s as file uploads seem to fail
	opts = append(opts, config.RequestTimeout(10*time.Minute))

	cfUserAgent := os.Getenv("CF_APPEND_USER_AGENT")
	if len(strings.TrimSpace(cfUserAgent)) == 0 {
		finalAgent = fmt.Sprintf("Terraform/%s terraform-provider-cloudfoundry/%s", req.TerraformVersion, version.ProviderVersion)
	} else {
		finalAgent = fmt.Sprintf("Terraform/%s terraform-provider-cloudfoundry/%s %s", req.TerraformVersion, version.ProviderVersion, cfUserAgent)
	}
	opts = append(opts, config.UserAgent(finalAgent))

	if httpClient != nil {
		opts = append(opts, config.HttpClient(httpClient))
	}
	if c.SkipSslValidation {
		opts = append(opts, config.SkipTLSValidation())
	}
	if c.Origin != "" {
		opts = append(opts, config.Origin(c.Origin))
	}
	switch {
	case c.User != "" && c.Password != "":
		opts = append(opts, config.UserPassword(c.User, c.Password))
		cfg, err = config.New(c.Endpoint, opts...)
	case c.CFClientID != "" && c.CFClientSecret != "":
		opts = append(opts, config.ClientCredentials(c.CFClientID, c.CFClientSecret))
		cfg, err = config.New(c.Endpoint, opts...)
	case c.AccessToken != "":
		opts = append(opts, config.Token(c.AccessToken, c.RefreshToken))
		cfg, err = config.New(c.Endpoint, opts...)
	case c.AssertionToken != "":
		opts = append(opts, config.JWTBearerAssertion(c.AssertionToken))
		cfg, err = config.New(c.Endpoint, opts...)
	default:
		cfg, err = config.NewFromCFHome(opts...)
	}
	if err != nil {
		return nil, err
	}
	cf, err := client.New(cfg)
	if err != nil {
		return nil, err
	}
	s := Session{
		CFClient: cf,
	}
	return &s, nil
}
