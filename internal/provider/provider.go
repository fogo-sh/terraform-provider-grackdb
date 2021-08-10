package provider

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func init() {
	// Set descriptions to support markdown syntax, this will be used in document generation
	// and the language server.
	schema.DescriptionKind = schema.StringMarkdown

	// Customize the content of descriptions when output. For example you can add defaults on
	// to the exported descriptions if present.
	// schema.SchemaDescriptionBuilder = func(s *schema.Schema) string {
	// 	desc := s.Description
	// 	if s.Default != nil {
	// 		desc += fmt.Sprintf(" Defaults to `%v`.", s.Default)
	// 	}
	// 	return strings.TrimSpace(desc)
	// }
}

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"api_url": {
					Type:     schema.TypeString,
					Optional: true,
					Default:  "https://grackdb.fogo.sh/query",
				},
				"token": {
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					DefaultFunc: schema.EnvDefaultFunc("GRACKDB_TOKEN", nil),
				},
			},
			DataSourcesMap: map[string]*schema.Resource{
				"grackdb_current_user": dataSourceCurrentUser(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"scaffolding_resource": resourceScaffolding(),
			},
		}

		p.ConfigureContextFunc = configure(version, p)

		return p
	}
}

type apiClient struct {
	httpClient *http.Client
	apiUrl     string
}

type withHeaderType struct {
	http.Header
	rt http.RoundTripper
}

func withHeader(rt http.RoundTripper) withHeaderType {
	if rt == nil {
		rt = http.DefaultTransport
	}

	return withHeaderType{Header: make(http.Header), rt: rt}
}

func (h withHeaderType) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range h.Header {
		req.Header[k] = v
	}

	return h.rt.RoundTrip(req)
}

func configure(version string, p *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		apiUrl := d.Get("api_url").(string)
		token := d.Get("token").(string)

		userAgent := p.UserAgent("terraform-provider-grackdb", version)
		httpClient := http.DefaultClient
		transport := withHeader(httpClient.Transport)
		transport.Set("User-Agent", userAgent)

		if token != "" {
			transport.Set("Authorization", "Bearer "+token)
		}

		httpClient.Transport = transport
		return &apiClient{
			httpClient: httpClient,
			apiUrl:     apiUrl,
		}, nil
	}
}
