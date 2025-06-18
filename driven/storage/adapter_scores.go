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
				"rank": bson.M{"$denseRank": bson.M{}},
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
func (a *Adapter) GetScores(orgID *string, appID *string, limit *int, offset *int, prevSurveyResponseDateMin *time.Time, prevSurveyResponseDateMax *time.Time) ([]model.Score, error) {
	filter := bson.M{
		"external_profile_id": bson.M{
			"$ne": "",
		},
		"score": bson.M{
			"$gt": 0,
		},
	}
	if orgID != nil && *orgID != "" {
		filter["org_id"] = *orgID
	}
	if appID != nil && *appID != "" {
		filter["app_id"] = *appID
	}
	if prevSurveyResponseDateMin != nil {
		filter["prev_survey_response_date"] = bson.M{"$gte": *prevSurveyResponseDateMin}
	}
	if prevSurveyResponseDateMax != nil {
		filter["prev_survey_response_date"] = bson.M{"$lt": *prevSurveyResponseDateMax}
	}

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: filter}},

		// add rank
		bson.D{{Key: "$setWindowFields", Value: bson.M{
			"sortBy": bson.M{"score": -1},
			"output": bson.M{
				"rank": bson.M{"$denseRank": bson.M{}},
			},
		}}},
	}
	if offset != nil {
		pipeline = append(pipeline, bson.D{{Key: "$skip", Value: *offset}})
	}
	if limit != nil {
		pipeline = append(pipeline, bson.D{{Key: "$limit", Value: *limit}})
	}

	var scores []model.Score
	err := a.db.scores.Aggregate(a.context, pipeline, &scores, nil)

	return scores, err
}

// GetScoresFromLeaderboards retrieves scores from specified leaderboards
func (a *Adapter) GetScoresFromLeaderboards(orgID string, appID string, leaderboardIDs []string, userID *string, limit *int, offset *int) ([]model.Score, error) {
	leaderboardEntryFilter := bson.M{
		"org_id": orgID,
		"app_id": appID,
	}
	if leaderboardIDs != nil && len(leaderboardIDs) > 0 {
		leaderboardEntryFilter["leaderboard_id"] = bson.M{
			"$in": leaderboardIDs,
		}
	}

	pipeline := mongo.Pipeline{}
	leaderboardEntryMatch := bson.D{{
		Key: "$match", Value: leaderboardEntryFilter,
	}}
	pipeline = append(pipeline, leaderboardEntryMatch)

	setWindow := bson.D{{
		Key: "$setWindowFields", Value: bson.D{
			{Key: "partitionBy", Value: "$leaderboard_id"},
			{Key: "sortBy", Value: bson.M{"score": -1}},
			{Key: "output", Value: bson.M{
				"rank": bson.M{"$denseRank": bson.M{}},
			}},
		},
	}}
	pipeline = append(pipeline, setWindow)

	if userID != nil {
		filterByUserID := bson.D{bson.E{
			Key: "$match",
			Value: bson.D{
				{Key: "user_id", Value: *userID},
			},
		},}
		pipeline = append(pipeline, filterByUserID)
	}

	groupStage := bson.D{{
		Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$user_id"},
			{Key: "scores", Value: bson.D{
				{Key: "$push", Value: bson.M{
					"_id":                       "$_id",
					"org_id":                    "$org_id",
					"app_id":                    "$app_id",
					"user_id":                   "$user_id",
					"survey_type":               "$survey_type",
					"external_profile_id":       "$external_profile_id",
					"score":                     "$score",
					"response_count":            "$response_count",
					"prev_survey_response_date": "$prev_survey_response_date",
					"current_streak":            "$current_streak",
					"streak_multiplier":         "$streak_multiplier",
					"answer_count":              "$answer_count",
					"correct_answer_count":      "$correct_answer_count",
					"rank":                      "$rank",
				}},
			}},
		}},
	}
	pipeline = append(pipeline, groupStage)

	projectStage := bson.D{{
		Key: "$project", Value: bson.D{
			{Key: "user_id", Value: "$_id"},
			{Key: "scores", Value: 1},
			{Key: "_id", Value: 0},
		},
	}}
	pipeline = append(pipeline, projectStage)

	skipStage := bson.D{{Key: "$skip", Value: offset}}
	limitStage := bson.D{{Key: "$limit", Value: limit}}
	pipeline = append(pipeline, skipStage, limitStage)

	var scores []model.Score
	err := a.db.leaderboardEntries.Aggregate(a.context, pipeline, &scores, nil)

	return scores, err
}

// GetTopAndLocalScores retrieves top and local scores closest to the user's score
func (a *Adapter) GetTopAndLocalScores(orgID string, appID string, userID string, limit *int, offset *int, abovePivotLimit *int, equalPivotLimit *int, belowPivotLimit *int) ([]model.Score, error) {
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
				"rank": bson.M{"$denseRank": bson.M{}},
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
	filters := bson.M{
		"topScores": 1,
		"userScore": 1,
	}

	facets["userScore"] = bson.A{
		bson.D{{Key: "$match", Value: bson.M{"user_id": userID}}},
	}
	concatArrays = append(concatArrays, "$userScore")

	if limit != nil && *limit > 0 {
		facets["topScores"] = bson.A{
			bson.D{{Key: "$skip", Value: *offset}},
			bson.D{{Key: "$limit", Value: *limit}},
		}
		concatArrays = append(concatArrays, "$topScores")
	}

	if abovePivotLimit != nil && *abovePivotLimit > 0 {
		facets["aboveScores"] = bson.A{
			bson.D{{Key: "$match", Value: bson.M{
				"$expr": bson.M{"$gt": bson.A{"$score", "$pivotScore"}},
			}}},
			bson.D{{Key: "$sort", Value: bson.M{"score": 1}}},
			bson.D{{Key: "$limit", Value: *abovePivotLimit}},
			bson.D{{Key: "$sort", Value: bson.M{"score": -1}}},
		}
		concatArrays = append(concatArrays, "$aboveScores")
		filters["aboveScores"] = bson.M{"$filter": bson.M{
			"input": "$aboveScores",
			"as":    "s",
			"cond":  bson.M{"$not": bson.M{"$in": bson.A{"$$s.user_id", "$topScores.user_id"}}},
		}}
	}

	if equalPivotLimit != nil && *equalPivotLimit > 0 {
		facets["equalScores"] = bson.A{
			bson.D{{Key: "$match", Value: bson.M{
				"$expr": bson.M{"$and": bson.A{
					bson.M{"$eq": bson.A{"$score", "$pivotScore"}},
					bson.M{"$ne": bson.A{"$user_id", userID}},
				}},
			}}},
			bson.D{{Key: "$limit", Value: *equalPivotLimit}},
		}
		concatArrays = append(concatArrays, "$equalScores")
		filters["equalScores"] = bson.M{"$filter": bson.M{
			"input": "$equalScores",
			"as":    "s",
			"cond":  bson.M{"$not": bson.M{"$in": bson.A{"$$s.user_id", "$topScores.user_id"}}},
		}}
	}

	if belowPivotLimit != nil && *belowPivotLimit > 0 {
		facets["belowScores"] = bson.A{
			bson.D{{Key: "$match", Value: bson.M{
				"$expr": bson.M{"$lt": bson.A{"$score", "$pivotScore"}},
			}}},
			bson.D{{Key: "$sort", Value: bson.M{"score": -1}}},
			bson.D{{Key: "$limit", Value: *belowPivotLimit}},
		}
		concatArrays = append(concatArrays, "$belowScores")
		filters["belowScores"] = bson.M{"$filter": bson.M{
			"input": "$belowScores",
			"as":    "s",
			"cond":  bson.M{"$not": bson.M{"$in": bson.A{"$$s.user_id", "$topScores.user_id"}}},
		}}
	}

	pipeline = append(pipeline, bson.D{{Key: "$facet", Value: facets}})

	// filter out anything in topScores from the other buckets
	pipeline = append(pipeline, bson.D{{Key: "$project", Value: filters}})

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
