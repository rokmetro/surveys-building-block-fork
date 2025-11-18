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
	"application/utils"
	"time"

	"github.com/google/uuid"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/errors"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logs"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logutils"
)

// appShared contains shared implementations
type appShared struct {
	app *Application
}

func (a appShared) getSurvey(id string, orgID string, appID string) (*model.Survey, error) {
	return a.app.storage.GetSurvey(id, orgID, appID)
}

func (a appShared) getSurveys(orgID string, appID string, userID *string, creatorID *string, surveyIDs []string, surveyTypes []string, calendarEventID string, limit *int, offset *int, filter *model.SurveyTimeFilter, public *bool, archived *bool, completed *bool, includeResponses *bool, sortByDateCreated *bool, unstrucProps map[string]interface{}, query *string) ([]model.Survey, error) {
	return a.app.storage.GetSurveysWithResponses(orgID, appID, userID, creatorID, surveyIDs, surveyTypes, calendarEventID, limit, offset, filter, public, archived, completed, includeResponses, sortByDateCreated, unstrucProps, query)
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

// createScore Creates a score object by iterating over all previous survey responses
func (a appShared) createScore(orgID string, appID string, userID string, externalProfileID string, apply bool, externalIDs map[string]string) (*model.Score, error) {
	surveyResponses, err := a.app.storage.GetSurveyResponses(&orgID, &appID, &userID, nil, []string{model.SurveyTypeFashionQuiz}, nil, nil, nil, nil)
	if err != nil {
		return nil, err
	}

	score := model.Score{
		ID:                 uuid.NewString(),
		OrgID:              orgID,
		AppID:              appID,
		UserID:             userID,
		ExternalProfileID:  externalProfileID,
		Score:              0,
		ResponseCount:      0,
		CurrentStreak:      0,
		AnswerCount:        0,
		CorrectAnswerCount: 0,
		SurveyType:         model.SurveyTypeFashionQuiz,
	}

	// Iterate over responses in reverse order (need to start with oldest first, date sorted descending by default)
	for i := len(surveyResponses) - 1; i >= 0; i-- {
		a.updateScore(&score, surveyResponses[i], nil)
	}

	// Sync AmgUUID from Core BB before saving the score
	// This populates the external_user_id field if we have a mastodon_id
	// Note: SyncAmgUUIDForScore handles errors gracefully (returns nil on errors, logs internally)
	a.app.SyncAmgUUIDForScore(&score, externalIDs)

	if apply {
		err = a.app.storage.CreateScore(score)
		if err != nil {
			return nil, errors.WrapErrorAction(logutils.ActionCreate, model.TypeScore, nil, err)
		}
	}
	return &score, nil
}

// updateScore updates the score model passed in
// Assumes that surveyResponse is a fashion quiz
func (a appShared) updateScore(score *model.Score, surveyResponse model.SurveyResponse, l *logs.Log) {
	survey := surveyResponse.Survey
	score.ResponseCount++
	score.AnswerCount += uint32(survey.SurveyStats.Total)
	pointsForResponse := float64(survey.SurveyStats.Scores[""])

	if survey.SurveyStats.CorrectAnswerCount == 0 && pointsForResponse >= 0 {
		// Handle clients that don't send correct answer count by assuming
		// pointsForResponse == correct answer count
		score.CorrectAnswerCount += uint32(pointsForResponse)
	} else {
		score.CorrectAnswerCount += uint32(survey.SurveyStats.CorrectAnswerCount)
	}

	responseTime := surveyResponse.DateCreated
	unstructProps := survey.UnstructuredProperties
	if unstructProps == nil {
		unstructProps = survey.UnstructuredPropertiesDep
	}
	if unstructProps != nil {
		externalProfileIDRaw, exists := unstructProps["external_profile_id"]
		if exists {
			externalProfileIDStr, isString := externalProfileIDRaw.(string)
			if isString {
				score.ExternalProfileID = externalProfileIDStr
			}
		}

		localResponseTimeRaw, exists := unstructProps["local_time"]
		if exists {
			localResponseTimeStr, isString := localResponseTimeRaw.(string)
			if isString {
				localResponseTime, err := time.Parse(time.DateTime, localResponseTimeStr)
				// Ensure that local time (from user's device clock) is current (within 24 hours of survey response date)
				if err == nil && surveyResponse.DateCreated.Sub(localResponseTime).Abs().Hours() < 24 {
					responseTime = localResponseTime
				} else {
					l.Errorf("Error parsing local time: %v", err)
				}
			}
		}
	}

	if utils.IsPrevOrSameDay(score.PrevSurveyResponseDate, responseTime) {
		// Don't update streak if day is same or previous
	} else if utils.IsNextDay(score.PrevSurveyResponseDate, responseTime) {
		// Update streak
		score.CurrentStreak++
	} else {
		// Reset streak to day 1
		l.Infof("Resetting streak (%d) on %v (%s)", score.CurrentStreak, responseTime, survey.ID)
		score.CurrentStreak = 1
	}
	score.PrevSurveyResponseDate = responseTime

	if score.CurrentStreak >= model.ScoreStreakMinDays {
		score.Score += pointsForResponse * model.ScoreStreakMultiplier
	} else {
		score.Score += pointsForResponse
	}
}

// newAppShared creates new appShared
func newAppShared(app *Application) appShared {
	return appShared{app: app}
}
