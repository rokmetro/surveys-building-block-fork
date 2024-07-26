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

package interfaces

import (
	"application/core/model"
	"application/driven/calendar"
	"time"
)

// Storage is used by core to storage data - DB storage adapter, file storage adapter etc
type Storage interface {
	RegisterStorageListener(listener StorageListener)
	PerformTransaction(func(storage Storage) error) error

	FindConfig(configType string, appID string, orgID string) (*model.Config, error)
	FindConfigByID(id string) (*model.Config, error)
	FindConfigs(configType *string) ([]model.Config, error)
	InsertConfig(config model.Config) error
	UpdateConfig(config model.Config) error
	DeleteConfig(id string) error

	GetSurvey(id string, orgID string, appID string) (*model.Survey, error)
	GetSurveys(orgID string, appID string, creatorID *string, surveyIDs []string, surveyTypes []string, calendarEventID string, limit *int, offset *int, filter *model.SurveyTimeFilter, public *bool, archived *bool, completed *bool) ([]model.Survey, error)
	CreateSurvey(survey model.Survey) (*model.Survey, error)
	UpdateSurvey(survey model.Survey, admin bool) error
	DeleteSurvey(id string, orgID string, appID string, creatorID string, admin bool) error
	DeleteSurveysWithIDs(orgID string, appID string, accountsIDs []string) error

	GetSurveyResponse(id string, orgID string, appID string, userID string) (*model.SurveyResponse, error)
	GetSurveyResponses(orgID *string, appID *string, userID *string, surveyIDs []string, surveyTypes []string, startDate *time.Time, endDate *time.Time, limit *int, offset *int) ([]model.SurveyResponse, error)
	CreateSurveyResponse(surveyResponse model.SurveyResponse) (*model.SurveyResponse, error)
	UpdateSurveyResponse(surveyResponse model.SurveyResponse) error
	DeleteSurveyResponse(id string, orgID string, appID string, userID string) error
	DeleteSurveyResponses(orgID string, appID string, userID string, surveyIDs []string, surveyTypes []string, startDate *time.Time, endDate *time.Time) error
	DeleteSurveyResponsesWithIDs(orgID string, appID string, accountsIDs []string) error

	GetAlertContacts(orgID string, appID string) ([]model.AlertContact, error)
	GetAlertContact(id string, orgID string, appID string) (*model.AlertContact, error)
	GetAlertContactsByKey(key string, orgID string, appID string) ([]model.AlertContact, error)
	CreateAlertContact(alertContact model.AlertContact) (*model.AlertContact, error)
	UpdateAlertContact(alertContact model.AlertContact) error
	DeleteAlertContact(id string, orgID string, appID string) error
}

// StorageListener represents storage listener
type StorageListener interface {
	OnConfigsUpdated()
	OnExamplesUpdated()
}

// Notifications is the interface for accessing the Notifications BB
type Notifications interface {
	SendNotification(notification model.NotificationMessage)
	SendMail(toEmail string, subject string, body string)
}

// Calendar is the interface for accessing the Calendar BB
type Calendar interface {
	GetEventUsers(orgID string, appID string, eventID string, users []calendar.User, registered *bool, role string, attended *bool) ([]calendar.EventPerson, error)
}
