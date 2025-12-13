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

package core

import (
	"application/core/interfaces"
	"application/core/model"
	"application/driven/airship"
	corebb "application/driven/core"
	"application/driven/notifications"
	"time"

	"github.com/rokwire/rokwire-building-block-sdk-go/utils/errors"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logs"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logutils"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/rokwireutils"
)

type storageListener struct {
	app *Application
	model.DefaultStorageListener
}

// OnExampleUpdated notifies that the example collection has changed
func (s *storageListener) OnExampleUpdated() {
	s.app.logger.Infof("OnExampleUpdated")

	// TODO: Implement listener
}

// Application represents the core application code based on hexagonal architecture
type Application struct {
	version string
	build   string

	Default   interfaces.Default   // expose to the drivers adapters
	Client    interfaces.Client    // expose to the drivers adapters
	Admin     interfaces.Admin     // expose to the drivers adapters
	Analytics interfaces.Analytics // expose to the drivers adapters
	BBs       interfaces.BBs       // expose to the drivers adapters
	TPS       interfaces.TPS       // expose to the drivers adapters
	System    interfaces.System    // expose to the drivers adapters
	shared    Shared

	logger *logs.Logger

	storage             interfaces.Storage
	notifications       interfaces.Notifications
	calendar            interfaces.Calendar
	corebb              *corebb.Adapter
	airship             *airship.Adapter
	deleteDataLogic     deleteDataLogic
	streakNotifications streakNotifications
	useExternalUserID   bool
}

// Start starts the core part of the application
func (a *Application) Start() {
	//set storage listener
	storageListener := storageListener{app: a}
	a.storage.RegisterStorageListener(&storageListener)
	a.deleteDataLogic.start()
	a.streakNotifications.start()
}

// GetEnvConfigs retrieves the cached database env configs
func (a *Application) GetEnvConfigs() (*model.EnvConfigData, error) {
	// Load env configs from database
	config, err := a.storage.FindConfig(model.ConfigTypeEnv, rokwireutils.AllApps, rokwireutils.AllOrgs)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, model.TypeConfig, nil, err)
	}
	if config == nil {
		return nil, errors.ErrorData(logutils.StatusMissing, model.TypeConfig, &logutils.FieldArgs{"type": model.ConfigTypeEnv, "app_id": rokwireutils.AllApps, "org_id": rokwireutils.AllOrgs})
	}
	return model.GetConfigData[model.EnvConfigData](*config)
}

// CheckTimersConfig retrieves the database timers config
func (a *Application) CheckTimersConfig(key string, filterTime time.Time, updateTime time.Time) (*model.Config, error) {
	// Load env configs from database
	config, err := a.storage.FindAndUpdateTimerConfig(rokwireutils.AllApps, rokwireutils.AllOrgs, key, filterTime, updateTime)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeConfig, nil, err)
	}
	return config, nil
}

// SendQuizNotifications sends quiz notifications to all notification services
func (a *Application) SendQuizNotifications(orgID string, appID string, userIDs []string, externalUserIDs []string, body string, notificationData map[string]string) {
	// Send to Airship (all externalUserIDs at once)
	a.airship.SendNotification(orgID, appID, externalUserIDs, airship.SubjectVogue, body, notificationData, []string{airship.TagQuizAll}, nil)

	// Send to Notifications BB (all userIDs)
	recipients := make([]model.NotificationMessageRecipient, len(userIDs))
	for i, userID := range userIDs {
		recipients[i] = model.NotificationMessageRecipient{UserID: userID}
	}
	topic := notifications.TopicQuizAll
	message := model.NotificationMessage{
		OrgID:      orgID,
		AppID:      appID,
		Subject:    notifications.SubjectVogue,
		Body:       body,
		Data:       notificationData,
		Recipients: recipients,
		Topic:      &topic,
	}
	a.notifications.SendNotification(message)
}

// NewApplication creates new Application
func NewApplication(version string, build string, storage interfaces.Storage, notifications interfaces.Notifications, calendar interfaces.Calendar,
	coreBB *corebb.Adapter, airship *airship.Adapter, serviceID string, logger *logs.Logger, useExternalUserID bool) *Application {
	deleteDataLogic := deleteDataLogic{logger: *logger, core: coreBB, serviceID: serviceID, storage: storage}

	application := Application{version: version, build: build, storage: storage, notifications: notifications,
		calendar: calendar, corebb: coreBB, airship: airship, deleteDataLogic: deleteDataLogic, logger: logger,
		useExternalUserID: useExternalUserID}

	streakNotificationsTimerDone := make(chan bool)
	streakNotifications := streakNotifications{application: &application, logger: logger, storage: storage, streakNotificationsTimerDone: streakNotificationsTimerDone}

	application.streakNotifications = streakNotifications

	//add the drivers ports/interfaces
	application.Default = newAppDefault(&application)
	application.Client = newAppClient(&application)
	application.Admin = newAppAdmin(&application)
	application.Analytics = newAppAnalytics(&application)
	application.BBs = newAppBBs(&application)
	application.TPS = newAppTPS(&application)
	application.System = newAppSystem(&application)
	application.shared = newAppShared(&application)

	return &application
}
