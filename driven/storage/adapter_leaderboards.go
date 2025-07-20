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
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetLeaderboard gets a leaderboard by ID
func (a *Adapter) GetLeaderboard(leaderboardID string, orgID string, appID string) (*model.Leaderboard, error) {
	filter := bson.M{
		"_id":    leaderboardID,
		"org_id": orgID,
		"app_id": appID,
	}

	var leaderboard model.Leaderboard
	err := a.db.leaderboards.FindOne(a.context, filter, &leaderboard, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeLeaderboard, filterArgs(filter), err)
	}

	return &leaderboard, nil
}

// GetLeaderboardWithUserContext gets a leaderboard by ID and populates the is_admin field for a specific user
func (a *Adapter) GetLeaderboardWithUserContext(leaderboardID string, orgID string, appID string, userID string) (*model.Leaderboard, error) {
	leaderboardFilter := bson.M{
		"_id":    leaderboardID,
		"org_id": orgID,
		"app_id": appID,
	}

	pipeline := mongo.Pipeline{
		// 1) filter leaderboards by ID, org, and app
		bson.D{{Key: "$match", Value: leaderboardFilter}},
		// 2) lookup the leaderboard entry for this specific user
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "leaderboard_entries"},
			{Key: "let", Value: bson.D{{Key: "leaderboard_id", Value: "$_id"}}},
			{Key: "pipeline", Value: mongo.Pipeline{
				bson.D{{Key: "$match", Value: bson.D{
					{Key: "$expr", Value: bson.D{
						{Key: "$and", Value: bson.A{
							bson.D{{Key: "$eq", Value: bson.A{"$leaderboard_id", "$$leaderboard_id"}}},
							bson.D{{Key: "$eq", Value: bson.A{"$user_id", userID}}},
							bson.D{{Key: "$eq", Value: bson.A{"$org_id", orgID}}},
							bson.D{{Key: "$eq", Value: bson.A{"$app_id", appID}}},
						}},
					}},
				}}},
			}},
			{Key: "as", Value: "user_entry"},
		}}},
		// 3) add the is_admin field from the user's entry (if it exists)
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "is_admin", Value: bson.D{
				{Key: "$cond", Value: bson.D{
					{Key: "if", Value: bson.D{{Key: "$gt", Value: bson.A{bson.D{{Key: "$size", Value: "$user_entry"}}, 0}}}},
					{Key: "then", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$user_entry.is_admin", 0}}}},
					{Key: "else", Value: nil},
				}},
			}},
		}}},
		// 4) remove the temporary user_entry field
		bson.D{{Key: "$unset", Value: "user_entry"}},
	}

	var leaderboards []model.Leaderboard
	err := a.db.leaderboards.Aggregate(a.context, pipeline, &leaderboards, nil)

	if len(leaderboards) == 0 {
		return nil, errors.ErrorData(logutils.StatusMissing, model.TypeLeaderboard, &logutils.FieldArgs{"leaderboard_id": leaderboardID})
	}

	return &leaderboards[0], err
}

// GetLeaderboards gets all leaderboards for a user
func (a *Adapter) GetLeaderboards(orgID string, appID string, userID string) ([]model.Leaderboard, error) {
	leaderboardEntryFilter := bson.M{
		"org_id":  orgID,
		"app_id":  appID,
		"user_id": userID,
	}

	pipeline := mongo.Pipeline{
		// 1) filter leadboard entries for this user
		bson.D{{Key: "$match", Value: leaderboardEntryFilter}},
		// 2) sort by date_created in descending order (most recent first)
		bson.D{{Key: "$sort", Value: bson.D{{Key: "date_created", Value: -1}}}},
		// 3) join each entry to its leaderboard
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "leaderboards"},
			{Key: "localField", Value: "leaderboard_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "leaderboard"},
		}}},
		// 4) unwind the resulting array so we get a single doc per leaderboard
		bson.D{{Key: "$unwind", Value: "$leaderboard"}},
		// 5) replace the root with a merged object:
		//   - all fields from the leaderboard
		//   - plus an "is_admin" field from the entry
		bson.D{{Key: "$replaceRoot", Value: bson.D{
			{Key: "newRoot", Value: bson.D{
				{Key: "$mergeObjects", Value: bson.A{
					"$leaderboard",
					bson.D{{Key: "is_admin", Value: "$is_admin"}},
				}},
			}},
		}}},
	}

	var leaderboards []model.Leaderboard
	err := a.db.leaderboardEntries.Aggregate(a.context, pipeline, &leaderboards, nil)

	return leaderboards, err
}

// CreateLeaderboard creates a new leaderboard
func (a *Adapter) CreateLeaderboard(leaderboard model.Leaderboard) (*model.Leaderboard, error) {
	// Create the leaderboard
	_, err := a.db.leaderboards.InsertOne(a.context, leaderboard)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionCreate, model.TypeLeaderboard, nil, err)
	}
	return &leaderboard, nil
}

// UpdateLeaderboard updates an existing leaderboard
func (a *Adapter) UpdateLeaderboard(leaderboard model.Leaderboard) error {
	filter := bson.M{
		"_id":    leaderboard.ID,
		"org_id": leaderboard.OrgID,
		"app_id": leaderboard.AppID,
	}

	setUpdate := bson.M{
		"name":         leaderboard.Name,
		"date_updated": leaderboard.DateUpdated,
	}
	if leaderboard.LastQuizTime != nil {
		setUpdate["last_quiz_time"] = leaderboard.LastQuizTime
	}
	update := bson.M{
		"$set": setUpdate,
	}

	_, err := a.db.leaderboards.UpdateOne(a.context, filter, update, nil)
	return err
}

// DeleteLeaderboard deletes a leaderboard by ID along with corresponding leaderboard entries
func (a *Adapter) DeleteLeaderboard(leaderboardID, orgID, appID, userID string) error {
	filter := bson.M{
		"_id":    leaderboardID,
		"org_id": orgID,
		"app_id": appID,
	}
	_, err := a.db.leaderboards.DeleteOne(a.context, filter, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionDelete, model.TypeLeaderboard, filterArgs(filter), err)
	}

	return err
}

// GetLeaderboardEntries retrieves a list of leaderboard entry
func (a *Adapter) GetLeaderboardEntries(leaderboardID string, orgID string, appID string, userID *string) ([]model.LeaderboardEntry, error) {
	filter := bson.M{
		"leaderboard_id": leaderboardID,
		"org_id":         orgID,
		"app_id":         appID,
	}

	if userID != nil && *userID != "" {
		filter["user_id"] = *userID
	}

	var results []model.LeaderboardEntry
	err := a.db.leaderboardEntries.Find(a.context, filter, &results, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeSurvey, filterArgs(filter), err)
	}

	return results, nil
}

// CreateLeaderboardEntry creates a new leaderboardEntry object
func (a *Adapter) CreateLeaderboardEntry(leaderboardEntry model.LeaderboardEntry) error {
	// Acquire read lock to allow concurrent operations but block during rank initialization
	a.ranksLock.RLock()
	defer a.ranksLock.RUnlock()

	_, err := a.db.leaderboardEntries.InsertOne(a.context, leaderboardEntry)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionCreate, model.TypeLeaderboardEntry, nil, err)
	}
	return nil
}

// DeleteLeaderboardEntries deletes non-admin leaderboard entries according to users in leavingUserIDs
func (a *Adapter) DeleteLeaderboardEntries(leaderboardID string, orgID string, appID string, leavingUserIDs []string) error {
	filter := bson.D{
		primitive.E{Key: "leaderboard_id", Value: leaderboardID},
		primitive.E{Key: "org_id", Value: orgID},
		primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "user_id", Value: bson.M{"$in": leavingUserIDs}},
		primitive.E{Key: "is_admin", Value: false}, // Only delete non-admin entries
	}

	result, err := a.db.leaderboardEntries.DeleteMany(a.context, filter, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionDelete, model.TypeLeaderboardEntry, &logutils.FieldArgs{"user_ids": leavingUserIDs}, err)
	}
	if result.DeletedCount == 0 {
		return errors.ErrorData(logutils.StatusMissing, model.TypeLeaderboardEntry, &logutils.FieldArgs{"user_ids": leavingUserIDs})
	}

	return nil
}

// DeleteAllLeaderboardEntries deletes all leaderboard entries for a given leaderboardID
func (a *Adapter) DeleteAllLeaderboardEntries(leaderboardID string, orgID string, appID string) error {
	filter := bson.M{
		"leaderboard_id": leaderboardID,
		"org_id":         orgID,
		"app_id":         appID,
	}
	_, err := a.db.leaderboardEntries.DeleteMany(a.context, filter, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionDelete, model.TypeLeaderboardEntry, filterArgs(filter), err)
	}

	return err
}

// GetLeaderboardScores retrieves paginated scores for a specific leaderboard
func (a *Adapter) GetLeaderboardScores(leaderboardID string, orgID string, appID string, limit *int, offset *int) ([]model.Score, error) {
	pipeline := mongo.Pipeline{}

	// 1) Match only this leaderboard
	leaderboardEntryFilter := bson.D{{Key: "$match", Value: bson.M{
		"leaderboard_id": leaderboardID,
		"org_id":         orgID,
		"app_id":         appID,
	}}}
	pipeline = append(pipeline, leaderboardEntryFilter)

	// 2) Sort by score descending to get proper ranking order
	sortByScore := bson.D{{Key: "$sort", Value: bson.M{"score": -1}}}
	pipeline = append(pipeline, sortByScore)

	// 3) Apply pagination
	if offset != nil {
		pipeline = append(pipeline, bson.D{{Key: "$skip", Value: *offset}})
	}

	if limit != nil {
		pipeline = append(pipeline, bson.D{{Key: "$limit", Value: *limit}})
	}

	// 4) Lookup to count higher scores for ranking (compatible with MongoDB 4.4+)
	// This creates a rank_data field with count of distinct higher scores
	rankLookup := bson.D{{Key: "$lookup", Value: bson.M{
		"from": "leaderboard_entries",
		"let":  bson.M{"currentScore": "$score", "lbID": "$leaderboard_id"},
		"pipeline": mongo.Pipeline{
			// Match entries in same leaderboard with higher scores
			bson.D{{Key: "$match", Value: bson.M{
				"$expr": bson.M{"$and": bson.A{ // TODO: Using $expr prevents the use of PROJECTION_COVERED/DISTINCT_SCAN
					bson.M{"$eq": bson.A{"$leaderboard_id", "$$lbID"}},
					bson.M{"$eq": bson.A{"$org_id", orgID}},
					bson.M{"$eq": bson.A{"$app_id", appID}},
					bson.M{"$gt": bson.A{"$score", "$$currentScore"}},
				}},
			}}},
			// Group by distinct scores to count them
			bson.D{{Key: "$group", Value: bson.M{
				"_id": "$score",
			}}},
			// Count the distinct higher scores
			bson.D{{Key: "$count", Value: "distinct_higher_scores"}},
		},
		"as": "rank_data",
	}}}
	pipeline = append(pipeline, rankLookup)

	// 5) Add rank field based on count of higher scores
	addRank := bson.D{{Key: "$addFields", Value: bson.M{
		"rank": bson.M{
			"$add": bson.A{
				1,
				bson.M{"$ifNull": bson.A{
					bson.M{"$arrayElemAt": bson.A{"$rank_data.distinct_higher_scores", 0}},
					0,
				}},
			},
		},
	}}}
	pipeline = append(pipeline, addRank)

	// 6) Remove the temporary rank_data field
	unsetRankData := bson.D{{Key: "$unset", Value: "rank_data"}}
	pipeline = append(pipeline, unsetRankData)

	// 7) Lookup into scores collection by user_id + org/app
	lookup := bson.D{{Key: "$lookup", Value: bson.M{
		"from": "scores",
		"let":  bson.M{"uid": "$user_id", "entryRank": "$rank"},
		"pipeline": mongo.Pipeline{
			bson.D{{Key: "$match", Value: bson.M{
				"$expr": bson.M{"$and": bson.A{
					bson.M{"$eq": bson.A{"$user_id", "$$uid"}},
					bson.M{"$eq": bson.A{"$org_id", orgID}},
					bson.M{"$eq": bson.A{"$app_id", appID}},
				}},
			}}},
			// Add the rank from the leaderboard entry to the score document
			bson.D{{Key: "$addFields", Value: bson.M{
				"rank": "$$entryRank",
			}}},
		},
		"as": "scores",
	}}}
	pipeline = append(pipeline, lookup)

	// 8) Unwind the scores array (should be single score per user)
	unwind := bson.D{{Key: "$unwind", Value: "$scores"}}
	pipeline = append(pipeline, unwind)

	// 9) Replace root with the score document (which now includes rank)
	replaceRoot := bson.D{{Key: "$replaceRoot", Value: bson.M{
		"newRoot": "$scores",
	}}}
	pipeline = append(pipeline, replaceRoot)

	var scores []model.Score
	err := a.db.leaderboardEntries.Aggregate(a.context, pipeline, &scores, nil)

	return scores, err
}

// GetLeaderboardUserRanks retrieves leaderboards for a specific user in each leaderboard they're in
func (a *Adapter) GetLeaderboardUserRanks(orgID string, appID string, userID string, limit *int, offset *int) ([]model.Leaderboard, error) {
	pipeline := mongo.Pipeline{}

	leaderboardEntryFilter := bson.D{{Key: "$match", Value: bson.M{
		"org_id":  orgID,
		"app_id":  appID,
		"user_id": userID,
	}}}
	pipeline = append(pipeline, leaderboardEntryFilter)

	sortByDateCreated := bson.D{{Key: "$sort", Value: bson.M{"date_created": -1}}}
	pipeline = append(pipeline, sortByDateCreated)

	if offset != nil {
		pipeline = append(pipeline, bson.D{{Key: "$skip", Value: *offset}})
	}

	if limit != nil {
		pipeline = append(pipeline, bson.D{{Key: "$limit", Value: *limit}})
	}

	lookup := bson.D{{Key: "$lookup", Value: bson.M{
		"from": "leaderboard_entries",
		"let":  bson.M{"userScore": "$score", "lbID": "$leaderboard_id"},
		"pipeline": bson.A{
			// match the same leaderboard/org/app
			bson.D{{Key: "$match", Value: bson.M{
				"$expr": bson.M{"$and": bson.A{ // TODO: Using $expr prevents the use of PROJECTION_COVERED/DISTINCT_SCAN
					bson.M{"$eq": bson.A{"$org_id", orgID}},
					bson.M{"$eq": bson.A{"$app_id", appID}},
					bson.M{"$eq": bson.A{"$leaderboard_id", "$$lbID"}},
					bson.M{"$gt": bson.A{"$score", "$$userScore"}},
				}},
			}}},
			bson.D{{Key: "$group", Value: bson.M{"_id": "$score"}}},
			bson.D{{Key: "$count", Value: "distinct_higher_scores"}},
		},
		"as": "rank_data",
	}}}
	pipeline = append(pipeline, lookup)

	addFields := bson.D{
		{Key: "$addFields", Value: bson.D{
			{Key: "rank", Value: bson.D{
				{Key: "$add", Value: bson.A{
					1,
					bson.D{{Key: "$ifNull", Value: bson.A{
						bson.D{{Key: "$arrayElemAt", Value: bson.A{"$rank_data.distinct_higher_scores", 0}}},
						0,
					}}},
				}},
			}},
		}},
	}
	pipeline = append(pipeline, addFields)

	unset := bson.D{{Key: "$unset", Value: "rank_data"}}
	pipeline = append(pipeline, unset)

	// lookup leaderboards for each entry
	addLeaderboardField := bson.D{{
		Key: "$lookup", Value: bson.M{
			"from":         "leaderboards",
			"localField":   "leaderboard_id",
			"foreignField": "_id",
			"as":           "leaderboard",
		},
	}}
	pipeline = append(pipeline, addLeaderboardField)

	unwindLeaderboard := bson.D{{
		Key: "$unwind", Value: "$leaderboard",
	}}
	pipeline = append(pipeline, unwindLeaderboard)

	// Add the calculated rank directly to the leaderboard document
	addRankToLeaderboard := bson.D{{
		Key: "$addFields", Value: bson.M{
			"leaderboard.rank": "$rank",
		},
	}}
	pipeline = append(pipeline, addRankToLeaderboard)

	// Replace root with the leaderboard document (which now includes the rank)
	replaceWithLeaderboard := bson.D{{
		Key: "$replaceRoot", Value: bson.M{
			"newRoot": "$leaderboard",
		},
	}}
	pipeline = append(pipeline, replaceWithLeaderboard)

	var leaderboards []model.Leaderboard
	err := a.db.leaderboardEntries.Aggregate(a.context, pipeline, &leaderboards, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, "leaderboard user ranks", &logutils.FieldArgs{"org_id": orgID, "app_id": appID, "user_id": userID}, err)
	}

	return leaderboards, err
}

// UpdateLeaderboardEntryScore updates the score field for a specific user's leaderboard entries
func (a *Adapter) UpdateLeaderboardEntryScore(orgID string, appID string, userID string, newScore float64) error {
	// Acquire read lock to allow concurrent operations but block during rank initialization
	a.ranksLock.RLock()
	defer a.ranksLock.RUnlock()

	filter := bson.M{
		"org_id":  orgID,
		"app_id":  appID,
		"user_id": userID,
	}

	update := bson.M{
		"$set": bson.M{
			"score":        newScore,
			"date_updated": time.Now().UTC(),
		},
	}

	_, err := a.db.leaderboardEntries.UpdateMany(a.context, filter, update, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeLeaderboardEntry, &logutils.FieldArgs{"org_id": orgID, "app_id": appID, "user_id": userID}, err)
	}

	return nil
}

// InitLeaderboardEntryScores finds and updates score fields for all leaderboard entries in the given org/app
func (a *Adapter) InitLeaderboardEntryScores(orgID string, appID string) error {

	pipeline := mongo.Pipeline{
		// Match leaderboard entries for the specific org/app
		bson.D{{Key: "$match", Value: bson.M{
			"org_id": orgID,
			"app_id": appID,
		}}},
		// Lookup scores for each user
		bson.D{{Key: "$lookup", Value: bson.M{
			"from": "scores",
			"let":  bson.M{"uid": "$user_id"},
			"pipeline": mongo.Pipeline{
				// match the same user/org/app
				{{Key: "$match", Value: bson.M{
					"$expr": bson.M{"$and": bson.A{
						bson.M{"$eq": bson.A{"$user_id", "$$uid"}},
						bson.M{"$eq": bson.A{"$org_id", orgID}},
						bson.M{"$eq": bson.A{"$app_id", appID}},
					}},
				}}},
			},
			"as": "user_scores",
		}}},
		// Unwind the scores array (should be single score per user)
		bson.D{{Key: "$unwind", Value: bson.M{
			"path":                       "$user_scores",
			"preserveNullAndEmptyArrays": true,
		}}},
		// Add score field from the looked up score document
		bson.D{{Key: "$addFields", Value: bson.M{
			"score": bson.M{
				"$ifNull": bson.A{"$user_scores.score", 0},
			},
		}}},
		// Remove the temporary user_scores field and merge back to leaderboard_entries
		bson.D{{Key: "$unset", Value: "user_scores"}},
		bson.D{{Key: "$merge", Value: bson.M{
			"into":           "leaderboard_entries",
			"whenMatched":    "merge",
			"whenNotMatched": "discard",
		}}},
	}

	// Execute the aggregation pipeline that updates the documents directly
	var results []bson.M // We don't need the results since $merge updates in place
	err := a.db.leaderboardEntries.Aggregate(a.context, pipeline, &results, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeLeaderboardEntry, &logutils.FieldArgs{"app_id": appID, "org_id": orgID}, err)
	}

	return nil
}
