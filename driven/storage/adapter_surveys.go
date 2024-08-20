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
	"go.mongodb.org/mongo-driver/bson/primitive"
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
func (a *Adapter) GetSurveys(orgID string, appID string, creatorID *string, surveyIDs []string, surveyTypes []string, calendarEventID string, limit *int, offset *int, timeFilter *model.SurveyTimeFilter, public *bool, archived *bool, completed *bool) ([]model.Survey, error) {
	filter := bson.D{
		{Key: "org_id", Value: orgID},
		{Key: "app_id", Value: appID},
	}

	if creatorID != nil {
		filter = append(filter, bson.E{Key: "creator_id", Value: *creatorID})
	}
	if len(surveyIDs) > 0 {
		filter = append(filter, bson.E{Key: "_id", Value: bson.M{"$in": surveyIDs}})
	}
	if len(surveyTypes) > 0 {
		filter = append(filter, bson.E{Key: "type", Value: bson.M{"$in": surveyTypes}})
	}
	if calendarEventID != "" {
		filter = append(filter, bson.E{Key: "calendar_event_id", Value: calendarEventID})
	}

	if timeFilter.StartTimeAfter != nil {
		filter = append(filter, primitive.E{Key: "$or", Value: bson.A{
			bson.M{"start_date": nil},
			bson.M{"start_date": primitive.M{"$gte": *timeFilter.StartTimeAfter}},
		}})
	}
	if timeFilter.StartTimeBefore != nil {
		filter = append(filter, primitive.E{Key: "$or", Value: bson.A{
			bson.M{"start_date": nil},
			bson.M{"start_date": primitive.M{"$lte": *timeFilter.StartTimeBefore}},
		}})
	}

	if timeFilter.EndTimeAfter != nil {
		filter = append(filter, primitive.E{Key: "$or", Value: bson.A{
			bson.M{"end_date": nil},
			bson.M{"end_date": primitive.M{"$gte": *timeFilter.EndTimeAfter}},
		}})
	}
	if timeFilter.EndTimeBefore != nil {
		filter = append(filter, primitive.E{Key: "$or", Value: bson.A{
			bson.M{"end_date": nil},
			bson.M{"end_date": primitive.M{"$lte": *timeFilter.EndTimeBefore}},
		}})
	}

	if public != nil {
		if *public == true {
			filter = append(filter, bson.E{Key: "public", Value: true})
		} else {
			filter = append(filter, bson.E{Key: "$or", Value: bson.A{
				bson.M{"public": false},
				bson.M{"public": bson.M{"$exists": false}},
				bson.M{"public": nil},
			}})
		}
	}

	if archived != nil {
		if *archived == true {
			filter = append(filter, bson.E{Key: "archived", Value: true})
		} else {
			filter = append(filter, bson.E{Key: "$or", Value: bson.A{
				bson.M{"archived": false},
				bson.M{"archived": bson.M{"$exists": false}},
				bson.M{"archived": nil},
			}})
		}
	}

	opts := options.Find()
	if limit != nil {
		opts.SetLimit(int64(*limit))
	}
	if offset != nil {
		opts.SetSkip(int64(*offset))
	}
	if timeFilter.StartTimeBefore != nil {
		opts.SetSort(bson.D{{Key: "start_date", Value: -1}})
	} else if timeFilter.StartTimeAfter != nil {
		opts.SetSort(bson.D{{Key: "start_date", Value: 1}})
	}

	if timeFilter.EndTimeBefore != nil {
		opts.SetSort(bson.D{{Key: "end_date", Value: -1}})
	} else if timeFilter.EndTimeAfter != nil {
		opts.SetSort(bson.D{{Key: "end_date", Value: 1}})

	}

	var results []model.Survey
	err := a.db.surveys.Find(a.context, filter, &results, opts)
	if err != nil {
		return nil, err
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
			"title":                     survey.Title,
			"more_info":                 survey.MoreInfo,
			"data":                      survey.Data,
			"scored":                    survey.Scored,
			"result_rules":              survey.ResultRules,
			"type":                      survey.Type,
			"stats":                     survey.SurveyStats,
			"default_data_key":          survey.DefaultDataKey,
			"default_data_key_rule":     survey.DefaultDataKeyRule,
			"constants":                 survey.Constants,
			"strings":                   survey.Strings,
			"sub_rules":                 survey.SubRules,
			"start_date":                survey.StartDate,
			"end_date":                  survey.EndDate,
			"public":                    survey.Public,
			"archived":                  survey.Archived,
			"estimated_completion_time": survey.EstimatedCompletionTime,
			"date_updated":              now,
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
