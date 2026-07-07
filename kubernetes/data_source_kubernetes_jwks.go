// Copyright IBM Corp. 2017, 2026
// SPDX-License-Identifier: MPL-2.0

package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceKubernetesJWKS() *schema.Resource {
	return &schema.Resource{
		Description: "This data source returns the cluster OpenID Connect JWKS payload from /openid/v1/jwks.",
		ReadContext: dataSourceKubernetesJWKSRead,
		Schema: map[string]*schema.Schema{
			"raw": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Raw JSON payload returned from /openid/v1/jwks.",
			},
			"keys": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeMap,
					Elem: &schema.Schema{Type: schema.TypeString},
				},
				Description: "List of JWKS keys as string maps.",
			},
		},
	}
}

func dataSourceKubernetesJWKSRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn, err := meta.(KubeClientsets).MainClientset()
	if err != nil {
		return diag.FromErr(err)
	}

	body, err := conn.Discovery().RESTClient().Get().AbsPath("/openid/v1/jwks").DoRaw(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("raw", string(body)); err != nil {
		return diag.FromErr(err)
	}

	keys, err := parseJWKSKeys(body)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("keys", keys); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("kubernetes_jwks")
	return nil
}

func parseJWKSKeys(body []byte) ([]interface{}, error) {
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("failed to decode JWKS payload: %w", err)
	}

	rawKeys, ok := payload["keys"]
	if !ok {
		return nil, nil
	}

	list, ok := rawKeys.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected JWKS keys type: %T", rawKeys)
	}

	result := make([]interface{}, 0, len(list))
	for _, item := range list {
		if m, ok := item.(map[string]interface{}); ok {
			result = append(result, flattenJWKSKey(m))
			continue
		}
		result = append(result, map[string]string{"raw": fmt.Sprintf("%v", item)})
	}

	return result, nil
}

func flattenJWKSKey(source map[string]interface{}) map[string]string {
	destination := map[string]string{}
	for k, v := range source {
		switch typed := v.(type) {
		case string:
			destination[k] = typed
		case float64, bool, int, int32, int64:
			destination[k] = fmt.Sprintf("%v", typed)
		case []interface{}, map[string]interface{}:
			encoded, err := json.Marshal(typed)
			if err != nil {
				destination[k] = fmt.Sprintf("%v", typed)
			} else {
				destination[k] = string(encoded)
			}
		default:
			destination[k] = fmt.Sprintf("%v", typed)
		}
	}
	return destination
}
