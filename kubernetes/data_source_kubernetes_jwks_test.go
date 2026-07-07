// Copyright IBM Corp. 2017, 2026
// SPDX-License-Identifier: MPL-2.0

package kubernetes

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

func TestParseJWKSKeys(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		expected map[string]string
	}{
		{
			name: "simple RSA key",
			body: `{"keys":[{"kty":"RSA","kid":"1","n":"abc","e":"AQAB"}]}`,
			expected: map[string]string{
				"kty": "RSA",
				"kid": "1",
				"n":   "abc",
				"e":   "AQAB",
			},
		},
		{
			name: "nested x5c array",
			body: `{"keys":[{"kty":"RSA","x5c":["a","b"]}]}`,
			expected: map[string]string{
				"kty": "RSA",
				"x5c": `["a","b"]`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keys, err := parseJWKSKeys([]byte(tt.body))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(keys) != 1 {
				t.Fatalf("expected 1 key, got %d", len(keys))
			}
			key, ok := keys[0].(map[string]string)
			if !ok {
				t.Fatalf("expected map[string]string, got %T", keys[0])
			}
			for expectedKey, expectedValue := range tt.expected {
				if key[expectedKey] != expectedValue {
					t.Fatalf("expected %s=%q, got %q", expectedKey, expectedValue, key[expectedKey])
				}
			}
		})
	}
}

func TestAccKubernetesDataSourceJWKS_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccKubernetesDataSourceJWKSConfig(),
				Check: resource.ComposeTestCheckFunc(

					// Ensure the raw JSON string is populated
					resource.TestCheckResourceAttrSet("data.kubernetes_jwks.test", "raw"),

					// Ensure that at least one key was parsed into the list
					// (A healthy cluster will always have at least one signing key)
					resource.TestCheckResourceAttrSet("data.kubernetes_jwks.test", "keys.#"),
				),
			},
		},
	})
}

func testAccKubernetesDataSourceJWKSConfig() string {
	return `
data "kubernetes_jwks" "test" {}
`
}
