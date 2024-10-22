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
	"application/core/model"
	"time"

	"github.com/google/uuid"
	"github.com/rokwire/logging-library-go/v2/errors"
	"github.com/rokwire/logging-library-go/v2/logutils"
)

// appClient contains client implementations
type appClient struct {
	app *Application
}

// Surveys
// GetSurvey returns the survey with the provided ID
func (a appClient) GetSurvey(id string, orgID string, appID string) (*model.Survey, error) {
	return a.app.shared.getSurvey(id, orgID, appID)
}

// GetSurvey returns surveys matching the provided query
func (a appClient) GetSurveys(orgID string, appID string, userID *string, creatorID *string, surveyIDs []string, surveyTypes []string, calendarEventID string,
	limit *int, offset *int, filter *model.SurveyTimeFilter, public *bool, archived *bool, completed *bool) ([]model.Survey, []model.SurveyResponse, error) {
	return a.app.shared.getSurveys(orgID, appID, userID, creatorID, surveyIDs, surveyTypes, calendarEventID, limit, offset, filter, public, archived, completed)
}

// CreateSurvey creates a new survey
func (a appClient) CreateSurvey(survey model.Survey, externalIDs map[string]string) (*model.Survey, error) {
	return a.app.shared.createSurvey(survey, externalIDs)
}

// UpdateSurvey updates the provided survey
func (a appClient) UpdateSurvey(survey model.Survey, userID string, externalIDs map[string]string) error {
	return a.app.shared.updateSurvey(survey, userID, externalIDs, false)
}

// DeleteSurvey deletes the survey with the specified ID
func (a appClient) DeleteSurvey(id string, orgID string, appID string, userID string, externalIDs map[string]string) error {
	return a.app.shared.deleteSurvey(id, orgID, appID, userID, externalIDs, false)
}

// Survey Response
// GetSurveyResponse returns the survey response with the provided ID
func (a appClient) GetSurveyResponse(id string, orgID string, appID string, userID string) (*model.SurveyResponse, error) {
	return a.app.storage.GetSurveyResponse(id, orgID, appID, userID)
}

// GetUserSurveyResponses returns the survey responses matching the provided filters for a specific user
func (a appClient) GetUserSurveyResponses(orgID string, appID string, userID string, surveyIDs []string, surveyTypes []string, startDate *time.Time, endDate *time.Time, limit *int, offset *int) ([]model.SurveyResponse, error) {
	return a.app.storage.GetSurveyResponses(&orgID, &appID, &userID, surveyIDs, surveyTypes, startDate, endDate, limit, offset)
}

// GetAllSurveyResponses returns the survey responses matching the provided filters
func (a appClient) GetAllSurveyResponses(orgID string, appID string, userID string, surveyID string, startDate *time.Time, endDate *time.Time, limit *int, offset *int, externalIDs map[string]string) ([]model.SurveyResponse, error) {
	var allResponses []model.SurveyResponse
	var err error

	survey, err := a.app.shared.getSurvey(surveyID, orgID, appID)
	if err != nil {
		return nil, err
	}

	// Check if survey is sensitive
	if survey.Sensitive {
		return nil, errors.Newf("Survey is sensitive and responses are not available")
	}

	// If no calendar event is associated then user should not have access to responses
	if survey.CalendarEventID == "" {
		return nil, errors.Newf("Survey responses are not available. No calendar event associated")
	}

	// Check if user is admin of calendar event
	admin, err := a.app.shared.isEventAdmin(survey.OrgID, survey.AppID, survey.CalendarEventID, userID, externalIDs)
	if err != nil {
		return nil, errors.WrapErrorAction("checking", "event admin", nil, err)
	}
	if !admin {
		return nil, errors.ErrorData(logutils.StatusInvalid, "user", &logutils.FieldArgs{"calendar_event_id": survey.CalendarEventID, "admin": false})
	}

	// Get responses
	allResponses, err = a.app.storage.GetSurveyResponses(&orgID, &appID, nil, []string{surveyID}, nil, startDate, endDate, limit, offset)
	if err != nil {
		return nil, err
	}

	// If survey is anonymous strip userIDs
	if survey.Anonymous {
		for i := range allResponses {
			allResponses[i].UserID = ""
		}
	}

	return allResponses, nil
}

// CreateSurveyResponse creates a new survey response
func (a appClient) CreateSurveyResponse(surveyResponse model.SurveyResponse, externalIDs map[string]string) (*model.SurveyResponse, error) {
	surveyResponse.ID = uuid.NewString()
	surveyResponse.DateCreated = time.Now().UTC()
	surveyResponse.DateUpdated = nil

	// Get survey from storage
	survey, err := a.app.storage.GetSurvey(surveyResponse.Survey.ID, surveyResponse.OrgID, surveyResponse.AppID)
	if err != nil {
		return nil, err
	}
	// Populate survey with data from client request
	survey.Data = surveyResponse.Survey.Data
	survey.SurveyStats = surveyResponse.Survey.SurveyStats
	survey.ResultJSON = surveyResponse.Survey.ResultJSON
	surveyResponse.Survey = *survey

	if survey.CalendarEventID != "" {
		// check if user attended calendar event
		attended, err := a.app.shared.hasAttendedEvent(surveyResponse.OrgID, surveyResponse.AppID, survey.CalendarEventID, surveyResponse.UserID, externalIDs)
		if err != nil {
			return nil, errors.WrapErrorAction("checking", "event attendance", nil, err)
		}
		if !attended {
			return nil, errors.Newf("user has not attended calendar event")
		}
	}

	return a.app.storage.CreateSurveyResponse(surveyResponse)
}

// UpdateSurveyResponse updates the provided survey response
func (a appClient) UpdateSurveyResponse(surveyResponse model.SurveyResponse) error {
	return a.app.storage.UpdateSurveyResponse(surveyResponse)
}

// DeleteSurveyResponse deletes the survey with the specified ID
func (a appClient) DeleteSurveyResponse(id string, orgID string, appID string, userID string) error {
	return a.app.storage.DeleteSurveyResponse(id, orgID, appID, userID)
}

// DeleteSurveyResponses deletes the survey responses matching the provided filters
func (a appClient) DeleteSurveyResponses(orgID string, appID string, userID string, surveyIDs []string, surveyTypes []string, startDate *time.Time, endDate *time.Time) error {
	return a.app.storage.DeleteSurveyResponses(orgID, appID, userID, surveyIDs, surveyTypes, startDate, endDate)
}

// Survey Alerts
// CreateSurveyAlert creates a new survey alert
func (a appClient) CreateSurveyAlert(surveyAlert model.SurveyAlert) error {
	contacts, err := a.app.storage.GetAlertContactsByKey(surveyAlert.ContactKey, surveyAlert.OrgID, surveyAlert.AppID)
	if err != nil {
		return err
	}

	for i := 0; i < len(contacts); i++ {
		if contacts[i].Type == "email" {
			subject, ok := surveyAlert.Content["subject"].(string)
			if !ok {
				return errors.ErrorData(logutils.StatusMissing, "subject", nil)
			}
			body, ok := surveyAlert.Content["body"].(string)
			if !ok {
				return errors.ErrorData(logutils.StatusMissing, "body", nil)
			}
			a.app.notifications.SendMail(contacts[i].Address, subject, body)
		}
	}

	return nil
}

// GetUserData returns surveys matching the provided query
func (a appClient) GetUserData(orgID string, appID string, userID *string) (*model.UserData, error) {
	return a.app.shared.getUserData(orgID, appID, userID)
}

// newAppClient creates new appClient
func newAppClient(app *Application) appClient {
	return appClient{app: app}
}
