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
	"slices"
	"time"

	"github.com/rokwire/rokwire-building-block-sdk-go/utils/errors"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logs"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logutils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetScore finds score object for user
func (a *Adapter) GetScore(orgID string, appID string, userID string) (*model.Score, error) {
	filter := bson.M{
		"org_id":  orgID,
		"app_id":  appID,
		"user_id": userID,
	}

	var score model.Score
	err := a.db.scores.FindOne(a.context, filter, &score, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeScore, filterArgs(filter), err)
	}

	return &score, nil
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

	opts := options.Find()
	opts.SetSort(bson.M{"score": -1})
	if limit != nil {
		opts.SetLimit(int64(*limit))
	}
	if offset != nil {
		opts.SetSkip(int64(*offset))
	}

	var scores []model.Score
	err := a.db.scores.Find(a.context, filter, &scores, opts)

	return scores, err
}

// GetTopAndLocalScores retrieves top and local scores closest to the user's score
func (a *Adapter) GetTopAndLocalScores(orgID string, appID string, userID string, limit *int, offset *int, abovePivotLimit *int, equalPivotLimit *int, belowPivotLimit *int, l *logs.Log) ([]model.Score, error) {
	// Get top scores
	scoreFilter := bson.M{
		"org_id": orgID,
		"app_id": appID,
		"external_profile_id": bson.M{
			"$ne": "",
		},
	}

	opts := options.Find()
	opts.SetSort(bson.M{"score": -1})
	if limit != nil {
		opts.SetLimit(int64(*limit))
	}
	if offset != nil {
		opts.SetSkip(int64(*offset))
	}

	var scores []model.Score
	err := a.db.scores.Find(a.context, scoreFilter, &scores, opts)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeScore, filterArgs(scoreFilter), err)
	}
	if len(scores) == 0 {
		return scores, nil
	}

	hasAboveLimit := abovePivotLimit != nil && *abovePivotLimit > 0
	hasEqualLimit := equalPivotLimit != nil && *equalPivotLimit > 0
	hasBelowLimit := belowPivotLimit != nil && *belowPivotLimit > 0

	missingAboveScores := 0
	missingBelowScores := 0

	// Check if user is in top scores
	var userScore *model.Score
	userInTopScores := false
	topUserIDs := make([]string, len(scores))
	for i, score := range scores {
		topUserIDs[i] = score.UserID
		if score.UserID == userID {
			userScore = &score
			userInTopScores = true
			if hasBelowLimit {
				missingBelowScores = *belowPivotLimit - (len(scores) - i - 1)
			}
		}
	}

	// Load user's score if not in top scores
	if !userInTopScores {
		score, err := a.GetScore(orgID, appID, userID)
		if err != nil {
			l.Warnf("failed to find score for user: %v", err)
		}
		topUserIDs = append(topUserIDs, userID)
		userScore = score
		if hasBelowLimit {
			missingBelowScores = *belowPivotLimit
		}
		// Check if user's score is below top scores (not equal)
		if hasAboveLimit && userScore != nil && userScore.Score < scores[len(scores)-1].Score {
			missingAboveScores = *abovePivotLimit
		}
	}

	// Load scores around user's score
	if userScore != nil {
		scoreFilter := bson.M{
			"org_id": orgID,
			"app_id": appID,
			"external_profile_id": bson.M{
				"$ne": "",
			},
			"user_id": bson.M{
				"$not": bson.M{"$in": topUserIDs},
			},
		}

		scoreOpts := options.Find()
		scoreOpts.SetSort(bson.M{"score": -1})

		aboveScores := []model.Score{}
		belowScores := []model.Score{}

		// Load equal scores
		if (missingAboveScores > 0 || missingBelowScores > 0) && hasEqualLimit {
			scoreFilter["score"] = userScore.Score
			scoreOpts.SetLimit(int64(*equalPivotLimit))

			var equalScores []model.Score
			err = a.db.scores.Find(a.context, scoreFilter, &equalScores, scoreOpts)
			if err != nil {
				l.Warnf("failed to find equal scores for user: %v", err)
			}

			if len(equalScores) > 0 {
				// Split equal scores in half centered around user
				splitIdx := len(equalScores) / 2
				aboveScores = equalScores[:splitIdx]
				belowScores = equalScores[splitIdx:]

				// Cap scores to missing score lengths
				if len(aboveScores) > missingAboveScores {
					aboveScores = aboveScores[:missingAboveScores]
				}
				if len(belowScores) > missingBelowScores {
					belowScores = belowScores[:missingBelowScores]
				}

				// Update missing scores
				missingAboveScores -= len(aboveScores)
				missingBelowScores -= len(belowScores)
			}
		}

		// Load above scores if still needed
		if missingAboveScores > 0 {
			scoreFilter["score"] = bson.M{
				"$gt": userScore.Score,
			}
			// Sort by score ascending to get lowest scores first
			scoreOpts.SetSort(bson.M{"score": 1})
			scoreOpts.SetLimit(int64(missingAboveScores))

			var foundAboveScores []model.Score
			err := a.db.scores.Find(a.context, scoreFilter, &foundAboveScores, scoreOpts)
			if err != nil {
				l.Warnf("failed to find above scores for user: %v", err)
			}
			// Reverse scores to get them in descending order
			slices.Reverse(foundAboveScores)
			// Prepend newly found above scores before equal scores above user
			aboveScores = append(foundAboveScores, aboveScores...)
		}

		// Load below scores if still needed
		if missingBelowScores > 0 {
			scoreFilter["score"] = bson.M{
				"$lt": userScore.Score,
			}
			// Sort by score descending to get highest scores first
			scoreOpts.SetSort(bson.M{"score": -1})
			scoreOpts.SetLimit(int64(missingBelowScores))

			var foundBelowScores []model.Score
			err := a.db.scores.Find(a.context, scoreFilter, &foundBelowScores, scoreOpts)
			if err != nil {
				l.Warnf("failed to find below scores for user: %v", err)
			}
			// Append newly found below scores after equal scores below user
			belowScores = append(belowScores, foundBelowScores...)
		}

		// Append user score window to scores
		scores = append(scores, aboveScores...)
		if !userInTopScores {
			scores = append(scores, *userScore)
		}
		scores = append(scores, belowScores...)
	}

	return scores, nil
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
