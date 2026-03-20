/*
Copyright 2025 Vates

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

// Package xok8scommon provides shared Kubernetes integration helpers for Xen Orchestra
// components (CSI driver, Cloud Controller Manager, etc.).
package xok8scommon

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

// XoConfig holds the Xen Orchestra connection configuration.
type XoConfig struct {
	URL      string `yaml:"url"`
	Insecure bool   `yaml:"insecure,omitempty"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
	Token    string `yaml:"token,omitempty"`
}

func validateXOConfig(cfg XoConfig) (XoConfig, error) {

	if cfg.Username != "" && cfg.Password != "" {
		if cfg.Token != "" {
			return XoConfig{}, fmt.Errorf("token is not allowed when username and password are set")
		}
	} else if cfg.Token == "" {
		return XoConfig{}, fmt.Errorf("either token or username/password are required for authentication")
	}

	if cfg.URL == "" || !strings.HasPrefix(cfg.URL, "http") {
		return XoConfig{}, fmt.Errorf("url is required")
	}

	return cfg, nil
}

// ReadCloudConfig reads and validates the XO cloud config from a reader.
func ReadCloudConfig(config io.Reader) (XoConfig, error) {
	cfg := XoConfig{}

	if config != nil {
		if err := yaml.NewDecoder(config).Decode(&cfg); err != nil {
			return XoConfig{}, err
		}
	}

	return validateXOConfig(cfg)
}

// ReadCloudConfigFromFile reads and validates the XO cloud config from a file path.
func ReadCloudConfigFromFile(file string) (XoConfig, error) {
	f, err := os.Open(filepath.Clean(file))
	if err != nil {
		return XoConfig{}, fmt.Errorf("error reading %s: %v", file, err)
	}
	defer f.Close() // nolint: errcheck

	return ReadCloudConfig(f)
}

// LoadXOConfigFromEnv loads Xen Orchestra configuration from environment variables.
// It uses the same environment variables as the Xen Orchestra Go SDK:
// - XOA_URL: the base URL of the Xen Orchestra API (required)
// - XOA_TOKEN: the authentication token (optional if username/password provided)
// - XOA_USER: the username (optional if token provided)
// - XOA_PASSWORD: the password (optional if token provided)
// - XOA_INSECURE: whether to skip TLS verification (optional, defaults to false)
func LoadXOConfigFromEnv() (XoConfig, error) {
	url := os.Getenv("XOA_URL")
	token := os.Getenv("XOA_TOKEN")
	username := os.Getenv("XOA_USER")
	password := os.Getenv("XOA_PASSWORD")
	insecureStr := os.Getenv("XOA_INSECURE")

	cfg := XoConfig{
		URL:      url,
		Token:    token,
		Username: username,
		Password: password,
	}

	// Parse insecure flag
	if insecureStr != "" {
		var err error
		cfg.Insecure, err = strconv.ParseBool(insecureStr)
		if err != nil {
			return XoConfig{}, fmt.Errorf("invalid XOA_INSECURE value: %v", err)
		}
	}

	return validateXOConfig(cfg)

}
