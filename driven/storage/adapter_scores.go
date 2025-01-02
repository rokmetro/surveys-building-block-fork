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

	"github.com/rokwire/logging-library-go/v2/errors"
	"github.com/rokwire/logging-library-go/v2/logutils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (a *Adapter) GetScore(orgID string, appID string, userID string) (*model.Score, error) {
	filter := bson.M{"org_id": orgID, "app_id": appID, "user_id": userID}
	var entry model.Score
	err := a.db.scores.FindOne(a.context, filter, &entry, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeScore, filterArgs(filter), err)
	}
	return &entry, nil
}

func (a *Adapter) GetScores(orgID string, appID string, limit *int, offset *int) ([]model.Score, error) {
	filter := bson.M{"org_id": orgID, "app_id": appID}

	opts := options.Find().SetSort(bson.M{"score": -1})
	if limit != nil {
		opts.SetLimit(int64(*limit))
	}
	if offset != nil {
		opts.SetSkip(int64(*offset))
	}
	var results []model.Score
	err := a.db.scores.Find(a.context, filter, &results, opts)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeScore, filterArgs(filter), err)
	}
	return results, nil
}

func (a *Adapter) CreateScore(score model.Score) error {
	_, err := a.db.scores.InsertOne(a.context, score)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionCreate, model.TypeScore, nil, err)
	}
	return nil
}

func (a *Adapter) UpdateScore(score model.Score) error {
	if len(score.ID) == 0 {
		return nil
	}

	filter := bson.M{"_id": score.ID, "org_id": score.OrgID, "app_id": score.AppID, "user_id": score.UserID}

	update := bson.M{"$set": bson.M{
		"external_profile_id":       score.ExternalProfileID,
		"score":                     score.Score,
		"response_count":            score.ResponseCount,
		"prev_survey_response_date": score.PrevSurveyResponseDate,
		"current_streak":            score.CurrentStreak,
		"answer_count":              score.AnswerCount,
		"correct_answer_count":      score.CorrectAnswerCount,
	}}

	res, err := a.db.scores.UpdateOne(a.context, filter, update, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeScore, filterArgs(filter), err)
	}
	if res.ModifiedCount != 1 {
		return errors.WrapErrorData(logutils.StatusMissing, model.TypeScore, filterArgs(filter), err)
	}

	return nil
}
