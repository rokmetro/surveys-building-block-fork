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
	"time"

	"github.com/rokwire/core-auth-library-go/v3/tokenauth"
)

// Default exposes client APIs for the driver adapters
type Default interface {
	GetVersion() string
}

// Client exposes client APIs for the driver adapters
type Client interface {
	// Surveys
	GetSurvey(id string, orgID string, appID string) (*model.Survey, error)
	GetSurveys(orgID string, appID string, userID *string, creatorID *string, surveyIDs []string, surveyTypes []string, calendarEventID string, limit *int, offset *int, filter *model.SurveyTimeFilter, public *bool, archived *bool, completed *bool) ([]model.Survey, []model.SurveyResponse, error)
	CreateSurvey(survey model.Survey, externalIDs map[string]string) (*model.Survey, error)
	UpdateSurvey(survey model.Survey, userID string, externalIDs map[string]string) error
	DeleteSurvey(id string, orgID string, appID string, userID string, externalIDs map[string]string) error

	// Survey Response
	GetSurveyResponse(id string, orgID string, appID string, userID string) (*model.SurveyResponse, error)
	GetUserSurveyResponses(orgID string, appID string, userID string, surveyIDs []string, surveyTypes []string, startDate *time.Time, endDate *time.Time, limit *int, offset *int) ([]model.SurveyResponse, error)
	GetAllSurveyResponses(orgID string, appID string, userID string, surveyID string, startDate *time.Time, endDate *time.Time, limit *int, offset *int, externalIDs map[string]string) ([]model.SurveyResponse, error)
	CreateSurveyResponse(surveyResponse model.SurveyResponse, externalIDs map[string]string) (*model.SurveyResponse, error)
	UpdateSurveyResponse(surveyResponse model.SurveyResponse) error
	DeleteSurveyResponse(id string, orgID string, appID string, userID string) error
	DeleteSurveyResponses(orgID string, appID string, userID string, surveyIDs []string, surveyTypes []string, startDate *time.Time, endDate *time.Time) error

	// Survey Alerts
	CreateSurveyAlert(surveyAlert model.SurveyAlert) error
}

// Admin exposes administrative APIs for the driver adapters
type Admin interface {
	// Configs
	GetConfig(id string, claims *tokenauth.Claims) (*model.Config, error)
	GetConfigs(configType *string, claims *tokenauth.Claims) ([]model.Config, error)
	CreateConfig(config model.Config, claims *tokenauth.Claims) (*model.Config, error)
	UpdateConfig(config model.Config, claims *tokenauth.Claims) error
	DeleteConfig(id string, claims *tokenauth.Claims) error

	// Surveys
	GetSurvey(id string, orgID string, appID string) (*model.Survey, error)
	GetSurveys(orgID string, appID string, userID *string, creatorID *string, surveyIDs []string, surveyTypes []string, calendarEventID string, limit *int, offset *int, filter *model.SurveyTimeFilter, public *bool, archived *bool, completed *bool) ([]model.Survey, []model.SurveyResponse, error)
	CreateSurvey(survey model.Survey, externalIDs map[string]string) (*model.Survey, error)
	UpdateSurvey(survey model.Survey, userID string, externalIDs map[string]string) error
	DeleteSurvey(id string, orgID string, appID string, userID string, externalIDs map[string]string) error

	// Survey Responses
	GetAllSurveyResponses(orgID string, appID string, surveyID string, userID string, externalIDs map[string]string, startDate *time.Time, endDate *time.Time, limit *int, offset *int) ([]model.SurveyResponse, error)
	GetAllSurveysResponses(orgID string, appID string, surveyID string, userID string, externalIDs map[string]string, startDate *time.Time, endDate *time.Time, limit *int, offset *int) ([]model.SurveyResponse, error)

	// Alert Contacts
	GetAlertContacts(orgID string, appID string) ([]model.AlertContact, error)
	GetAlertContact(id string, orgID string, appID string) (*model.AlertContact, error)
	CreateAlertContact(alertContact model.AlertContact) (*model.AlertContact, error)
	UpdateAlertContact(alertContact model.AlertContact) error
	DeleteAlertContact(id string, orgID string, appID string) error
}

// Analytics exposes Analytics APIs for the driver adapters
type Analytics interface {
	GetAnonymousSurveyResponses(surveyTypes []string, startDate *time.Time, endDate *time.Time) ([]model.SurveyResponseAnonymous, error)
}

// BBs exposes Building Block APIs for the driver adapters
type BBs interface {
}

// TPS exposes third-party service APIs for the driver adapters
type TPS interface {
}

// System exposes system administrative APIs for the driver adapters
type System interface {
}
