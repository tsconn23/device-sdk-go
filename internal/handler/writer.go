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
	"encoding/json"
	"fmt"
	"github.com/edgexfoundry/device-sdk-go/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/transformer"
	"github.com/edgexfoundry/device-sdk-go/pkg/models"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"strconv"
	"strings"
	"time"
)

type writer struct {
	Profiles ProfileCache
}

func newWriter(p ProfileCache) writer {
	return writer{Profiles: p}
}

func (w writer) Execute(device *contract.Device, cmd string, params string) common.AppError {
	ros, err := w.Profiles.ResourceOperations(device.Profile.Name, cmd, "set")
	if err != nil {
		msg := fmt.Sprintf("Handler - execWriteCmd: can't find ResrouceOperations in Profile(%s) and Command(%s), %v", device.Profile.Name, cmd, err)
		common.LoggingClient.Error(msg)
		return common.NewBadRequestError(msg, err)
	}

	if len(ros) > common.CurrentConfig.Device.MaxCmdOps {
		msg := fmt.Sprintf("Handler - execWriteCmd: MaxCmdOps (%d) execeeded for dev: %s cmd: %s method: PUT",
			common.CurrentConfig.Device.MaxCmdOps, device.Name, cmd)
		common.LoggingClient.Error(msg)
		return common.NewServerError(msg, nil)
	}

	roMap := roSliceToMap(ros)

	cvs, err := w.parseWriteParams(device.Profile.Name, roMap, params)
	if err != nil {
		msg := fmt.Sprintf("Handler - execWriteCmd: Put parameters parsing failed: %s", params)
		common.LoggingClient.Error(msg)
		return common.NewBadRequestError(msg, err)
	}

	reqs := make([]models.CommandRequest, len(cvs))
	for i, cv := range cvs {
		drName := cv.RO.Object
		common.LoggingClient.Debug(fmt.Sprintf("Handler - execWriteCmd: putting deviceResource: %s", drName))

		// TODO: add recursive support for resource command chaining. This occurs when a
		// deviceprofile resource command operation references another resource command
		// instead of a device resource (see BoschXDK for reference).

		dr, ok := cache.Profiles().DeviceResource(device.Profile.Name, drName)
		common.LoggingClient.Debug(fmt.Sprintf("Handler - execWriteCmd: putting deviceResource: %v", dr))
		if !ok {
			msg := fmt.Sprintf("Handler - execWriteCmd: no deviceResource: %s for dev: %s cmd: %s method: GET", drName, device.Name, cmd)
			common.LoggingClient.Error(msg)
			return common.NewServerError(msg, nil)
		}

		reqs[i].RO = *cv.RO
		reqs[i].DeviceResource = dr

		if common.CurrentConfig.Device.DataTransform {
			err = transformer.TransformWriteParameter(cv, dr.Properties.Value)
			if err != nil {
				msg := fmt.Sprintf("Handler - execWriteCmd: CommandValue (%s) transformed failed: %v", cv.String(), err)
				common.LoggingClient.Error(msg)
				return common.NewServerError(msg, err)
			}
		}
	}

	err = common.Driver.HandleWriteCommands(&device.Addressable, reqs, cvs)
	if err != nil {
		msg := fmt.Sprintf("Handler - execWriteCmd: error for Device: %s cmd: %s, %v", device.Name, cmd, err)
		return common.NewServerError(msg, err)
	}

	return nil
}

func (w writer) parseWriteParams(profileName string, roMap map[string]*contract.ResourceOperation, params string) ([]*models.CommandValue, error) {
	var paramMap map[string]string
	err := json.Unmarshal([]byte(params), &paramMap)
	if err != nil {
		common.LoggingClient.Error(fmt.Sprintf("Handler - parseWriteParams: parsing Write parameters failed %s, %v", params, err))
		return []*models.CommandValue{}, err
	}

	result := make([]*models.CommandValue, 0, len(paramMap))
	for k, v := range paramMap {
		ro, ok := roMap[k]
		if ok {
			if len(ro.Mappings) > 0 {
				newV, ok := ro.Mappings[v]
				if ok {
					v = newV
				} else {
					msg := fmt.Sprintf("Handler - parseWriteParams: Resource Operation (%v) mapping value (%s) failed with the mapping table: %v", ro, v, ro.Mappings)
					common.LoggingClient.Warn(msg)
					//return result, fmt.Errorf(msg) // issue #89 will discuss how to handle there is no mapping matched
				}
			}
			dr, err := checkDeviceResource(profileName, ro, w.Profiles)
			if err != nil {
				return result, err
			}
			cv, err := createCommandValueForParam(dr.Properties.Value.Type, ro, v)
			if err == nil {
				result = append(result, cv)
			} else {
				return result, err
			}
		} else {
			common.LoggingClient.Warn(fmt.Sprintf("Handler - parseWriteParams: The parameter %s cannot find the matched ResourceOperation", k))
		}
	}

	return result, nil
}

func roSliceToMap(ros []contract.ResourceOperation) map[string]*contract.ResourceOperation {
	roMap := make(map[string]*contract.ResourceOperation, len(ros))
	for i, ro := range ros {
		roMap[ro.Parameter] = &ros[i]
	}
	return roMap
}

func createCommandValueForParam(valueType string, ro *contract.ResourceOperation, v string) (*models.CommandValue, error) {
	var result *models.CommandValue
	var err error
	var value interface{}
	var t models.ValueType

	origin := time.Now().UnixNano() / int64(time.Millisecond)

	switch strings.ToLower(valueType) {
	case "bool":
		value, err = strconv.ParseBool(v)
		t = models.Bool
	case "string":
		value = v
		t = models.String
	case "uint8":
		n, e := strconv.ParseUint(v, 10, 8)
		value = uint8(n)
		err = e
		t = models.Uint8
	case "uint16":
		n, e := strconv.ParseUint(v, 10, 16)
		value = uint16(n)
		err = e
		t = models.Uint16
	case "uint32":
		n, e := strconv.ParseUint(v, 10, 32)
		value = uint32(n)
		err = e
		t = models.Uint32
	case "uint64":
		value, err = strconv.ParseUint(v, 10, 64)
		t = models.Uint64
	case "int8":
		n, e := strconv.ParseInt(v, 10, 8)
		value = int8(n)
		err = e
		t = models.Int8
	case "int16":
		n, e := strconv.ParseInt(v, 10, 16)
		value = int16(n)
		err = e
		t = models.Int16
	case "int32":
		n, e := strconv.ParseInt(v, 10, 32)
		value = int32(n)
		err = e
		t = models.Int32
	case "int64":
		value, err = strconv.ParseInt(v, 10, 64)
		t = models.Int64
	case "float32":
		n, e := strconv.ParseFloat(v, 32)
		value = float32(n)
		err = e
		t = models.Float32
	case "float64":
		value, err = strconv.ParseFloat(v, 64)
		t = models.Float64
	}

	if err != nil {
		common.LoggingClient.Error(fmt.Sprintf("Handler - Command: Parsing parameter value (%s) to %s failed: %v", v, valueType, err))
		return result, err
	}

	result, err = models.NewCommandValue(ro, origin, value, t)

	return result, err
}
