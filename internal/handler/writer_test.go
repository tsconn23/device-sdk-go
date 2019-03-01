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
	cache2 "github.com/edgexfoundry/device-sdk-go/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logging"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/google/uuid"
	"strings"
	"testing"
)

func TestCommandValueCreate(t *testing.T) {
	common.LoggingClient = logger.MockLogger{}
	testOperation := &contract.ResourceOperation{}
	tests := []struct {
		testName   string
		valueType  string
		op         *contract.ResourceOperation
		v          string
		parseCheck models.ValueType
		expectErr  bool
	}{
		{"BoolTruePass", "Bool", testOperation, "true", models.Bool, false},
		{"BoolFalsePass", "Bool", testOperation, "false", models.Bool, false},
		{"BoolFail", "Bool", testOperation, "error", models.Bool, true},
		{"StringPass", "String", testOperation, "hello", models.String, false},
		{"Uint8Pass", "Uint8", testOperation, "123", models.Uint8, false},
		{"Uint8NegativeFail", "Uint8", testOperation, "-123", models.Uint8, true},
		{"Uint8WordFail", "Uint8", testOperation, "hello", models.Uint8, true},
		{"Uint8OverflowFail", "Uint8", testOperation, "9999999999", models.Uint8, true},
		{"Uint16Pass", "Uint16", testOperation, "123", models.Uint16, false},
		{"Uint16NegativeFail", "Uint16", testOperation, "-123", models.Uint16, true},
		{"Uint16WordFail", "Uint16", testOperation, "hello", models.Uint16, true},
		{"Uint16OverflowFail", "Uint16", testOperation, "9999999999", models.Uint16, true},
		{"Uint32Pass", "Uint32", testOperation, "123", models.Uint32, false},
		{"Uint32NegativeFail", "Uint32", testOperation, "-123", models.Uint32, true},
		{"Uint32WordFail", "Uint32", testOperation, "hello", models.Uint32, true},
		{"Uint32OverflowFail", "Uint32", testOperation, "9999999999", models.Uint32, true},
		{"Uint64Pass", "Uint64", testOperation, "123", models.Uint64, false},
		{"Uint64NegativeFail", "Uint64", testOperation, "-123", models.Uint64, true},
		{"Uint64WordFail", "Uint64", testOperation, "hello", models.Uint64, true},
		{"Uint64OverflowFail", "Uint64", testOperation, "99999999999999999999", models.Uint64, true},
		{"Int8Pass", "Int8", testOperation, "123", models.Int8, false},
		{"Int8NegativePass", "Int8", testOperation, "-123", models.Int8, false},
		{"Int8WordFail", "Int8", testOperation, "hello", models.Int8, true},
		{"Int8OverflowFail", "Int8", testOperation, "9999999999", models.Int8, true},
		{"Int16Pass", "Int16", testOperation, "123", models.Int16, false},
		{"Int16NegativePass", "Int16", testOperation, "-123", models.Int16, false},
		{"Int16WordFail", "Int16", testOperation, "hello", models.Int16, true},
		{"Int16OverflowFail", "Int16", testOperation, "9999999999", models.Int16, true},
		{"Int32Pass", "Int32", testOperation, "123", models.Int32, false},
		{"Int32NegativePass", "Int32", testOperation, "-123", models.Int32, false},
		{"Int32WordFail", "Int32", testOperation, "hello", models.Int32, true},
		{"Int32OverflowFail", "Int32", testOperation, "9999999999", models.Int32, true},
		{"Int64Pass", "Int64", testOperation, "123", models.Int64, false},
		{"Int64NegativePass", "Int64", testOperation, "-123", models.Int64, false},
		{"Int64WordFail", "Int64", testOperation, "hello", models.Int64, true},
		{"Int64OverflowFail", "Int64", testOperation, "99999999999999999999", models.Int64, true},
		{"Float32Pass", "Float32", testOperation, "123.000", models.Float32, false},
		{"Float32IntPass", "Float32", testOperation, "123", models.Float32, false},
		{"Float32NegativePass", "Float32", testOperation, "-123.000", models.Float32, false},
		{"Float32WordFail", "Float32", testOperation, "hello", models.Float32, true},
		{"Float32OverflowFail", "Float32", testOperation, "440282346638528859811704183484516925440.0000000000000000", models.Float32, true},
		{"Float64Pass", "Float64", testOperation, "123.000", models.Float64, false},
		{"Float64IntPass", "Float64", testOperation, "123", models.Float64, false},
		{"Float64NegativePass", "Float64", testOperation, "-123.000", models.Float64, false},
		{"Float64WordFail", "Float64", testOperation, "hello", models.Float64, true},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			cv, err := createCommandValueForParam(tt.valueType, tt.op, tt.v)
			if !tt.expectErr && err != nil {
				t.Errorf("%s expectErr:%v error:%v", tt.testName, tt.expectErr, err)
				return
			}
			if tt.expectErr && err == nil {
				t.Errorf("%s expectErr:%v no error thrown", tt.testName, tt.expectErr)
				return
			}
			if cv != nil {
				var check models.ValueType
				switch strings.ToLower(tt.valueType) {
				case "bool":
					check = models.Bool
				case "string":
					check = models.String
				case "uint8":
					check = models.Uint8
				case "uint16":
					check = models.Uint16
				case "uint32":
					check = models.Uint32
				case "uint64":
					check = models.Uint64
				case "int8":
					check = models.Int8
				case "int16":
					check = models.Int16
				case "int32":
					check = models.Int32
				case "int64":
					check = models.Int64
				case "float32":
					check = models.Float32
				case "float64":
					check = models.Float64
				}
				if cv.Type != check {
					t.Errorf("%s incorrect parsing. valueType: %s result: %v", tt.testName, tt.valueType, cv.Type)
				}
			}
		})
	}
}

func TestResourceOpSliceToMap(t *testing.T) {
	ops := []contract.ResourceOperation{}

	ops = append(ops, contract.ResourceOperation{Parameter: "first"})
	ops = append(ops, contract.ResourceOperation{Parameter: "second"})
	ops = append(ops, contract.ResourceOperation{Parameter: "third"})

	mapped := roSliceToMap(ops)

	if len(mapped) != 3 {
		t.Errorf("unexpected map length. wanted 3, got %v", len(mapped))
		return
	}

	tests := []struct {
		testName  string
		key       string
		expectErr bool
	}{
		{"findFirst", "first", false},
		{"findSecond", "second", false},
		{"findThird", "third", false},
		{"notFoundKey", "fourth", true},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			_, ok := mapped[tt.key]
			if !tt.expectErr && !ok {
				t.Errorf("expected entry %s not found in map.", tt.key)
				return
			}
			if tt.expectErr && ok {
				t.Errorf("test %s expected error not received.", tt.testName)
				return
			}
		})
	}
}

func TestParseWriteParams(t *testing.T) {
	profileName := "Simple-Device"
	common.LoggingClient = logger.MockLogger{}
	cache := newTestProfileCache()
	w := newWriter(cache)
	profile, ok := cache.ForName(profileName)
	if !ok {
		t.Errorf("device profile was not found, cannot continue")
		return
	}
	roMap := roSliceToMap(profile.Resources[0].Set)
	tests := []struct {
		testName    string
		profile     string
		resourceOps map[string]*contract.ResourceOperation
		params      string
		expectErr   bool
	}{
		{"ValidWriteParam", profileName, roMap, "{\"Switch\":\"True\"}", false},
		//The expectErr on the test below is false because parseWriteParams does NOT throw an error
		//if the specified parameter isn't found.
		{"InvalidWriteParam", profileName, roMap, "{\"NotFound\":\"True\"}", false},
		{"InvalidWriteParamType", profileName, roMap, "{\"Switch\":\"123\"}", true},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			_, err := w.parseWriteParams(tt.profile, tt.resourceOps, tt.params)
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected parse error params:%s %s", tt.params, err.Error())
				return
			}
			if tt.expectErr && err == nil {
				t.Errorf("expected error was not received params:%s", tt.params)
				return
			}
			//I would like to add an additional check here to ensure the returned array of CommandValues isn't empty
			//in the case of a successful parse. However that isn't possible due to the swallowing of the invalid
			//parameter described above. Both successful parse and invalid parameter look the same from an error handling
			//perspective.
		})
	}
}

func newTestProfileCache() ProfileCache {
	profile := newTestDeviceProfile()
	cache := cache2.NewProfileCache([]contract.DeviceProfile{profile})
	return cache
}

func newTestDeviceProfile() contract.DeviceProfile {
	profile := contract.DeviceProfile{}
	profile.Description = "Example of Simple Device"
	profile.Id = uuid.New().String()
	profile.Labels = []string{"modbus"}
	profile.Manufacturer = "Simple Corp."
	profile.Model = "SP-01"
	profile.Name = "Simple-Device"
	profile.Commands = []contract.Command{}

	dResource := newTestDeviceResource()
	profile.DeviceResources = []contract.DeviceResource{dResource}

	pResource := newTestProfileResource()
	profile.Resources = []contract.ProfileResource{pResource}

	getResponses := []contract.Response{}
	getResponses = append(getResponses, contract.Response{Code: "200", ExpectedValues: []string{"Switch"}})
	getResponses = append(getResponses, contract.Response{Code: "503", Description: "service unavailable"})

	getAction := contract.Action{Path: "/api/v1/device/{deviceId}/Switch", Responses: getResponses}
	get := contract.Get{Action: getAction}

	deviceCommand := contract.Command{Name: "Switch", Get: &get}

	putResponses := []contract.Response{}
	putResponses = append(putResponses, contract.Response{Code: "200"})
	putResponses = append(putResponses, contract.Response{Code: "503", Description: "service unavailable"})

	putAction := contract.Action{Path: "/api/v1/device/{deviceId}/Switch", Responses: putResponses}
	put := contract.Put{Action: putAction, ParameterNames: []string{"Switch"}}
	deviceCommand.Put = &put

	profile.Commands = append(profile.Commands, deviceCommand)

	return profile
}

func newTestDeviceResource() contract.DeviceResource {
	dr := contract.DeviceResource{Name: "SwitchButton", Description: "Switch On/Off."}

	pv := contract.PropertyValue{Type: "Bool", ReadWrite: "RW"}
	u := contract.Units{Type: "String", ReadWrite: "R", DefaultValue: "On/Off"}

	dr.Properties = contract.ProfileProperty{Value: pv, Units: u}
	return dr
}

func newTestProfileResource() contract.ProfileResource {
	pr := contract.ProfileResource{Name: "Switch"}

	get := contract.ResourceOperation{Operation: "Get", Object: "SwitchButton", Parameter: "Switch"}
	pr.Get = []contract.ResourceOperation{get}

	set := contract.ResourceOperation{Operation: "Set", Object: "SwitchButton", Parameter: "Switch"}
	pr.Set = []contract.ResourceOperation{set}

	return pr
}
