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

import "github.com/edgexfoundry/go-mod-core-contracts/models"

type DeviceCache interface {
	ForName(name string) (models.Device, bool)
	ForId(id string) (models.Device, bool)
	All() []models.Device
	Add(device models.Device) error
	Update(device models.Device) error
	UpdateAddressable(addressable models.Addressable) error
	Remove(id string) error
	RemoveByName(name string) error
	UpdateAdminState(id string, state models.AdminState) error
}

type ProfileCache interface {
	ForName(name string) (models.DeviceProfile, bool)
	ForId(id string) (models.DeviceProfile, bool)
	All() []models.DeviceProfile
	Add(profile models.DeviceProfile) error
	Update(profile models.DeviceProfile) error
	Remove(id string) error
	RemoveByName(name string) error
	DeviceResource(profileName string, resourceName string) (models.DeviceResource, bool)
	CommandExists(profileName string, cmd string) (bool, error)
	ResourceOperations(profileName string, cmd string, method string) ([]models.ResourceOperation, error)
	ResourceOperation(profileName string, object string, method string) (models.ResourceOperation, error)
}