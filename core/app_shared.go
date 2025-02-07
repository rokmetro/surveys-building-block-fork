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
	"application/driven/calendar"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rokwire/logging-library-go/v2/errors"
	"github.com/rokwire/logging-library-go/v2/logutils"
)

// appShared contains shared implementations
type appShared struct {
	app *Application
}

func (a appShared) getSurvey(id string, orgID string, appID string) (*model.Survey, error) {
	return a.app.storage.GetSurvey(id, orgID, appID)
}

func (a appShared) getSurveys(orgID string, appID string, userID *string, creatorID *string, surveyIDs []string, surveyTypes []string, calendarEventID string, limit *int, offset *int, filter *model.SurveyTimeFilter, public *bool, archived *bool, completed *bool) ([]model.Survey, []model.SurveyResponse, error) {
	surveys, surveysResponse, err := a.app.storage.GetSurveysAndSurveyResponses(orgID, appID, creatorID, surveyIDs, surveyTypes, calendarEventID,
		public, archived, limit, offset, userID, filter)
	if err != nil {
		return nil, nil, err
	}

	return surveys, surveysResponse, nil
}

func (a appShared) createSurvey(survey model.Survey, externalIDs map[string]string) (*model.Survey, error) {
	survey.ID = uuid.NewString()
	survey.DateCreated = time.Now().UTC()
	survey.DateUpdated = nil

	if survey.CalendarEventID != "" {
		// check if user is admin of calendar event
		admin, err := a.isEventAdmin(survey.OrgID, survey.AppID, survey.CalendarEventID, survey.CreatorID, externalIDs)
		if err != nil {
			return nil, errors.WrapErrorAction("checking", "event admin", nil, err)
		}
		if !admin {
			return nil, errors.Newf("account not an admin of calendar event")
		}
	}

	return a.app.storage.CreateSurvey(survey)
}

func (a appShared) updateSurvey(survey model.Survey, userID string, externalIDs map[string]string, admin bool) error {
	// if user is not already an admin and survey has associated event, check if user is event admin
	if !admin && survey.CalendarEventID != "" {
		var err error
		admin, err = a.isEventAdmin(survey.OrgID, survey.AppID, survey.CalendarEventID, userID, externalIDs)
		if err != nil {
			return errors.WrapErrorAction("checking", "event admin", nil, err)
		}
	}

	return a.app.storage.UpdateSurvey(survey, admin)
}

func (a appShared) deleteSurvey(id string, orgID string, appID string, userID string, externalIDs map[string]string, admin bool) error {
	transaction := func(storage interfaces.Storage) error {
		//1. find survey
		survey, err := storage.GetSurvey(id, orgID, appID)
		if err != nil {
			return errors.WrapErrorAction(logutils.ActionGet, model.TypeSurvey, nil, err)
		}
		if survey == nil {
			return errors.ErrorData(logutils.StatusMissing, model.TypeSurvey, &logutils.FieldArgs{"id": id, "app_id": appID, "org_id": orgID})
		}

		//2. if user is not already an admin and survey has associated event, check if user is event admin
		if !admin && survey.CalendarEventID != "" {
			admin, err = a.isEventAdmin(survey.OrgID, survey.AppID, survey.CalendarEventID, userID, externalIDs)
			if err != nil {
				return errors.WrapErrorAction("checking", "event admin", nil, err)
			}
		}

		//3. delete survey
		err = storage.DeleteSurvey(survey.ID, survey.OrgID, survey.AppID, userID, admin)
		if err != nil {
			return errors.WrapErrorAction(logutils.ActionDelete, model.TypeSurvey, nil, err)
		}

		return nil
	}

	return a.app.storage.PerformTransaction(transaction)
}

func (a appShared) isEventAdmin(orgID string, appID string, eventID string, userID string, externalIDs map[string]string) (bool, error) {
	// Get external ID
	envConfig, err := a.app.GetEnvConfigs()
	if err != nil {
		return false, errors.WrapErrorAction(logutils.ActionGet, model.TypeConfig, logutils.StringArgs(model.ConfigTypeEnv), err)
	}
	externalID := externalIDs[envConfig.ExternalID]

	eventUsers, err := a.app.calendar.GetEventUsers(orgID, appID, eventID, []calendar.User{{AccountID: userID, ExternalID: externalID}}, nil, calendar.EventRoleAdmin, nil)
	if err != nil {
		return false, errors.WrapErrorAction(logutils.ActionGet, calendar.TypeCalendarUser, &logutils.FieldArgs{"calendar_event_id": eventID, "user_id": userID, "external_id": externalID, "role": calendar.EventRoleAdmin}, err)
	}
	for _, eventUser := range eventUsers {
		// the user is an event admin if there is an account ID match or external ID match and the user has the admin role
		if ((externalID != "" && eventUser.User.ExternalID == externalID) || eventUser.User.AccountID == userID) && eventUser.Role == calendar.EventRoleAdmin {
			return true, nil
		}
	}

	return false, nil
}

func (a appShared) hasAttendedEvent(orgID string, appID string, eventID string, userID string, externalIDs map[string]string) (bool, error) {
	// Get external ID
	envConfig, err := a.app.GetEnvConfigs()
	if err != nil {
		return false, errors.WrapErrorAction(logutils.ActionGet, model.TypeConfig, logutils.StringArgs(model.ConfigTypeEnv), err)
	}
	externalID := externalIDs[envConfig.ExternalID]

	attended := true
	registered := true
	eventUsers, err := a.app.calendar.GetEventUsers(orgID, appID, eventID, []calendar.User{{AccountID: userID, ExternalID: externalID}}, &registered, "", &attended)
	if err != nil {
		return false, errors.WrapErrorAction(logutils.ActionGet, calendar.TypeCalendarUser, &logutils.FieldArgs{"calendar_event_id": eventID, "user_id": userID, "external_id": externalID, "role": calendar.EventRoleAdmin}, err)
	}
	for _, eventUser := range eventUsers {
		if ((externalID != "" && eventUser.User.ExternalID == externalID) || eventUser.User.AccountID == userID) && eventUser.Attended {
			return true, nil
		}
	}

	return false, nil
}

func (a appShared) getUserData(orgID string, appID string, userID *string) (*model.UserData, error) {
	var (
		surveys          []model.Survey
		surveysResponses []model.SurveyResponse
		surveysErr       error
		responsesErr     error
	)

	// Create a WaitGroup to wait for goroutines to finish
	var wg sync.WaitGroup
	wg.Add(2)

	// Fetch surveys asynchronously
	go func() {
		defer wg.Done()
		surveys, surveysErr = a.app.storage.GetSurveysLight(orgID, appID, userID)
	}()

	// Fetch survey responses asynchronously
	go func() {
		defer wg.Done()
		surveysResponses, responsesErr = a.app.storage.GetSurveyResponses(&orgID, &appID, userID, nil, nil, nil, nil, nil, nil)
	}()

	// Wait for both goroutines to complete
	wg.Wait()

	// Check for errors from either function
	if surveysErr != nil {
		return nil, surveysErr
	}

	if responsesErr != nil {
		return nil, responsesErr
	}

	// If surveys or surveyResponses are nil, initialize them to nil (already nil by default in Go)
	userData := model.UserData{
		SurveyUserData:         &surveys,
		SurveyResponseUserData: &surveysResponses,
	}

	return &userData, nil
}

// newAppShared creates new appShared
func newAppShared(app *Application) appShared {
	return appShared{app: app}
}
