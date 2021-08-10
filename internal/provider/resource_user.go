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

func resourceUser() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Create and manage a GrackDB User.",

		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "Unique ID for this user.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"username": {
				Description: "Username for this user.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"avatar_url": {
				Description: "URL to this user's avatar.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

type createUserResp struct {
	Data struct {
		CreateUser types.User `json:"createUser"`
	} `json:"data"`
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)

	variables := map[string]interface{}{
		"username": d.Get("username").(string),
	}

	avatarUrl := d.Get("avatar_url").(string)
	if avatarUrl != "" {
		variables["avatarUrl"] = avatarUrl
	}

	reqBody, err := json.Marshal(map[string]interface{}{
		"operationName": nil,
		"query": `
			mutation($input: CreateUserInput!) {
				createUser(input: $input) {
			  		id
			  		username
			  		avatarUrl
				}
		  	}
		`,
		"variables": map[string]interface{}{
			"input": variables,
		},
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

	respData := new(createUserResp)
	err = json.Unmarshal(body, &respData)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(respData.Data.CreateUser.ID)

	return resourceUserRead(ctx, d, meta)
}

type readUserResp struct {
	Data struct {
		Users struct {
			Edges []struct {
				Node types.User
			} `json:"edges"`
		} `json:"users"`
	} `json:"data"`
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)

	reqBody, err := json.Marshal(map[string]interface{}{
		"operationName": nil,
		"query": `
			query($userId: ID!) {
				users(where: { id: $userId }) {
			  		edges {
						node {
				  			id
				  			username
						}
			  		}
				}
			}
		`,
		"variables": map[string]interface{}{
			"userId": d.Id(),
		},
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

	respData := new(readUserResp)
	err = json.Unmarshal(body, &respData)
	if err != nil {
		return diag.FromErr(err)
	}

	if len(respData.Data.Users.Edges) == 0 {
		return diag.Diagnostics{
			diag.Diagnostic{
				Summary: "Unable to refresh user state, unable to find requested user.",
			},
		}
	}

	user := respData.Data.Users.Edges[0].Node
	if err = d.Set("id", user.ID); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("username", user.Username); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("avatar_url", user.AvatarURL); err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}

type updateUserResp struct {
	Data struct {
		UpdateUser types.User `json:"updateUser"`
	} `json:"data"`
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)

	variables := map[string]interface{}{}

	if d.HasChange("username") {
		variables["username"] = d.Get("username").(string)
	}
	if d.HasChange("avatar_url") {
		avatarUrlVal := d.Get("avatar_url").(string)
		avatarUrl := &avatarUrlVal
		if avatarUrlVal == "" {
			avatarUrl = nil
		}
		variables["avatarUrl"] = avatarUrl
	}

	reqBody, err := json.Marshal(map[string]interface{}{
		"operationName": nil,
		"query": `
			mutation($userId: ID!, $input: UpdateUserInput!) {
				updateUser(id: $userId, input: $input) {
			  		id
			  		username
			  		avatarUrl
				}
		  	}
		`,
		"variables": map[string]interface{}{
			"userId": d.Id(),
			"input":  variables,
		},
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

	respData := new(updateUserResp)
	err = json.Unmarshal(body, &respData)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceUserRead(ctx, d, meta)
}

type deleteUserResp struct {
	Data struct {
		DeleteUser types.User `json:"deleteUser"`
	} `json:"data"`
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)

	reqBody, err := json.Marshal(map[string]interface{}{
		"operationName": nil,
		"query": `
			mutation($userId: ID!) {
				deleteUser(id: $userId) {
					id
					username
					avatarUrl
				}
			}
		`,
		"variables": map[string]interface{}{
			"userId": d.Id(),
		},
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

	respData := new(deleteUserResp)
	err = json.Unmarshal(body, &respData)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diag.Diagnostics{}
}
