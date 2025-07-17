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
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetScore finds score object for user
func (a *Adapter) GetScore(orgID string, appID string, userID string) (*model.Score, error) {
	pipeline := mongo.Pipeline{
		// 1. Match the specific user
		bson.D{{Key: "$match", Value: bson.M{
			"org_id":  orgID,
			"app_id":  appID,
			"user_id": userID,
		}}},

		// 2. Lookup to count distinct scores higher than this user's score
		bson.D{{Key: "$lookup", Value: bson.M{
			"from": "scores",
			"let":  bson.M{"userScore": "$score"},
			"pipeline": mongo.Pipeline{
				bson.D{{Key: "$match", Value: bson.M{
					"$expr": bson.M{"$and": bson.A{
						bson.M{"$eq": bson.A{"$org_id", orgID}},
						bson.M{"$eq": bson.A{"$app_id", appID}},
						bson.M{"$ne": bson.A{"$external_profile_id", ""}},
						bson.M{"$gt": bson.A{"$score", "$$userScore"}},
					}},
				}}},
				bson.D{{Key: "$group", Value: bson.M{
					"_id": "$score", // Group by distinct scores
				}}},
				bson.D{{Key: "$count", Value: "distinct_higher_scores"}},
			},
			"as": "rank_data",
		}}},

		// 3. Add the rank field
		bson.D{{Key: "$addFields", Value: bson.M{
			"rank": bson.M{
				"$add": bson.A{
					1, // Base rank is 1
					bson.M{"$ifNull": bson.A{
						bson.M{"$arrayElemAt": bson.A{"$rank_data.distinct_higher_scores", 0}},
						0,
					}},
				},
			},
		}}},

		// 4. Remove the temporary rank_data field
		bson.D{{Key: "$unset", Value: "rank_data"}},
	}

	var scores []model.Score
	err := a.db.scores.Aggregate(a.context, pipeline, &scores, nil)
	if err != nil {
		return nil, err
	}

	if len(scores) == 0 {
		return nil, errors.ErrorData(logutils.StatusMissing, model.TypeScore, &logutils.FieldArgs{"user_id": userID})
	}

	return &scores[0], nil
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

	// Use Find with sort and pagination options - much simpler than aggregation pipeline
	findOptions := &options.FindOptions{}
	findOptions.SetSort(bson.M{"score": -1}) // Sort by score descending

	if offset != nil {
		findOptions.SetSkip(int64(*offset))
	}
	if limit != nil {
		findOptions.SetLimit(int64(*limit))
	}

	var scores []model.Score
	err := a.db.scores.Find(a.context, filter, &scores, findOptions)
	if err != nil {
		return nil, err
	}

	// If no scores found, return empty slice
	if len(scores) == 0 {
		return scores, nil
	}

	if appID != nil && orgID != nil {
		err = a.rankScoreList(scores, *orgID, *appID, offset)
		if err != nil {
			return nil, err
		}
	}

	return scores, nil
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

	err = a.rankScoreList(scores, orgID, appID, offset)
	if err != nil {
		return nil, err
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
		if len(aboveScores) > 0 || len(belowScores) > 0 || !userInTopScores {
			userScoreWindow := make([]model.Score, 0, len(aboveScores)+1+len(belowScores))
			userScoreWindow = append(userScoreWindow, aboveScores...)
			if !userInTopScores {
				userScoreWindow = append(userScoreWindow, *userScore)
			}
			userScoreWindow = append(userScoreWindow, belowScores...)

			// Rank user score window

			// If above scores are not full then last top score is ranked at/above first score in window
			// No need to find rank for first score in window in this case
			if hasAboveLimit && len(scores) > 0 && len(aboveScores) < *abovePivotLimit {
				prevScore := scores[len(scores)-1]
				a.setScoreListRanks(userScoreWindow, prevScore.Rank, &prevScore.Score)
			} else {
				// Otherwise we need to find the rank for the first score in the window
				fakeOffset := 1 // Fake an offset > 0to trigger find rank for first score
				a.rankScoreList(userScoreWindow, orgID, appID, &fakeOffset)
			}

			// Append user score window below scores
			scores = append(scores, userScoreWindow...)
		}
	}

	return scores, nil
}

// CreateScore creates a new score object
func (a *Adapter) CreateScore(score model.Score) error {
	// Acquire read lock to allow concurrent score operations but block during rank initialization
	a.ranksLock.RLock()
	defer a.ranksLock.RUnlock()

	_, err := a.db.scores.InsertOne(a.context, score)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionCreate, model.TypeScore, nil, err)
	}
	return nil
}

// UpdateScore updates existing score object
func (a *Adapter) UpdateScore(score model.Score) error {
	// Acquire read lock to allow concurrent score operations but block during rank initialization
	a.ranksLock.RLock()
	defer a.ranksLock.RUnlock()

	if len(score.ID) == 0 {
		return nil
	}

	filter := bson.M{"_id": score.ID, "org_id": score.OrgID, "app_id": score.AppID, "user_id": score.UserID}

	setUpdate := bson.M{
		"score":                     score.Score,
		"response_count":            score.ResponseCount,
		"prev_survey_response_date": score.PrevSurveyResponseDate,
		"current_streak":            score.CurrentStreak,
		"answer_count":              score.AnswerCount,
		"correct_answer_count":      score.CorrectAnswerCount,
	}
	if score.ExternalProfileID != "" {
		setUpdate["external_profile_id"] = score.ExternalProfileID
	}
	update := bson.M{"$set": setUpdate}

	res, err := a.db.scores.UpdateOne(a.context, filter, update, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeScore, filterArgs(filter), err)
	}
	if res.ModifiedCount != 1 {
		return errors.WrapErrorData(logutils.StatusMissing, model.TypeScore, filterArgs(filter), err)
	}

	return nil
}

// Helpers

// rankScoreList ranks a list of scores by counting distinct scores higher than each score
func (a *Adapter) rankScoreList(scores []model.Score, orgID string, appID string, offset *int) error {
	if len(scores) == 0 {
		return nil
	}

	var err error
	firstScoreRank := uint32(1)

	// If there is an offset, we need to compute the rank for the first score
	if offset != nil && *offset > 0 {
		// Compute the rank for the first (highest) score by counting distinct scores higher than it
		firstScore := scores[0].Score
		firstScoreRank, err = a.findRankForScore(firstScore, orgID, appID)
		if err != nil {
			return err
		}
	}

	a.setScoreListRanks(scores, firstScoreRank, nil)

	return nil
}

// setScoreListRanks sets the rank for a list of scores
func (a *Adapter) setScoreListRanks(scores []model.Score, firstScoreRank uint32, prevScore *float64) {
	if len(scores) == 0 {
		return
	}

	// Assign ranks programmatically based on score changes
	currentRank := firstScoreRank
	if prevScore != nil && *prevScore != scores[0].Score {
		currentRank++
	}
	scores[0].Rank = currentRank

	for i := 1; i < len(scores); i++ {
		// If the score is different from the previous score, increment rank
		if scores[i].Score != scores[i-1].Score {
			currentRank++
		}
		scores[i].Rank = currentRank
	}
}

// findRankForScore finds the rank for a specific score by counting distinct scores higher than it
func (a *Adapter) findRankForScore(score float64, orgID string, appID string) (uint32, error) {
	filter := bson.M{
		"org_id": orgID,
		"app_id": appID,
		"external_profile_id": bson.M{
			"$ne": "",
		},
		"score": bson.M{
			"$gt": score, // Only scores higher than the given score
		},
	}

	pipeline := mongo.Pipeline{
		// Match scores higher than the given score with same filter criteria
		bson.D{{Key: "$match", Value: filter}},
		// Group by distinct scores to count them
		bson.D{{Key: "$group", Value: bson.M{
			"_id": "$score", // Group by distinct scores
		}}},
		// Count the distinct higher scores
		bson.D{{Key: "$count", Value: "distinct_higher_scores"}},
	}

	var result []bson.M
	err := a.db.scores.Aggregate(a.context, pipeline, &result, nil)
	if err != nil {
		return 0, errors.WrapErrorAction(logutils.ActionFind, model.TypeRank, &logutils.FieldArgs{"score": score}, err)
	}

	// If no higher scores found, rank is 1
	if len(result) == 0 {
		return 1, nil
	}

	// Rank = 1 + count of distinct higher scores
	count, ok := result[0]["distinct_higher_scores"].(int32)
	if !ok {
		return 1, nil
	}

	return uint32(count) + 1, nil
}
