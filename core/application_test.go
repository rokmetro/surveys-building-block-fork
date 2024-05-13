// Copyright 2022 Board of Trustees of the University of Illinois.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package core_test

import (
	"application/core"
	"application/core/interfaces"
	"application/core/interfaces/mocks"
	"application/core/model"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/rokwire/core-auth-library-go/v3/authutils"
	"github.com/rokwire/logging-library-go/v2/logs"
)

const (
	serviceID string = "surveys"
)

func buildTestApplication(storage interfaces.Storage) *core.Application {
	loggerOpts := logs.LoggerOpts{SuppressRequests: logs.NewStandardHealthCheckHTTPRequestProperties(serviceID + "/version")}
	logger := logs.NewLogger(serviceID, &loggerOpts)
	return core.NewApplication("1.1.1", "build", storage, nil, nil, logger)
}

func TestApplication_Start(t *testing.T) {
	storage := mocks.NewStorage(t)
	storage.On("RegisterStorageListener", mock.AnythingOfType("*core.storageListener"))
	app := buildTestApplication(storage)

	app.Start()

	storage.AssertCalled(t, "RegisterStorageListener", mock.AnythingOfType("*core.storageListener"))
}

func TestApplication_GetEnvConfigs(t *testing.T) {
	data := model.EnvConfigData{AnalyticsToken: "example"}
	config := model.Config{Type: model.ConfigTypeEnv, AppID: authutils.AllApps, OrgID: authutils.AllOrgs, Data: data, DateCreated: time.Now(), DateUpdated: nil}

	storage := mocks.NewStorage(t)
	storage.On("FindConfig", model.ConfigTypeEnv, authutils.AllApps, authutils.AllOrgs).Return(&config, nil)
	app := buildTestApplication(storage)

	storage2 := mocks.NewStorage(t)
	storage2.On("FindConfig", model.ConfigTypeEnv, authutils.AllApps, authutils.AllOrgs).Return(nil, errors.New("no config found"))
	app2 := buildTestApplication(storage2)

	tests := []struct {
		name    string
		a       *core.Application
		want    *model.EnvConfigData
		wantErr bool
	}{
		{"exists", app, &data, false},
		{"missing", app2, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.a.GetEnvConfigs()
			if (err != nil) != tt.wantErr {
				t.Errorf("Application.GetEnvConfigs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Application.GetEnvConfigs() = %v, want %v", got, tt.want)
			}
		})
	}
}
