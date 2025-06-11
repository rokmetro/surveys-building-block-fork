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

	"github.com/rokwire/rokwire-building-block-sdk-go/utils/errors"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logutils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetScore finds score object for user
func (a *Adapter) GetScore(orgID string, appID string, userID string) (*model.Score, error) {
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.M{"org_id": orgID, "app_id": appID}}},
		// add rank
		bson.D{{Key: "$setWindowFields", Value: bson.M{
			"sortBy": bson.M{"score": -1},
			"output": bson.M{
				"rank": bson.M{"$rank": bson.M{}},
			},
		}}},
		bson.D{{Key: "$match", Value: bson.M{"user_id": userID}}},
	}

	var scores []model.Score
	err := a.db.scores.Aggregate(a.context, pipeline, &scores, nil)
	if len(scores) > 0 {
		score := scores[0]
		return &score, err
	}
	return nil, errors.ErrorData(logutils.StatusMissing, model.TypeScore, &logutils.FieldArgs{"user_id": userID})
}

// GetScores returns a list of scores in descending order
func (a *Adapter) GetScores(orgID string, appID string, limit *int, offset *int) ([]model.Score, error) {
	filter := bson.M{
		"org_id": orgID,
		"app_id": appID,
		"external_profile_id": bson.M{
			"$ne": "",
		},
		"score": bson.M{
			"$gt": 0,
		},
	}

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: filter}},

		// add rank
		bson.D{{Key: "$setWindowFields", Value: bson.M{
			"sortBy": bson.M{"score": -1},
			"output": bson.M{
				"rank": bson.M{"$rank": bson.M{}},
			},
		}}},

		bson.D{{Key: "$skip", Value: *offset}},
		bson.D{{Key: "$limit", Value: *limit}},
	}

	var scores []model.Score
	err := a.db.scores.Aggregate(a.context, pipeline, &scores, nil)

	return scores, err
}

// GetScoresWithPivot retrieves scores closest to the user's score
func (a *Adapter) GetScoresWithPivot(orgID string, appID string, userID string, aboveLimit *int, equalLimit *int, belowLimit *int) ([]model.Score, error) {
	scoreFilter := bson.M{
		"org_id": orgID,
		"app_id": appID,
		"external_profile_id": bson.M{
			"$ne": "",
		},
	}

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: scoreFilter}},

		// Add rank + pivotScore via window
		bson.D{{Key: "$setWindowFields", Value: bson.M{
			"sortBy": bson.M{"score": -1},
			"output": bson.M{
				"rank": bson.M{"$rank": bson.M{}},
				"pivotScore": bson.M{"$max": bson.M{
					"$cond": bson.A{
						bson.M{"$eq": bson.A{"$user_id", userID}},
						"$score",
						nil,
					},
				}},
			},
		}}},
	}

	// facet into above/equal/below
	// build the facet map only for non‐nil, positive limits
	facets := bson.M{}
	concatArrays := bson.A{}

	if aboveLimit != nil && *aboveLimit > 0 {
		facets["aboveScores"] = bson.A{
			bson.D{{Key: "$match", Value: bson.M{
				"$expr": bson.M{"$gt": bson.A{"$score", "$pivotScore"}},
			}}},
			bson.D{{Key: "$sort", Value: bson.M{"score": 1}}},
			bson.D{{Key: "$limit", Value: *aboveLimit}},
			bson.D{{Key: "$sort", Value: bson.M{"score": -1}}},
		}
		concatArrays = append(concatArrays, "$aboveScores")
	}

	if equalLimit != nil && *equalLimit > 0 {
		facets["equalScores"] = bson.A{
			bson.D{{Key: "$match", Value: bson.M{
				"$expr": bson.M{"$and": bson.A{
					bson.M{"$eq": bson.A{"$score", "$pivotScore"}},
					bson.M{"$ne": bson.A{"$user_id", userID}},
				}},
			}}},
			bson.D{{Key: "$limit", Value: *equalLimit}},
		}
		concatArrays = append(concatArrays, "$equalScores")
	}

	if belowLimit != nil && *belowLimit > 0 {
		facets["belowScores"] = bson.A{
			bson.D{{Key: "$match", Value: bson.M{
				"$expr": bson.M{"$lt": bson.A{"$score", "$pivotScore"}},
			}}},
			bson.D{{Key: "$sort", Value: bson.M{"score": -1}}},
			bson.D{{Key: "$limit", Value: *belowLimit}},
		}
		concatArrays = append(concatArrays, "$belowScores")
	}

	// only append a $facet if we have at least one bucket
	if len(facets) > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$facet", Value: facets}})
	}

	// Combine the three facets into one sorted array.
	projectFacetStage := bson.D{{Key: "$project", Value: bson.D{
		{Key: "results", Value: bson.D{
			{Key: "$concatArrays", Value: concatArrays},
		}},
	}}}
	pipeline = append(pipeline, projectFacetStage)

	// Unwind the concatenated array and set each element as the new root.
	unwindStage := bson.D{{Key: "$unwind", Value: "$results"}}
	replaceRootStage := bson.D{{Key: "$replaceRoot", Value: bson.D{{Key: "newRoot", Value: "$results"}}}}
	pipeline = append(pipeline, unwindStage, replaceRootStage)

	var scores []model.Score
	err := a.db.scores.Aggregate(a.context, pipeline, &scores, nil)

	return scores, err
}

// CreateScore creates a new score object
func (a *Adapter) CreateScore(score model.Score) error {
	_, err := a.db.scores.InsertOne(a.context, score)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionCreate, model.TypeScore, nil, err)
	}
	return nil
}

// UpdateScore updates existing score object
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
