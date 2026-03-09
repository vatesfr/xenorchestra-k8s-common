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
	"fmt"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/vatesfr/xenorchestra-go-sdk/pkg/payloads"
	mock_library "github.com/vatesfr/xenorchestra-k8s-common/mocks"
)

func TestConvertLittleEndianUUID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "All zeros",
			input:    "00000000-0000-0000-0000-000000000000",
			expected: "00000000-0000-0000-0000-000000000000",
		},
		{
			name:     "All ones",
			input:    "ffffffff-ffff-ffff-ffff-ffffffffffff",
			expected: "ffffffff-ffff-ffff-ffff-ffffffffffff",
		},
		{
			name:     "Example 1",
			input:    "59aee6ae-f0b8-a2bd-a89d-55638d1e9725",
			expected: "aee6ae59-b8f0-bda2-a89d-55638d1e9725",
		},
		{
			name:     "Example 2",
			input:    "6c0bc35f-7e48-34fe-04c3-67812e3b17a7",
			expected: "5fc30b6c-487e-fe34-04c3-67812e3b17a7",
		},
		{
			name:     "Example 3",
			input:    "77f4080b-1a49-82a9-23c4-d224723624ea",
			expected: "0b08f477-491a-a982-23c4-d224723624ea",
		},
		{
			name:     "Example 4",
			input:    "6a87cb0f-ca4c-ffa5-3ca2-fc398fb25eac",
			expected: "0fcb876a-4cca-a5ff-3ca2-fc398fb25eac",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, err := uuid.FromString(tt.input)
			assert.NoError(t, err, "Failed to parse input UUID")

			expected, err := uuid.FromString(tt.expected)
			assert.NoError(t, err, "Failed to parse expected UUID")

			result := convertLittleEndianUUID(input)

			assert.Equal(t, expected, result, "Converted UUID does not match expected value")
		})
	}
}

// func newInstanceEnv() (*XOConfig, error) {
// 	// TODO: replace with real test environment variables or configuration
// 	cfg, err := readCloudConfig(strings.NewReader(`
// url: http://127.0.0.1:9000
// token: "secret"
// insecure: true
// `))

// 	return &cfg, err
// }

func newMockedVMClient(_ *testing.T, ctrl *gomock.Controller) *XoClient {
	// Mock VM service
	mockVM := mock_library.NewMockVM(ctrl)
	mockVM.EXPECT().GetAll(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*payloads.VM{
		{
			ID:        uuid.Must(uuid.FromString("550e8400-e29b-41d4-a716-446655440001")),
			NameLabel: "test1-vm",
			PoolID:    uuid.Must(uuid.FromString("a3c8f86b-9c2f-4c3d-8a7b-2d44e6f77f1d")),
		},
		{
			ID:        uuid.Must(uuid.FromString("550e8400-e29b-41d4-a716-446655440002")),
			NameLabel: "test2-vm",
			PoolID:    uuid.Must(uuid.FromString("a3c8f86b-9c2f-4c3d-8a7b-2d44e6f77f1d")),
		},
	}, nil).AnyTimes()
	// Mock Library
	mockLib := mock_library.NewMockLibrary(ctrl)
	mockLib.EXPECT().VM().Return(mockVM).AnyTimes()

	// Inject mock into XOClient
	return &XoClient{
		Client: mockLib,
	}
}

// func TestNewClient(t *testing.T) {
// 	cfg, err := newInstanceEnv()
// 	assert.Nil(t, err)
// 	assert.NotNil(t, cfg)

// 	client, err := xenorchestra.NewInstance(&xenorchestra.XOConfig{})
// 	assert.NotNil(t, err)
// 	assert.Nil(t, client)

// 	client, err = xenorchestra.NewInstance(cfg)
// 	assert.Nil(t, err)
// }

func TestCheckInstance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := newMockedVMClient(t, ctrl)
	err := client.CheckClient(t.Context())
	assert.Nil(t, err)
}

func TestFindVMByNameNonExist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := newMockedVMClient(t, ctrl)

	vm, poolID, err := client.FindVMByName(t.Context(), "non-existing-vm")
	assert.NotNil(t, err)
	assert.Equal(t, uuid.Nil, poolID)
	assert.Nil(t, vm)
	assert.Contains(t, err.Error(), "vm \"non-existing-vm\" not found")
}

func TestFindVMByNameExist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := newMockedVMClient(t, ctrl)

	tests := []struct {
		msg            string
		vmName         string
		expectedError  error
		expectedVMID   uuid.UUID
		expectedPoolID uuid.UUID
	}{
		{
			msg:           "vm not found",
			vmName:        "non-existing-vm",
			expectedError: fmt.Errorf("vm \"non-existing-vm\" not found"),
		},
		{
			msg:            "Test1-VM",
			vmName:         "test1-vm",
			expectedVMID:   uuid.Must(uuid.FromString("550e8400-e29b-41d4-a716-446655440001")),
			expectedPoolID: uuid.Must(uuid.FromString("a3c8f86b-9c2f-4c3d-8a7b-2d44e6f77f1d")),
		},
		{
			msg:            "Test2-VM",
			vmName:         "test2-vm",
			expectedVMID:   uuid.Must(uuid.FromString("550e8400-e29b-41d4-a716-446655440002")),
			expectedPoolID: uuid.Must(uuid.FromString("a3c8f86b-9c2f-4c3d-8a7b-2d44e6f77f1d")),
		},
	}

	for _, testCase := range tests {
		t.Run(fmt.Sprint(testCase.msg), func(t *testing.T) {
			vmr, poolID, err := client.FindVMByName(t.Context(), testCase.vmName)

			if testCase.expectedError == nil {
				assert.Nil(t, err)
				assert.NotNil(t, vmr)
				assert.Equal(t, testCase.expectedVMID, vmr.ID)
				assert.Equal(t, testCase.expectedPoolID, poolID)
			} else {
				assert.NotNil(t, err)
				assert.Equal(t, uuid.Nil, poolID)
				assert.Nil(t, vmr)
				assert.Contains(t, err.Error(), "vm \"non-existing-vm\" not found")
			}
		})
	}
}
