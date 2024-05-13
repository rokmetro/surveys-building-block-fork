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

package storage

import (
	"application/core/model"
	"time"

	"github.com/rokwire/logging-library-go/v2/errors"
	"github.com/rokwire/logging-library-go/v2/logutils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetSurvey retrieves a single survey
func (a *Adapter) GetSurvey(id string, orgID string, appID string) (*model.Survey, error) {
	filter := bson.M{"_id": id, "org_id": orgID, "app_id": appID}
	var entry model.Survey
	err := a.db.surveys.FindOne(a.context, filter, &entry, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeSurvey, filterArgs(filter), err)
	}
	return &entry, nil
}

// GetSurveys gets matching surveys
func (a *Adapter) GetSurveys(orgID string, appID string, creatorID *string, surveyIDs []string, surveyTypes []string, calendarEventID string, limit *int, offset *int) ([]model.Survey, error) {
	filter := bson.M{"org_id": orgID, "app_id": appID}
	if creatorID != nil {
		filter["creator_id"] = *creatorID
	}
	if len(surveyIDs) > 0 {
		filter["_id"] = bson.M{"$in": surveyIDs}
	}
	if len(surveyTypes) > 0 {
		filter["type"] = bson.M{"$in": surveyTypes}
	}
	if len(calendarEventID) > 0 {
		filter["calendar_event_id"] = calendarEventID
	}

	opts := options.Find()
	if limit != nil {
		opts.SetLimit(int64(*limit))
	}
	if offset != nil {
		opts.SetSkip(int64(*offset))
	}
	var results []model.Survey
	err := a.db.surveys.Find(a.context, filter, &results, opts)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeSurvey, filterArgs(filter), err)
	}
	return results, nil
}

// CreateSurvey creates a poll
func (a *Adapter) CreateSurvey(survey model.Survey) (*model.Survey, error) {
	_, err := a.db.surveys.InsertOne(a.context, survey)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionCreate, model.TypeSurvey, nil, err)
	}

	return &survey, nil
}

// UpdateSurvey updates a survey
func (a *Adapter) UpdateSurvey(survey model.Survey, admin bool) error {
	if len(survey.ID) > 0 {
		now := time.Now().UTC()
		filter := bson.M{"_id": survey.ID, "org_id": survey.OrgID, "app_id": survey.AppID}
		if !admin {
			filter["creator_id"] = survey.CreatorID
		}
		update := bson.M{"$set": bson.M{
			"title":                 survey.Title,
			"more_info":             survey.MoreInfo,
			"data":                  survey.Data,
			"scored":                survey.Scored,
			"result_rules":          survey.ResultRules,
			"type":                  survey.Type,
			"stats":                 survey.SurveyStats,
			"default_data_key":      survey.DefaultDataKey,
			"default_data_key_rule": survey.DefaultDataKeyRule,
			"constants":             survey.Constants,
			"strings":               survey.Strings,
			"sub_rules":             survey.SubRules,
			"date_updated":          now,
		}}

		res, err := a.db.surveys.UpdateOne(a.context, filter, update, nil)
		if err != nil {
			return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeSurvey, filterArgs(filter), err)
		}
		if res.ModifiedCount != 1 {
			return errors.WrapErrorData(logutils.StatusMissing, model.TypeSurvey, filterArgs(filter), err)
		}
	}

	return nil
}

// DeleteSurvey deletes a survey
func (a *Adapter) DeleteSurvey(id string, orgID string, appID string, creatorID string, admin bool) error {
	filter := bson.M{"_id": id, "org_id": orgID, "app_id": appID}
	if !admin {
		filter["creator_id"] = creatorID
	}
	res, err := a.db.surveys.DeleteOne(a.context, filter, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionDelete, model.TypeSurvey, filterArgs(filter), err)
	}
	if res.DeletedCount != 1 {
		return errors.WrapErrorData(logutils.StatusMissing, model.TypeSurvey, filterArgs(filter), err)
	}

	return nil
}
