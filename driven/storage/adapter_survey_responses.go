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

// GetSurveyResponse gets a survey response by ID
func (a *Adapter) GetSurveyResponse(id string, orgID string, appID string, userID string) (*model.SurveyResponse, error) {
	filter := bson.M{"_id": id, "user_id": userID, "org_id": orgID, "app_id": appID}
	var entry model.SurveyResponse
	err := a.db.surveyResponses.FindOne(a.context, filter, &entry, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeSurveyResponse, filterArgs(filter), err)
	}
	return &entry, nil
}

// GetSurveyResponses gets matching surveys for a user
func (a *Adapter) GetSurveyResponses(orgID *string, appID *string, userID *string, surveyIDs []string, surveyTypes []string, startDate *time.Time, endDate *time.Time, limit *int, offset *int) ([]model.SurveyResponse, error) {
	filter := bson.M{}
	if userID != nil {
		filter["user_id"] = userID
	}
	if orgID != nil {
		filter["org_id"] = orgID
	}
	if appID != nil {
		filter["app_id"] = appID
	}

	if len(surveyIDs) > 0 {
		filter["survey._id"] = bson.M{"$in": surveyIDs}
	}
	if len(surveyTypes) > 0 {
		filter["survey.type"] = bson.M{"$in": surveyTypes}
	}
	if startDate != nil || endDate != nil {
		dateFilter := bson.M{}
		if startDate != nil {
			dateFilter["$gte"] = startDate
		}
		if endDate != nil {
			dateFilter["$lt"] = endDate
		}
		filter["date_created"] = dateFilter
	}

	opts := options.Find().SetSort(bson.M{"date_created": -1})
	if limit != nil {
		opts.SetLimit(int64(*limit))
	}
	if offset != nil {
		opts.SetSkip(int64(*offset))
	}
	var results []model.SurveyResponse
	err := a.db.surveyResponses.Find(a.context, filter, &results, opts)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeSurveyResponse, filterArgs(filter), err)
	}
	return results, nil
}

// CreateSurveyResponse creates a new survey response
func (a *Adapter) CreateSurveyResponse(surveyResponse model.SurveyResponse) (*model.SurveyResponse, error) {
	_, err := a.db.surveyResponses.InsertOne(a.context, surveyResponse)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionCreate, model.TypeSurveyResponse, nil, err)
	}
	return &surveyResponse, nil
}

// UpdateSurveyResponse updates an existing service response
func (a *Adapter) UpdateSurveyResponse(surveyResponse model.SurveyResponse) error {
	now := time.Now().UTC()
	filter := bson.M{"_id": surveyResponse.ID, "user_id": surveyResponse.UserID, "org_id": surveyResponse.OrgID, "app_id": surveyResponse.AppID}
	update := bson.M{"$set": bson.M{
		"survey":       surveyResponse.Survey,
		"date_updated": now,
	}}

	res, err := a.db.surveyResponses.UpdateOne(a.context, filter, update, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeSurveyResponse, filterArgs(filter), err)
	}
	if res.ModifiedCount != 1 {
		return errors.WrapErrorData(logutils.StatusMissing, model.TypeSurveyResponse, filterArgs(filter), err)
	}
	return nil
}

// DeleteSurveyResponse deletes a survey response
func (a *Adapter) DeleteSurveyResponse(orgID string, appID string, userID string, id string) error {
	filter := bson.M{"_id": id, "user_id": userID, "org_id": orgID, "app_id": appID}
	res, err := a.db.surveyResponses.DeleteOne(a.context, filter, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionDelete, model.TypeSurveyResponse, filterArgs(filter), err)
	}
	if res.DeletedCount != 1 {
		return errors.WrapErrorData(logutils.StatusMissing, model.TypeSurveyResponse, filterArgs(filter), err)
	}
	return nil
}

// DeleteSurveyResponses deletes matching surveys
func (a *Adapter) DeleteSurveyResponses(orgID string, appID string, userID string, surveyIDs []string, surveyTypes []string, startDate *time.Time, endDate *time.Time) error {
	filter := bson.M{"user_id": userID, "org_id": orgID, "app_id": appID}
	if len(surveyIDs) > 0 {
		filter["survey._id"] = bson.M{"$in": surveyIDs}
	}
	if len(surveyTypes) > 0 {
		filter["survey.type"] = bson.M{"$in": surveyTypes}
	}
	if startDate != nil || endDate != nil {
		dateFilter := bson.M{}
		if startDate != nil {
			dateFilter["$gte"] = startDate
		}
		if endDate != nil {
			dateFilter["$lt"] = endDate
		}
		filter["date_created"] = dateFilter
	}

	result, err := a.db.surveyResponses.DeleteMany(a.context, filter, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionDelete, model.TypeSurveyResponse, filterArgs(filter), err)
	}
	if result.DeletedCount == 0 {
		return errors.WrapErrorData(logutils.StatusMissing, model.TypeSurveyResponse, filterArgs(filter), err)
	}
	return nil
}
