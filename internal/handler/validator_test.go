/*******************************************************************************
 * Copyright 2019 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/

package handler

import (
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logging"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"testing"
)

func TestCheckDeviceResource(t *testing.T) {
	common.LoggingClient = logger.MockLogger{}
	cache := newTestProfileCache()
	tests := []struct {
		testName    string
		profileName string
		op          *contract.ResourceOperation
		cache       ProfileCache
		expectErr   bool
	}{
		{"ResourceIsFound", "Simple-Device", &contract.ResourceOperation{Object: "SwitchButton"}, cache, false},
		{"ResourceNotFound", "Simple-Device", &contract.ResourceOperation{Object: "HoldMyBeer"}, cache, true},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			_, err := checkDeviceResource(tt.profileName, tt.op, cache)
			if !tt.expectErr && err != nil {
				t.Errorf("%s: unexpected error for resource %s.", tt.testName, tt.op.Object)
				return
			}
			if tt.expectErr && err == nil {
				t.Errorf("%s: did not receive expected error for resource %s.", tt.testName, tt.op.Object)
				return
			}
		})
	}
}
