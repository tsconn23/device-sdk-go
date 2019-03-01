// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package handler

import (
	"fmt"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"strings"
	"sync"
)

type Command struct {
	devices  DeviceCache
	profiles ProfileCache
}

func NewCommand(d DeviceCache, p ProfileCache) Command {
	return Command { devices:d, profiles: p}
}

// Note, every HTTP request to ServeHTTP is made in a separate goroutine, which
// means care needs to be taken with respect to shared data accessed through *Server.
func (c Command) Handle(vars map[string]string, body string, method string) (*contract.Event, common.AppError) {
	dKey := vars["id"]
	cmd := vars["command"]

	var ok bool
	var d contract.Device
	if dKey != "" {
		d, ok = c.devices.ForId(dKey)
	} else {
		dKey = vars["name"]
		d, ok = c.devices.ForName(dKey)
	}
	if !ok {
		msg := fmt.Sprintf("Device: %s not found; %s", dKey, method)
		common.LoggingClient.Error(msg)
		return nil, common.NewNotFoundError(msg, nil)
	}

	if d.AdminState == "LOCKED" {
		msg := fmt.Sprintf("%s is locked; %s", d.Name, method)
		common.LoggingClient.Error(msg)
		return nil, common.NewLockedError(msg, nil)
	}

	// TODO: need to mark device when operation in progress, so it can't be removed till completed

	// NOTE: as currently implemented, CommandExists checks the existence of a deviceprofile
	// *resource* name, not a *command* name! A deviceprofile's command section is only used
	// to trigger valuedescriptor creation.
	exists, err := c.profiles.CommandExists(d.Profile.Name, cmd)

	// TODO: once cache locking has been implemented, this should never happen
	if err != nil {
		msg := fmt.Sprintf("internal error; Device: %s searching %s in cache failed; %s", d.Name, cmd, method)
		common.LoggingClient.Error(msg)
		return nil, common.NewServerError(msg, err)
	}

	if !exists {
		msg := fmt.Sprintf("%s for Device: %s not found; %s", cmd, d.Name, method)
		common.LoggingClient.Error(msg)
		return nil, common.NewNotFoundError(msg, nil)
	}

	if strings.ToLower(method) == "get" {
		r := newReader(c.profiles)
		return r.Execute(&d, cmd)
	} else {
		w := newWriter(c.profiles)
		appErr := w.Execute(&d, cmd, body)
		return nil, appErr
	}
}

func (c Command) HandleAll(cmd string, body string, method string) ([]*contract.Event, common.AppError) {
	common.LoggingClient.Debug(fmt.Sprintf("Handler - CommandAll: execute the %s command %s from all operational devices", method, cmd))
	filtered := filterOperationalDevices(c.devices.All())

	devCount := len(filtered)
	var waitGroup sync.WaitGroup
	waitGroup.Add(devCount)
	cmdResults := make(chan struct {
		event  *contract.Event
		appErr common.AppError
	}, devCount)

	for i, _ := range filtered {
		go func(device *contract.Device) {
			defer waitGroup.Done()
			var event *contract.Event = nil
			var appErr common.AppError = nil
			if strings.ToLower(method) == "get" {
				r := newReader(c.profiles)
				event, appErr = r.Execute(device, cmd)
			} else {
				w := newWriter(c.profiles)
				appErr = w.Execute(device, cmd, body)
			}
			cmdResults <- struct {
				event  *contract.Event
				appErr common.AppError
			}{event, appErr}
		}(filtered[i])
	}
	waitGroup.Wait()
	close(cmdResults)

	errCount := 0
	getResults := make([]*contract.Event, 0, devCount)
	var appErr common.AppError
	for r := range cmdResults {
		if r.appErr != nil {
			errCount++
			common.LoggingClient.Error("Handler - CommandAll: " + r.appErr.Message())
			appErr = r.appErr // only the last error will be returned
		} else if r.event != nil {
			getResults = append(getResults, r.event)
		}
	}

	if errCount < devCount {
		common.LoggingClient.Info("Handler - CommandAll: part of commands executed successfully, returning 200 OK")
		appErr = nil
	}

	return getResults, appErr

}

func filterOperationalDevices(devices []contract.Device) []*contract.Device {
	result := make([]*contract.Device, 0, len(devices))
	for i, d := range devices {
		if (d.AdminState == contract.Locked) || (d.OperatingState == contract.Disabled) {
			continue
		}
		result = append(result, &devices[i])
	}
	return result
}
