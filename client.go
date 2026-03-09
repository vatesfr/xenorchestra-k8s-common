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

package xok8scommon

import (
	"context"
	"fmt"

	"github.com/gofrs/uuid"

	xocfg "github.com/vatesfr/xenorchestra-go-sdk/pkg/config"
	"github.com/vatesfr/xenorchestra-go-sdk/pkg/payloads"
	"github.com/vatesfr/xenorchestra-go-sdk/pkg/services/library"
	v2 "github.com/vatesfr/xenorchestra-go-sdk/v2"

	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

// XoClient wraps the XO SDK library with Kubernetes-aware helpers.
type XoClient struct {
	config *XoConfig
	Client library.Library
}

// NewXOClient creates a new XoClient from an XoConfig.
// The http/https URL scheme is automatically converted to ws/wss for the XO WebSocket API.
func NewXOClient(cfg *XoConfig) (*XoClient, error) {
	url := cfg.URL
	if len(url) >= 5 {
		if url[:5] == "https" {
			url = "wss" + url[5:]
		} else if url[:4] == "http" {
			url = "ws" + url[4:]
		}
	}

	xoConfig := &xocfg.Config{
		Url:                url,
		Username:           cfg.Username,
		Password:           cfg.Password,
		Token:              cfg.Token,
		InsecureSkipVerify: cfg.Insecure,
	}

	client, err := v2.New(xoConfig)
	if err != nil {
		klog.Errorf("Failed to create Xen Orchestra client: %v", err)
		return nil, err
	}

	return &XoClient{
		config: cfg,
		Client: client,
	}, nil
}

// CheckClient verifies the XO connection is working by listing VMs.
func (c *XoClient) CheckClient(ctx context.Context) error {
	vms, err := c.Client.VM().GetAll(ctx, 1, "")
	if err != nil {
		return fmt.Errorf("failed to get list of VMs: %v", err)
	}
	if len(vms) > 0 {
		klog.V(4).InfoS("Xen Orchestra instance has VMs", "count", len(vms))
	} else {
		klog.InfoS("Xen Orchestra instance has no VMs, or check the account permission")
	}
	return nil
}

// FindVMByNode finds the XO VM corresponding to a Kubernetes node.
// It returns the VM, its pool UUID, and any error.
// The node's SystemUUID is used first; if empty, the ProviderID is parsed instead.
// A little-endian UUID conversion is attempted if the primary lookup fails.
func (c *XoClient) FindVMByNode(ctx context.Context, node *v1.Node) (vm *payloads.VM, poolID uuid.UUID, err error) {
	var vmID uuid.UUID
	if node.Status.NodeInfo.SystemUUID == "" {
		vmID, err = GetVMID(node.Spec.ProviderID)
		if err != nil {
			return nil, uuid.Nil, fmt.Errorf("node SystemUUID is empty: %v", err)
		}
	} else {
		vmID, err = uuid.FromString(node.Status.NodeInfo.SystemUUID)
		if err != nil {
			return nil, uuid.Nil, fmt.Errorf("invalid SystemUUID format: %v", err)
		}
	}

	vmClient := c.Client.VM()

	vm, err = vmClient.GetByID(ctx, vmID)
	if err != nil {
		// Some systems (SMBIOS 2.8 / XCP-ng UEFI) report UUIDs with the first 3
		// fields in little-endian; try the converted form as a fallback.
		// See: https://xcp-ng.org/forum/topic/11078
		convertedID := convertLittleEndianUUID(vmID)
		klog.V(4).InfoS("Trying with converted UUID", "original", vmID.String(), "converted", convertedID.String())
		vm, err = vmClient.GetByID(ctx, convertedID)
		if err != nil {
			return nil, uuid.Nil, fmt.Errorf("VM not found with UUID %s or %s: %v", vmID.String(), convertedID.String(), err)
		}
		klog.V(2).InfoS("Found VM with converted UUID (little-endian fix applied)",
			"original", vmID.String(), "converted", convertedID.String())
	}

	klog.V(4).InfoS("Found VM by node", "vm", vm.NameLabel, "uuid", vm.ID.String())
	return vm, vm.PoolID, nil
}

// FindVMByName finds a VM by its name_label across all pools.
func (c *XoClient) FindVMByName(ctx context.Context, name string) (*payloads.VM, uuid.UUID, error) {
	allVMs, err := c.Client.VM().GetAll(ctx, 0, "name_label:"+name)
	if err != nil {
		return nil, uuid.Nil, fmt.Errorf("failed to get list of VMs: %v", err)
	}
	for _, vm := range allVMs {
		if vm.NameLabel == name {
			klog.V(4).InfoS("Found VM by name", "vm", vm.NameLabel)
			return vm, vm.PoolID, nil
		}
	}
	klog.V(4).InfoS("VM not found by name", "name", name)
	return nil, uuid.Nil, fmt.Errorf("vm %q not found", name)
}

// convertLittleEndianUUID converts a little-endian UUID to big-endian.
//
// Due to a bug in XCP-ng UEFI supports, SMBIOS 2.8 is used instead of SMBIOS 2.4
// UUIDs are sent in big-endian format instead of little-endian (Microsoft GUID format) as
// mandated by SMBIOS 2.8. When Linux reads the SMBIOS table, it interprets the UUID
// as little-endian, causing a mismatch between the UUID reported by the guest OS
// (via dmidecode/SystemUUID) and the actual VM UUID in Xen Orchestra.
//
// This function swaps the first 8 bytes to convert from the
// incorrectly-interpreted little-endian format to the correct big-endian format.
//
// Reference: https://xcp-ng.org/forum/topic/11078/vm-uuid-via-dmidecode-does-not-match-vm-id-in-xen-orchestra/23
func convertLittleEndianUUID(u uuid.UUID) uuid.UUID {
	var result uuid.UUID
	copy(result[:], u[:])

	// Swap bytes for first field (4 bytes)
	result[0], result[1], result[2], result[3] = u[3], u[2], u[1], u[0]

	// Swap bytes for second field (2 bytes)
	result[4], result[5] = u[5], u[4]

	// Swap bytes for third field (2 bytes)
	result[6], result[7] = u[7], u[6]

	// Last 8 bytes remain unchanged
	return result
}
