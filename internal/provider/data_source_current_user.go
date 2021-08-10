package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"

	"github.com/fogo-sh/terraform-provider-grackdb/internal/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCurrentUser() *schema.Resource {
	return &schema.Resource{
		Description: "Get details on the current authenticated user.",

		ReadContext: dataSourceCurrentUserRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "Unique ID for this user.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"username": {
				Description: "Unique username for this user.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"avatar_url": {
				Description: "URL for this user's avatar.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

type currentUserResp struct {
	Data struct {
		CurrentUser *types.User `json:"currentUser"`
	} `json:"data"`
}

func dataSourceCurrentUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*apiClient)

	reqBody, err := json.Marshal(map[string]interface{}{
		"operationName": nil,
		"query": `
			{
				currentUser {
					id
					username
					avatarUrl
				}
			}
		`,
		"variables": map[string]interface{}{},
	})
	if err != nil {
		return diag.FromErr(err)
	}

	resp, err := client.httpClient.Post(
		client.apiUrl,
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return diag.FromErr(err)
	}

	respData := new(currentUserResp)
	err = json.Unmarshal(body, &respData)
	if err != nil {
		return diag.FromErr(err)
	}

	if respData.Data.CurrentUser == nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Summary: "Failed to retrieve current user. Please ensure you've provided a valid api token.",
			},
		}
	}

	d.SetId(respData.Data.CurrentUser.ID)

	if err = d.Set("id", respData.Data.CurrentUser.ID); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("username", respData.Data.CurrentUser.Username); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("avatar_url", respData.Data.CurrentUser.AvatarURL); err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}
