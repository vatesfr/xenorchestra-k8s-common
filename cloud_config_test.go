/*
Copyright 2023 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package xok8scommon_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	xok8s "github.com/vatesfr/xenorchestra-k8s-common"
)

func TestReadCloudConfig(t *testing.T) {
	cfg, err := xok8s.ReadCloudConfig(nil)
	assert.NotNil(t, err)
	assert.NotNil(t, cfg)

	// Empty config
	cfg, err = xok8s.ReadCloudConfig(strings.NewReader(`
`))
	assert.NotNil(t, err)
	assert.NotNil(t, cfg)

	// Wrong config
	cfg, err = xok8s.ReadCloudConfig(strings.NewReader(`
test: false
`))

	assert.NotNil(t, err)
	assert.NotNil(t, cfg)

	// Non full config
	cfg, err = xok8s.ReadCloudConfig(strings.NewReader(`
url: abcd
token: 123ABC
`))

	assert.NotNil(t, err)
	assert.NotNil(t, cfg)

	// Valid config with one cluster
	cfg, err = xok8s.ReadCloudConfig(strings.NewReader(`
url: https://example.com
insecure: false
token: "123ABC"
`))
	assert.Nil(t, err)
	assert.NotNil(t, cfg)

	// Valid config with one cluster (username/password)
	cfg, err = xok8s.ReadCloudConfig(strings.NewReader(`
url: https://example.com
insecure: false
username: "user@pam"
password: "secret"
`))
	assert.Nil(t, err)
	assert.NotNil(t, cfg)
}

func TestReadCloudConfigFromFile(t *testing.T) {
	cfg, err := xok8s.ReadCloudConfigFromFile("testdata/cloud-config.yaml")
	assert.NotNil(t, err)
	assert.EqualError(t, err,
		"error reading testdata/cloud-config.yaml: open testdata/cloud-config.yaml: no such file or directory")
	assert.NotNil(t, cfg)

	cfg, err = xok8s.ReadCloudConfigFromFile("./hack/xo-config.yaml")
	assert.Nil(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "https://xoa.example.com", cfg.URL)
	assert.Equal(t, "123ABC", cfg.Token)
	assert.Equal(t, false, cfg.Insecure)
	assert.Equal(t, 60*time.Second, cfg.ClientTimeout)
}

func TestLoadXOConfigFromEnv(t *testing.T) {
	// Save original environment
	originalURL := os.Getenv("XOA_URL")
	originalToken := os.Getenv("XOA_TOKEN")
	originalUser := os.Getenv("XOA_USER")
	originalPassword := os.Getenv("XOA_PASSWORD")
	originalInsecure := os.Getenv("XOA_INSECURE")
	originalClientTimeout := os.Getenv("XOA_CLIENT_TIMEOUT")

	// Clean environment for testing
	t.Setenv("XOA_URL", "")
	t.Setenv("XOA_TOKEN", "")
	t.Setenv("XOA_USER", "")
	t.Setenv("XOA_PASSWORD", "")
	t.Setenv("XOA_INSECURE", "")
	t.Setenv("XOA_CLIENT_TIMEOUT", "")

	// Test with missing required environment variables
	// Should fail because authentication is missing
	cfg, err := xok8s.LoadXOConfigFromEnv()
	assert.NotNil(t, err)
	assert.EqualError(t, err, "either token or username/password are required for authentication")
	assert.Equal(t, xok8s.XoConfig{}, cfg)

	// Test with authentication but missing URL
	t.Setenv("XOA_TOKEN", "test-token")
	_, err = xok8s.LoadXOConfigFromEnv()
	assert.NotNil(t, err)
	assert.EqualError(t, err, "url is required")

	// Test with valid token authentication
	t.Setenv("XOA_URL", "https://example.com")
	cfg, err = xok8s.LoadXOConfigFromEnv()
	assert.Nil(t, err)
	assert.Equal(t, "https://example.com", cfg.URL)
	assert.Equal(t, "test-token", cfg.Token)
	assert.False(t, cfg.Insecure)

	// Test with username/password authentication
	t.Setenv("XOA_TOKEN", "") // Clear token
	t.Setenv("XOA_USER", "test-user")
	t.Setenv("XOA_PASSWORD", "test-password")
	cfg, err = xok8s.LoadXOConfigFromEnv()
	assert.Nil(t, err)
	assert.Equal(t, "https://example.com", cfg.URL)
	assert.Equal(t, "test-user", cfg.Username)
	assert.Equal(t, "test-password", cfg.Password)
	assert.False(t, cfg.Insecure)

	// Test with insecure flag
	t.Setenv("XOA_INSECURE", "true")
	cfg, err = xok8s.LoadXOConfigFromEnv()
	assert.Nil(t, err)
	assert.True(t, cfg.Insecure)

	// Test with invalid insecure flag
	t.Setenv("XOA_INSECURE", "invalid")
	cfg, err = xok8s.LoadXOConfigFromEnv()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid XOA_INSECURE value")

	// Test with invalid client timeout
	t.Setenv("XOA_INSECURE", "false")
	t.Setenv("XOA_CLIENT_TIMEOUT", "invalid")
	cfg, err = xok8s.LoadXOConfigFromEnv()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid XOA_CLIENT_TIMEOUT value")

	// Test with valid client timeout
	t.Setenv("XOA_INSECURE", "false")
	t.Setenv("XOA_CLIENT_TIMEOUT", "15s")
	cfg, err = xok8s.LoadXOConfigFromEnv()
	assert.Nil(t, err)
	assert.Equal(t, 15*time.Second, cfg.ClientTimeout)

	// Restore original environment
	t.Setenv("XOA_URL", originalURL)
	t.Setenv("XOA_TOKEN", originalToken)
	t.Setenv("XOA_USER", originalUser)
	t.Setenv("XOA_PASSWORD", originalPassword)
	t.Setenv("XOA_INSECURE", originalInsecure)
	t.Setenv("XOA_CLIENT_TIMEOUT", originalClientTimeout)
}
