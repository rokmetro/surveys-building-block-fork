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

import "application/core/model"

// Shared exposes shared APIs for other interface implementations
type Shared interface {
	// Surveys
	getSurvey(id string, orgID string, appID string) (*model.Survey, error)
	getSurveys(orgID string, appID string, userID *string, creatorID *string, surveyIDs []string, surveyTypes []string, calendarEventID string, limit *int, offset *int, filter *model.SurveyTimeFilter, public *bool, archived *bool, completed *bool) ([]model.Survey, []model.SurveyResponse, error)
	createSurvey(survey model.Survey, externalIDs map[string]string) (*model.Survey, error)
	updateSurvey(survey model.Survey, userID string, externalIDs map[string]string, admin bool) error
	deleteSurvey(id string, orgID string, appID string, userID string, externalIDs map[string]string, admin bool) error

	isEventAdmin(orgID string, appID string, eventID string, userID string, externalIDs map[string]string) (bool, error)
	hasAttendedEvent(orgID string, appID string, eventID string, userID string, externalIDs map[string]string) (bool, error)
}

// Core exposes Core APIs for the driver adapters
type Core interface {
	LoadDeletedMemberships() ([]model.DeletedUserData, error)
}
