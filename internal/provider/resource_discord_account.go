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

func resourceDiscordAccount() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Create and manage a GrackDB Discord account.",

		CreateContext: resourceDiscordAccountCreate,
		ReadContext:   resourceDiscordAccountRead,
		UpdateContext: resourceDiscordAccountUpdate,
		DeleteContext: resourceDiscordAccountDelete,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "Unique ID for this Discord account.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"discord_id": {
				Description: "Discord snowflake for this account.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"username": {
				Description: "Username for this account.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"discriminator": {
				Description: "Discriminator for this account.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"owner": {
				Description: "ID of the User that owns this account.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"bot": {
				Description: "ID of the bot that owns this account.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

type createDiscordAccountResp struct {
	Data struct {
		CreateDiscordAccount types.DiscordAccount `json:"createDiscordAccount"`
	} `json:"data"`
}

func resourceDiscordAccountCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)

	variables := map[string]interface{}{
		"discordId":     d.Get("discord_id").(string),
		"username":      d.Get("username").(string),
		"discriminator": d.Get("discriminator").(string),
	}

	owner := d.Get("owner").(string)
	if owner != "" {
		variables["owner"] = owner
	}

	reqBody, err := json.Marshal(map[string]interface{}{
		"operationName": nil,
		"query": `
			mutation($input: CreateDiscordAccountInput!) {
				createDiscordAccount(input: $input) {
					id
					discordId
					username
					discriminator
					owner {
						id
						username
						avatarUrl
					}
					bot {
						id
					}
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

	respData := new(createDiscordAccountResp)
	err = json.Unmarshal(body, &respData)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(respData.Data.CreateDiscordAccount.ID)

	return resourceDiscordAccountRead(ctx, d, meta)
}

type readDiscordAccountResp struct {
	Data struct {
		DiscordAccounts struct {
			Edges []struct {
				Node types.DiscordAccount
			} `json:"edges"`
		} `json:"DiscordAccounts"`
	} `json:"data"`
}

func resourceDiscordAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)

	reqBody, err := json.Marshal(map[string]interface{}{
		"operationName": nil,
		"query": `
			query($accountId: ID!) {
				discordAccounts(where: { id: $accountId }) {
					edges {
						node {
							id
							discordId
							username
							discriminator
							owner {
								id
								username
								avatarUrl
							}
							bot {
								id
							}
						}
					}
				}
			}		  
		`,
		"variables": map[string]interface{}{
			"accountId": d.Id(),
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

	respData := new(readDiscordAccountResp)
	err = json.Unmarshal(body, &respData)
	if err != nil {
		return diag.FromErr(err)
	}

	if len(respData.Data.DiscordAccounts.Edges) == 0 {
		return diag.Diagnostics{
			diag.Diagnostic{
				Summary: "Unable to refresh discord account state, unable to find requested account.",
			},
		}
	}

	account := respData.Data.DiscordAccounts.Edges[0].Node
	if err = d.Set("id", account.ID); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("discord_id", account.DiscordID); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("username", account.Username); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("discriminator", account.Discriminator); err != nil {
		return diag.FromErr(err)
	}

	if account.Owner != nil {
		if err = d.Set("owner", account.Owner.ID); err != nil {
			return diag.FromErr(err)
		}
	}

	if account.Bot != nil {
		if err = d.Set("bot", account.Bot.ID); err != nil {
			return diag.FromErr(err)
		}
	}

	return diag.Diagnostics{}
}

type updateDiscordAccountResp struct {
	Data struct {
		UpdateDiscordAccount types.DiscordAccount `json:"updateDiscordAccount"`
	} `json:"data"`
}

func resourceDiscordAccountUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)

	variables := map[string]interface{}{}

	if d.HasChange("username") {
		variables["username"] = d.Get("username").(string)
	}
	if d.HasChange("discriminator") {
		variables["discriminator"] = d.Get("discriminator").(string)
	}
	// TODO: Fix unsetting owner
	if d.HasChange("owner") {
		variables["owner"] = d.Get("owner").(string)
	}

	reqBody, err := json.Marshal(map[string]interface{}{
		"operationName": nil,
		"query": `
			mutation($accountId: ID!, $input: UpdateDiscordAccountInput!) {
				updateDiscordAccount(id: $accountId, input: $input) {
					id
					discordId
					username
					discriminator
					owner {
						id
						username
						avatarUrl
					}
					bot {
						id
					}
				}
		  	}
		`,
		"variables": map[string]interface{}{
			"accountId": d.Id(),
			"input":     variables,
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

	respData := new(updateDiscordAccountResp)
	err = json.Unmarshal(body, &respData)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceDiscordAccountRead(ctx, d, meta)
}

type deleteDiscordAccountResp struct {
	Data struct {
		DeleteDiscordAccount types.DiscordAccount `json:"deleteDiscordAccount"`
	} `json:"data"`
}

func resourceDiscordAccountDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)

	reqBody, err := json.Marshal(map[string]interface{}{
		"operationName": nil,
		"query": `
			mutation($accountId: ID!) {
				deleteDiscordAccount(id: $accountId) {
					id
					discordId
					username
					discriminator
					owner {
						id
						username
						avatarUrl
					}
					bot {
						id
					}
				}
			}
		`,
		"variables": map[string]interface{}{
			"accountId": d.Id(),
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

	respData := new(deleteDiscordAccountResp)
	err = json.Unmarshal(body, &respData)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diag.Diagnostics{}
}
