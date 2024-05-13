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

	"github.com/rokwire/logging-library-go/v2/errors"
	"github.com/rokwire/logging-library-go/v2/logutils"
)

// appAnalytics contains analytics implementations
type appAnalytics struct {
	app *Application
}

// Survey Response
// GetAnonymousSurveyResponses returns the anonymized survey responses matching the provided filters
func (a appAnalytics) GetAnonymousSurveyResponses(surveyTypes []string, startDate *time.Time, endDate *time.Time) ([]model.SurveyResponseAnonymous, error) {
	// GetUserSurveyResponses returns the survey responses matching the provided filters
	responses, err := a.app.storage.GetSurveyResponses(nil, nil, nil, nil, surveyTypes, startDate, endDate, nil, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, model.TypeSurveyResponse, nil, err)
	}

	anonResData := make([]model.SurveyResponseAnonymous, len(responses))
	for i, surveyRes := range responses {
		anonResData[i] = model.SurveyResponseAnonymous{ID: surveyRes.Survey.ID, CreatorID: surveyRes.Survey.CreatorID, AppID: surveyRes.Survey.AppID,
			OrgID: surveyRes.Survey.OrgID, Title: surveyRes.Survey.Title, Type: surveyRes.Survey.Type, SurveyStats: surveyRes.Survey.SurveyStats,
			DateCreated: surveyRes.Survey.DateCreated, DateUpdated: surveyRes.Survey.DateUpdated}
	}

	return anonResData, nil
}

// newAppAnalytics creates new appAnalytics
func newAppAnalytics(app *Application) appAnalytics {
	return appAnalytics{app: app}
}
